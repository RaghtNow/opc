package classroom

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/RaghtNow/opc/apps/edu-insight-api/internal/domain/classroom"
)

const CurrentClassID = "admin-class-g2-3"

type Service interface {
	GetWorkspace(classID string) (classroom.Workspace, error)
	CreateClass(req classroom.CreateClassRequest) (classroom.Workspace, error)
	CreateStudent(classID string, req classroom.SaveStudentRequest) (classroom.Workspace, error)
	UpdateStudent(classID, id string, req classroom.SaveStudentRequest) (classroom.Workspace, bool, error)
	DeleteStudent(classID, id string) (classroom.Workspace, bool, error)
	ImportStudents(classID, fileName string, content []byte) (classroom.Workspace, classroom.ImportSummary, error)
	CreateTeacher(classID string, req classroom.SaveTeacherRequest) (classroom.Workspace, error)
	UpdateTeacher(classID, id string, req classroom.SaveTeacherRequest) (classroom.Workspace, bool, error)
	DeleteTeacher(classID, id string) (classroom.Workspace, bool, error)
	ImportTeachers(classID, fileName string, content []byte) (classroom.Workspace, classroom.ImportSummary, error)
	BindTeacherAccount(classID, id string) (classroom.Workspace, bool, error)
	SyncTeacherPermission(classID, id string) (classroom.Workspace, bool, error)
	UpdatePolicy(classID, id string, req classroom.SavePolicyRequest) (classroom.Workspace, bool, error)
}

type Repository interface {
	GetWorkspace(classID string) (classroom.Workspace, error)
	ListClassOptions() ([]classroom.ClassOption, error)
	SaveClass(workspace classroom.Workspace) error
	SaveStudent(classID string, student classroom.Student) error
	ArchiveStudent(classID, studentID string) error
	SaveTeacher(classID string, teacher classroom.TeacherAssignment) error
	DeleteTeacher(classID, teacherID string) error
	SavePolicy(classID string, policy classroom.Policy) error
	SeedIfEmpty(workspace classroom.Workspace) error
}

type PersistentService struct {
	repo    Repository
	classID string
}

func NewPersistentService(repo Repository) (*PersistentService, error) {
	service := &PersistentService{repo: repo, classID: CurrentClassID}
	for _, workspace := range defaultWorkspaces() {
		if err := repo.SeedIfEmpty(workspace); err != nil {
			return nil, err
		}
	}
	return service, nil
}

func (s *PersistentService) GetWorkspace(classID string) (classroom.Workspace, error) {
	return s.workspace(classID)
}

func (s *PersistentService) CreateClass(req classroom.CreateClassRequest) (classroom.Workspace, error) {
	workspace := classWorkspaceFromRequest(req)
	if err := s.repo.SaveClass(workspace); err != nil {
		return classroom.Workspace{}, err
	}
	return s.workspace(workspace.ClassID)
}

func (s *PersistentService) workspace(classID string) (classroom.Workspace, error) {
	workspace, err := s.repo.GetWorkspace(defaultClassID(classID))
	if err != nil {
		return classroom.Workspace{}, err
	}
	options, err := s.repo.ListClassOptions()
	if err != nil {
		return classroom.Workspace{}, err
	}
	workspace.ClassOptions = options
	return workspace, nil
}

func (s *PersistentService) CreateStudent(classID string, req classroom.SaveStudentRequest) (classroom.Workspace, error) {
	student := studentFromRequest(fmt.Sprintf("student-%d", time.Now().UnixNano()), req)
	if student.StudentNo == "" {
		workspace, err := s.repo.GetWorkspace(defaultClassID(classID))
		if err != nil {
			return classroom.Workspace{}, err
		}
		student.StudentNo = fmt.Sprintf("NEW%03d", len(workspace.Students)+1)
	}
	if student.Name == "" {
		student.Name = "新学生"
	}
	if err := s.repo.SaveStudent(defaultClassID(classID), student); err != nil {
		return classroom.Workspace{}, err
	}
	return s.workspace(classID)
}

func (s *PersistentService) UpdateStudent(classID, id string, req classroom.SaveStudentRequest) (classroom.Workspace, bool, error) {
	workspace, err := s.repo.GetWorkspace(defaultClassID(classID))
	if err != nil {
		return classroom.Workspace{}, false, err
	}
	if !studentExists(workspace.Students, id) {
		return classroom.Workspace{}, false, nil
	}
	if err := s.repo.SaveStudent(defaultClassID(classID), studentFromRequest(id, req)); err != nil {
		return classroom.Workspace{}, false, err
	}
	workspace, err = s.workspace(classID)
	return workspace, true, err
}

func (s *PersistentService) DeleteStudent(classID, id string) (classroom.Workspace, bool, error) {
	workspace, err := s.repo.GetWorkspace(defaultClassID(classID))
	if err != nil {
		return classroom.Workspace{}, false, err
	}
	if !studentExists(workspace.Students, id) {
		return classroom.Workspace{}, false, nil
	}
	if err := s.repo.ArchiveStudent(defaultClassID(classID), id); err != nil {
		return classroom.Workspace{}, false, err
	}
	workspace, err = s.workspace(classID)
	return workspace, true, err
}

func (s *PersistentService) ImportStudents(classID, fileName string, content []byte) (classroom.Workspace, classroom.ImportSummary, error) {
	rows, err := parseCSVRows(fileName, content)
	if err != nil {
		return classroom.Workspace{}, classroom.ImportSummary{}, err
	}
	workspace, err := s.repo.GetWorkspace(defaultClassID(classID))
	if err != nil {
		return classroom.Workspace{}, classroom.ImportSummary{}, err
	}
	byNo := map[string]classroom.Student{}
	for _, student := range workspace.Students {
		byNo[student.StudentNo] = student
	}

	summary := classroom.ImportSummary{Errors: []string{}}
	for index, row := range rows {
		studentNo := row["学号"]
		name := row["姓名"]
		if studentNo == "" || name == "" {
			summary.Skipped++
			summary.Errors = append(summary.Errors, fmt.Sprintf("第 %d 行缺少学号或姓名", index+2))
			continue
		}
		id := fmt.Sprintf("student-%d-%d", time.Now().UnixNano(), index)
		if existing, ok := byNo[studentNo]; ok {
			id = existing.ID
			summary.Updated++
		} else {
			summary.Created++
		}
		parentStatus := "待补充"
		if row["家长手机号"] != "" {
			parentStatus = "已绑定"
		}
		selectionStatus := "待确认"
		if row["选科组合"] != "" {
			selectionStatus = "已登记"
		}
		student := studentFromRequest(id, classroom.SaveStudentRequest{
			StudentNo:       studentNo,
			Name:            name,
			Gender:          defaultString(row["性别"], "男"),
			Combination:     row["选科组合"],
			ParentMobile:    row["家长手机号"],
			ParentStatus:    parentStatus,
			SelectionStatus: selectionStatus,
		})
		if err := s.repo.SaveStudent(defaultClassID(classID), student); err != nil {
			return classroom.Workspace{}, classroom.ImportSummary{}, err
		}
	}
	workspace, err = s.workspace(classID)
	return workspace, summary, err
}

func (s *PersistentService) CreateTeacher(classID string, req classroom.SaveTeacherRequest) (classroom.Workspace, error) {
	teacher := teacherFromRequest(fmt.Sprintf("teacher-%d", time.Now().UnixNano()), req)
	if teacher.Subject == "" {
		teacher.Subject = "新学科"
	}
	if teacher.Teacher == "" {
		teacher.Teacher = "新老师"
	}
	if teacher.Classes == "" {
		teacher.Classes = "待设置范围"
	}
	if err := s.repo.SaveTeacher(defaultClassID(classID), teacher); err != nil {
		return classroom.Workspace{}, err
	}
	return s.workspace(classID)
}

func (s *PersistentService) UpdateTeacher(classID, id string, req classroom.SaveTeacherRequest) (classroom.Workspace, bool, error) {
	workspace, err := s.repo.GetWorkspace(defaultClassID(classID))
	if err != nil {
		return classroom.Workspace{}, false, err
	}
	for _, teacher := range workspace.Teachers {
		if teacher.ID == id {
			teacher.Subject = req.Subject
			teacher.Teacher = req.Teacher
			teacher.Mobile = req.Mobile
			teacher.Classes = req.Classes
			if err := s.repo.SaveTeacher(defaultClassID(classID), normalizeTeacherState(teacher)); err != nil {
				return classroom.Workspace{}, false, err
			}
			workspace, err = s.workspace(classID)
			return workspace, true, err
		}
	}
	return classroom.Workspace{}, false, nil
}

func (s *PersistentService) DeleteTeacher(classID, id string) (classroom.Workspace, bool, error) {
	workspace, err := s.repo.GetWorkspace(defaultClassID(classID))
	if err != nil {
		return classroom.Workspace{}, false, err
	}
	if !teacherExists(workspace.Teachers, id) {
		return classroom.Workspace{}, false, nil
	}
	if err := s.repo.DeleteTeacher(defaultClassID(classID), id); err != nil {
		return classroom.Workspace{}, false, err
	}
	workspace, err = s.workspace(classID)
	return workspace, true, err
}

func (s *PersistentService) ImportTeachers(classID, fileName string, content []byte) (classroom.Workspace, classroom.ImportSummary, error) {
	rows, err := parseCSVRows(fileName, content)
	if err != nil {
		return classroom.Workspace{}, classroom.ImportSummary{}, err
	}
	workspace, err := s.repo.GetWorkspace(defaultClassID(classID))
	if err != nil {
		return classroom.Workspace{}, classroom.ImportSummary{}, err
	}
	bySubjectTeacher := map[string]classroom.TeacherAssignment{}
	for _, teacher := range workspace.Teachers {
		bySubjectTeacher[teacher.Subject+"|"+teacher.Teacher] = teacher
	}

	summary := classroom.ImportSummary{Errors: []string{}}
	for index, row := range rows {
		subject := row["学科"]
		teacherName := firstNonEmpty(row["老师姓名"], row["任课老师"], row["教师"])
		if subject == "" || teacherName == "" {
			summary.Skipped++
			summary.Errors = append(summary.Errors, fmt.Sprintf("第 %d 行缺少学科或老师姓名", index+2))
			continue
		}
		key := subject + "|" + teacherName
		id := fmt.Sprintf("teacher-%d-%d", time.Now().UnixNano(), index)
		if existing, ok := bySubjectTeacher[key]; ok {
			id = existing.ID
			summary.Updated++
		} else {
			summary.Created++
		}
		teacher := teacherFromRequest(id, classroom.SaveTeacherRequest{
			Subject:          subject,
			Teacher:          teacherName,
			Mobile:           firstNonEmpty(row["手机号"], row["账号手机号"], row["教师手机号"]),
			Classes:          defaultString(row["授课范围"], "待设置范围"),
			AccountStatus:    "pending",
			PermissionStatus: "pending",
		})
		if existing, ok := bySubjectTeacher[key]; ok {
			teacher.AccountStatus = existing.AccountStatus
			teacher.AccountID = existing.AccountID
			teacher.AccountBoundAt = existing.AccountBoundAt
			teacher.PermissionStatus = existing.PermissionStatus
			teacher.PermissionSyncedAt = existing.PermissionSyncedAt
		}
		if err := s.repo.SaveTeacher(defaultClassID(classID), teacher); err != nil {
			return classroom.Workspace{}, classroom.ImportSummary{}, err
		}
	}
	workspace, err = s.workspace(classID)
	return workspace, summary, err
}

func (s *PersistentService) BindTeacherAccount(classID, id string) (classroom.Workspace, bool, error) {
	return s.updateTeacherState(classID, id, func(teacher classroom.TeacherAssignment) (classroom.TeacherAssignment, error) {
		if strings.TrimSpace(teacher.Mobile) == "" {
			return teacher, fmt.Errorf("请先维护教师手机号，再绑定账号")
		}
		teacher.AccountStatus = "bound"
		if teacher.AccountID == "" {
			teacher.AccountID = accountIDFromTeacherID(teacher.ID)
		}
		teacher.AccountBoundAt = time.Now().Format("2006-01-02 15:04")
		return teacher, nil
	})
}

func (s *PersistentService) SyncTeacherPermission(classID, id string) (classroom.Workspace, bool, error) {
	return s.updateTeacherState(classID, id, func(teacher classroom.TeacherAssignment) (classroom.TeacherAssignment, error) {
		if teacher.AccountStatus != "bound" {
			return teacher, fmt.Errorf("请先绑定教师账号，再同步权限")
		}
		teacher.PermissionStatus = "synced"
		teacher.PermissionSyncedAt = time.Now().Format("2006-01-02 15:04")
		return teacher, nil
	})
}

func (s *PersistentService) updateTeacherState(classID, id string, mutate func(classroom.TeacherAssignment) (classroom.TeacherAssignment, error)) (classroom.Workspace, bool, error) {
	workspace, err := s.repo.GetWorkspace(defaultClassID(classID))
	if err != nil {
		return classroom.Workspace{}, false, err
	}
	for _, teacher := range workspace.Teachers {
		if teacher.ID == id {
			teacher, err = mutate(teacher)
			if err != nil {
				return classroom.Workspace{}, false, err
			}
			teacher = normalizeTeacherState(teacher)
			if err := s.repo.SaveTeacher(defaultClassID(classID), teacher); err != nil {
				return classroom.Workspace{}, false, err
			}
			workspace, err = s.workspace(classID)
			return workspace, true, err
		}
	}
	return classroom.Workspace{}, false, nil
}

func (s *PersistentService) UpdatePolicy(classID, id string, req classroom.SavePolicyRequest) (classroom.Workspace, bool, error) {
	workspace, err := s.repo.GetWorkspace(defaultClassID(classID))
	if err != nil {
		return classroom.Workspace{}, false, err
	}
	for _, policy := range workspace.Policies {
		if policy.ID == id {
			policy.Value = req.Value
			policy.Note = req.Note
			if err := s.repo.SavePolicy(defaultClassID(classID), policy); err != nil {
				return classroom.Workspace{}, false, err
			}
			workspace, err = s.workspace(classID)
			return workspace, true, err
		}
	}
	return classroom.Workspace{}, false, nil
}

type MemoryService struct {
	mu         sync.RWMutex
	workspaces map[string]classroom.Workspace
}

func NewMemoryService() *MemoryService {
	workspaces := map[string]classroom.Workspace{}
	for _, workspace := range defaultWorkspaces() {
		workspaces[workspace.ClassID] = withRosterInsights(workspace)
	}
	return &MemoryService{workspaces: workspaces}
}

func (s *MemoryService) GetWorkspace(classID string) (classroom.Workspace, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.snapshotLocked(defaultClassID(classID)), nil
}

func (s *MemoryService) CreateClass(req classroom.CreateClassRequest) (classroom.Workspace, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	workspace := classWorkspaceFromRequest(req)
	s.workspaces[workspace.ClassID] = withRosterInsights(workspace)
	return s.snapshotLocked(workspace.ClassID), nil
}

func (s *MemoryService) CreateStudent(classID string, req classroom.SaveStudentRequest) (classroom.Workspace, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := defaultClassID(classID)
	workspace := s.workspaces[id]
	student := studentFromRequest(fmt.Sprintf("student-%d", time.Now().UnixNano()), req)
	if student.StudentNo == "" {
		student.StudentNo = fmt.Sprintf("NEW%03d", len(workspace.Students)+1)
	}
	if student.Name == "" {
		student.Name = "新学生"
	}
	workspace.Students = append([]classroom.Student{student}, workspace.Students...)
	s.workspaces[id] = withRosterInsights(workspace)
	return s.snapshotLocked(id), nil
}

func (s *MemoryService) UpdateStudent(classID, id string, req classroom.SaveStudentRequest) (classroom.Workspace, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	classKey := defaultClassID(classID)
	workspace := s.workspaces[classKey]
	for i := range workspace.Students {
		if workspace.Students[i].ID == id {
			workspace.Students[i] = studentFromRequest(id, req)
			s.workspaces[classKey] = withRosterInsights(workspace)
			return s.snapshotLocked(classKey), true, nil
		}
	}
	return classroom.Workspace{}, false, nil
}

func (s *MemoryService) DeleteStudent(classID, id string) (classroom.Workspace, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	classKey := defaultClassID(classID)
	workspace := s.workspaces[classKey]
	for i := range workspace.Students {
		if workspace.Students[i].ID == id {
			workspace.Students = append(workspace.Students[:i], workspace.Students[i+1:]...)
			s.workspaces[classKey] = withRosterInsights(workspace)
			return s.snapshotLocked(classKey), true, nil
		}
	}
	return classroom.Workspace{}, false, nil
}

func (s *MemoryService) ImportStudents(classID, fileName string, content []byte) (classroom.Workspace, classroom.ImportSummary, error) {
	rows, err := parseCSVRows(fileName, content)
	if err != nil {
		return classroom.Workspace{}, classroom.ImportSummary{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	classKey := defaultClassID(classID)
	workspace := s.workspaces[classKey]
	byNo := map[string]int{}
	for index, student := range workspace.Students {
		byNo[student.StudentNo] = index
	}
	summary := classroom.ImportSummary{Errors: []string{}}
	for index, row := range rows {
		if row["学号"] == "" || row["姓名"] == "" {
			summary.Skipped++
			summary.Errors = append(summary.Errors, fmt.Sprintf("第 %d 行缺少学号或姓名", index+2))
			continue
		}
		parentStatus := "待补充"
		if row["家长手机号"] != "" {
			parentStatus = "已绑定"
		}
		selectionStatus := "待确认"
		if row["选科组合"] != "" {
			selectionStatus = "已登记"
		}
		student := studentFromRequest(fmt.Sprintf("student-%d-%d", time.Now().UnixNano(), index), classroom.SaveStudentRequest{
			StudentNo:       row["学号"],
			Name:            row["姓名"],
			Gender:          defaultString(row["性别"], "男"),
			Combination:     row["选科组合"],
			ParentMobile:    row["家长手机号"],
			ParentStatus:    parentStatus,
			SelectionStatus: selectionStatus,
		})
		if existingIndex, ok := byNo[student.StudentNo]; ok {
			student.ID = workspace.Students[existingIndex].ID
			workspace.Students[existingIndex] = student
			summary.Updated++
		} else {
			workspace.Students = append([]classroom.Student{student}, workspace.Students...)
			summary.Created++
		}
	}
	s.workspaces[classKey] = withRosterInsights(workspace)
	return s.snapshotLocked(classKey), summary, nil
}

func (s *MemoryService) CreateTeacher(classID string, req classroom.SaveTeacherRequest) (classroom.Workspace, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	classKey := defaultClassID(classID)
	workspace := s.workspaces[classKey]
	teacher := teacherFromRequest(fmt.Sprintf("teacher-%d", time.Now().UnixNano()), req)
	if teacher.Subject == "" {
		teacher.Subject = "新学科"
	}
	if teacher.Teacher == "" {
		teacher.Teacher = "新老师"
	}
	if teacher.Classes == "" {
		teacher.Classes = "待设置范围"
	}
	workspace.Teachers = append([]classroom.TeacherAssignment{teacher}, workspace.Teachers...)
	s.workspaces[classKey] = withRosterInsights(workspace)
	return s.snapshotLocked(classKey), nil
}

func (s *MemoryService) UpdateTeacher(classID, id string, req classroom.SaveTeacherRequest) (classroom.Workspace, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	classKey := defaultClassID(classID)
	workspace := s.workspaces[classKey]
	for i := range workspace.Teachers {
		if workspace.Teachers[i].ID == id {
			workspace.Teachers[i].Subject = req.Subject
			workspace.Teachers[i].Teacher = req.Teacher
			workspace.Teachers[i].Mobile = req.Mobile
			workspace.Teachers[i].Classes = req.Classes
			workspace.Teachers[i] = normalizeTeacherState(workspace.Teachers[i])
			s.workspaces[classKey] = withRosterInsights(workspace)
			return s.snapshotLocked(classKey), true, nil
		}
	}
	return classroom.Workspace{}, false, nil
}

func (s *MemoryService) DeleteTeacher(classID, id string) (classroom.Workspace, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	classKey := defaultClassID(classID)
	workspace := s.workspaces[classKey]
	for i := range workspace.Teachers {
		if workspace.Teachers[i].ID == id {
			workspace.Teachers = append(workspace.Teachers[:i], workspace.Teachers[i+1:]...)
			s.workspaces[classKey] = withRosterInsights(workspace)
			return s.snapshotLocked(classKey), true, nil
		}
	}
	return classroom.Workspace{}, false, nil
}

func (s *MemoryService) ImportTeachers(classID, fileName string, content []byte) (classroom.Workspace, classroom.ImportSummary, error) {
	rows, err := parseCSVRows(fileName, content)
	if err != nil {
		return classroom.Workspace{}, classroom.ImportSummary{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	classKey := defaultClassID(classID)
	workspace := s.workspaces[classKey]
	byKey := map[string]int{}
	for index, teacher := range workspace.Teachers {
		byKey[teacher.Subject+"|"+teacher.Teacher] = index
	}
	summary := classroom.ImportSummary{Errors: []string{}}
	for index, row := range rows {
		subject := row["学科"]
		teacherName := firstNonEmpty(row["老师姓名"], row["任课老师"], row["教师"])
		if subject == "" || teacherName == "" {
			summary.Skipped++
			summary.Errors = append(summary.Errors, fmt.Sprintf("第 %d 行缺少学科或老师姓名", index+2))
			continue
		}
		teacher := teacherFromRequest(fmt.Sprintf("teacher-%d-%d", time.Now().UnixNano(), index), classroom.SaveTeacherRequest{
			Subject:          subject,
			Teacher:          teacherName,
			Mobile:           firstNonEmpty(row["手机号"], row["账号手机号"], row["教师手机号"]),
			Classes:          defaultString(row["授课范围"], "待设置范围"),
			AccountStatus:    "pending",
			PermissionStatus: "pending",
		})
		key := subject + "|" + teacherName
		if existingIndex, ok := byKey[key]; ok {
			teacher.ID = workspace.Teachers[existingIndex].ID
			teacher.AccountStatus = workspace.Teachers[existingIndex].AccountStatus
			teacher.AccountID = workspace.Teachers[existingIndex].AccountID
			teacher.AccountBoundAt = workspace.Teachers[existingIndex].AccountBoundAt
			teacher.PermissionStatus = workspace.Teachers[existingIndex].PermissionStatus
			teacher.PermissionSyncedAt = workspace.Teachers[existingIndex].PermissionSyncedAt
			workspace.Teachers[existingIndex] = teacher
			summary.Updated++
		} else {
			workspace.Teachers = append([]classroom.TeacherAssignment{teacher}, workspace.Teachers...)
			summary.Created++
		}
	}
	s.workspaces[classKey] = withRosterInsights(workspace)
	return s.snapshotLocked(classKey), summary, nil
}

func (s *MemoryService) BindTeacherAccount(classID, id string) (classroom.Workspace, bool, error) {
	return s.updateTeacherState(classID, id, func(teacher classroom.TeacherAssignment) (classroom.TeacherAssignment, error) {
		if strings.TrimSpace(teacher.Mobile) == "" {
			return teacher, fmt.Errorf("请先维护教师手机号，再绑定账号")
		}
		teacher.AccountStatus = "bound"
		if teacher.AccountID == "" {
			teacher.AccountID = accountIDFromTeacherID(teacher.ID)
		}
		teacher.AccountBoundAt = time.Now().Format("2006-01-02 15:04")
		return teacher, nil
	})
}

func (s *MemoryService) SyncTeacherPermission(classID, id string) (classroom.Workspace, bool, error) {
	return s.updateTeacherState(classID, id, func(teacher classroom.TeacherAssignment) (classroom.TeacherAssignment, error) {
		if teacher.AccountStatus != "bound" {
			return teacher, fmt.Errorf("请先绑定教师账号，再同步权限")
		}
		teacher.PermissionStatus = "synced"
		teacher.PermissionSyncedAt = time.Now().Format("2006-01-02 15:04")
		return teacher, nil
	})
}

func (s *MemoryService) updateTeacherState(classID, id string, mutate func(classroom.TeacherAssignment) (classroom.TeacherAssignment, error)) (classroom.Workspace, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	classKey := defaultClassID(classID)
	workspace := s.workspaces[classKey]
	for i := range workspace.Teachers {
		if workspace.Teachers[i].ID == id {
			teacher, err := mutate(workspace.Teachers[i])
			if err != nil {
				return classroom.Workspace{}, false, err
			}
			workspace.Teachers[i] = normalizeTeacherState(teacher)
			s.workspaces[classKey] = withRosterInsights(workspace)
			return s.snapshotLocked(classKey), true, nil
		}
	}
	return classroom.Workspace{}, false, nil
}

func (s *MemoryService) UpdatePolicy(classID, id string, req classroom.SavePolicyRequest) (classroom.Workspace, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	classKey := defaultClassID(classID)
	workspace := s.workspaces[classKey]
	for i := range workspace.Policies {
		if workspace.Policies[i].ID == id {
			workspace.Policies[i].Value = req.Value
			workspace.Policies[i].Note = req.Note
			s.workspaces[classKey] = workspace
			return s.snapshotLocked(classKey), true, nil
		}
	}
	return classroom.Workspace{}, false, nil
}

func (s *MemoryService) snapshotLocked(classID string) classroom.Workspace {
	workspace := snapshot(s.workspaces[classID])
	workspace.ClassOptions = classOptionsFromWorkspaces(s.workspaces)
	return workspace
}

func withRosterInsights(workspace classroom.Workspace) classroom.Workspace {
	parentReady := 0
	selectionReady := 0
	for _, student := range workspace.Students {
		if student.ParentStatus == "已绑定" {
			parentReady++
		}
		if student.SelectionStatus == "已登记" {
			selectionReady++
		}
	}

	teacherReady := 0
	for _, teacher := range workspace.Teachers {
		if teacher.AccountStatus == "bound" {
			teacherReady++
		}
	}

	totalStudents := len(workspace.Students)
	totalTeachers := len(workspace.Teachers)
	workspace.RosterInsights = []classroom.RosterInsight{
		{Title: "已完成家长绑定", Count: fmt.Sprintf("%d / %d", parentReady, totalStudents), Detail: fmt.Sprintf("%d 位学生缺少有效家长手机号，影响成绩同步。", totalStudents-parentReady)},
		{Title: "已完成选科登记", Count: fmt.Sprintf("%d / %d", selectionReady, totalStudents), Detail: fmt.Sprintf("%d 位学生选科仍待确认，会影响选考科成绩校验。", totalStudents-selectionReady)},
		{Title: "任课老师已绑定账号", Count: fmt.Sprintf("%d / %d", teacherReady, totalTeachers), Detail: fmt.Sprintf("%d 位任课老师尚未绑定账号，暂无法自动同步单科分析。", totalTeachers-teacherReady)},
	}
	return workspace
}

func snapshot(workspace classroom.Workspace) classroom.Workspace {
	workspace.BaseFields = append([]classroom.BaseField{}, workspace.BaseFields...)
	workspace.RosterInsights = append([]classroom.RosterInsight{}, workspace.RosterInsights...)
	workspace.Students = append([]classroom.Student{}, workspace.Students...)
	workspace.Teachers = append([]classroom.TeacherAssignment{}, workspace.Teachers...)
	workspace.Policies = append([]classroom.Policy{}, workspace.Policies...)
	return workspace
}

func studentExists(students []classroom.Student, id string) bool {
	for _, student := range students {
		if student.ID == id {
			return true
		}
	}
	return false
}

func teacherExists(teachers []classroom.TeacherAssignment, id string) bool {
	for _, teacher := range teachers {
		if teacher.ID == id {
			return true
		}
	}
	return false
}

func studentFromRequest(id string, req classroom.SaveStudentRequest) classroom.Student {
	parentStatus := req.ParentStatus
	if parentStatus == "" {
		parentStatus = "待补充"
	}
	selectionStatus := req.SelectionStatus
	if selectionStatus == "" {
		selectionStatus = "待确认"
	}

	status := "已完整"
	profileStatus := "ready"
	if parentStatus != "已绑定" {
		status = "待补家长"
		profileStatus = "missing_parent"
	} else if selectionStatus != "已登记" {
		status = "选科待确认"
		profileStatus = "missing_selection"
	}

	return classroom.Student{
		ID:               id,
		StudentNo:        req.StudentNo,
		Name:             req.Name,
		Gender:           req.Gender,
		Combination:      req.Combination,
		ElectiveSubjects: splitCombination(req.Combination),
		ParentMobile:     req.ParentMobile,
		Status:           status,
		ParentStatus:     parentStatus,
		SelectionStatus:  selectionStatus,
		ProfileStatus:    profileStatus,
	}
}

func teacherFromRequest(id string, req classroom.SaveTeacherRequest) classroom.TeacherAssignment {
	accountStatus := req.AccountStatus
	if accountStatus == "" {
		accountStatus = "pending"
	}
	permissionStatus := req.PermissionStatus
	if permissionStatus == "" {
		permissionStatus = "pending"
	}

	return normalizeTeacherState(classroom.TeacherAssignment{
		ID:               id,
		Subject:          req.Subject,
		Teacher:          req.Teacher,
		Mobile:           req.Mobile,
		Classes:          req.Classes,
		AccountStatus:    accountStatus,
		PermissionStatus: permissionStatus,
	})
}

func normalizeTeacherState(teacher classroom.TeacherAssignment) classroom.TeacherAssignment {
	if teacher.AccountStatus == "" {
		teacher.AccountStatus = "pending"
	}
	if strings.TrimSpace(teacher.Mobile) == "" {
		teacher.AccountStatus = "pending"
		teacher.AccountID = ""
		teacher.AccountBoundAt = ""
		teacher.PermissionStatus = "pending"
		teacher.PermissionSyncedAt = ""
	}
	if teacher.PermissionStatus == "" {
		teacher.PermissionStatus = "pending"
	}
	teacher.SyncStatus = "待补账号绑定"
	if teacher.AccountStatus == "bound" && teacher.PermissionStatus == "synced" {
		teacher.SyncStatus = "已同步"
	} else if teacher.AccountStatus == "bound" {
		teacher.SyncStatus = "待同步权限"
	}
	return teacher
}

func splitCombination(combination string) []string {
	replacer := strings.NewReplacer("物", "物理,", "化", "化学,", "生", "生物,", "史", "历史,", "地", "地理,", "政", "政治,")
	expanded := strings.TrimSuffix(replacer.Replace(combination), ",")
	if expanded == "" || expanded == combination {
		return []string{}
	}
	return strings.Split(expanded, ",")
}

func parseCSVRows(fileName string, content []byte) ([]map[string]string, error) {
	if !strings.HasSuffix(strings.ToLower(fileName), ".csv") {
		return nil, fmt.Errorf("当前批量导入先支持 CSV 文件")
	}
	reader := csv.NewReader(bytes.NewReader(content))
	reader.TrimLeadingSpace = true
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("CSV 解析失败：%w", err)
	}
	if len(records) < 2 {
		return []map[string]string{}, nil
	}
	headers := make([]string, len(records[0]))
	for index, header := range records[0] {
		headers[index] = strings.TrimSpace(header)
	}
	rows := []map[string]string{}
	for _, record := range records[1:] {
		row := map[string]string{}
		empty := true
		for index, header := range headers {
			value := ""
			if index < len(record) {
				value = strings.TrimSpace(record[index])
			}
			if value != "" {
				empty = false
			}
			row[header] = value
		}
		if !empty {
			rows = append(rows, row)
		}
	}
	return rows, nil
}

func defaultString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func accountIDFromTeacherID(teacherID string) string {
	return "teacher-account-" + teacherID
}

func classWorkspaceFromRequest(req classroom.CreateClassRequest) classroom.Workspace {
	schoolName := defaultString(req.SchoolName, "未设置学校")
	gradeName := defaultString(req.GradeName, "未设置年级")
	className := defaultString(req.ClassName, "新建班级")
	homeroomTeacher := defaultString(req.HomeroomTeacher, "未设置班主任")
	academicYear := defaultString(req.AcademicYear, "2025-2026 学年")
	stage := selectionStage(req.StageID)
	classID := fmt.Sprintf("class-%d-%s", time.Now().UnixNano(), compactID(gradeName+className))

	return classroom.Workspace{
		ClassID:   classID,
		ClassName: className,
		Stage:     stage,
		BaseFields: []classroom.BaseField{
			{Label: "学校", Value: schoolName},
			{Label: "年级", Value: gradeName},
			{Label: "行政班", Value: className},
			{Label: "班主任", Value: homeroomTeacher},
			{Label: "学年", Value: academicYear},
			{Label: "班级状态", Value: "正常使用中"},
		},
		Students: []classroom.Student{},
		Teachers: []classroom.TeacherAssignment{},
		Policies: defaultPolicies(stage.ID),
	}
}

func selectionStage(stageID string) classroom.SelectionStage {
	if stageID == "pre-selection" {
		return classroom.SelectionStage{
			ID:          "pre-selection",
			Label:       "未选科阶段",
			Description: "适用于高一前期或尚未完成选科分流的班级，以行政班和通用科目分析为主。",
		}
	}
	return classroom.SelectionStage{
		ID:          "post-selection",
		Label:       "已选科阶段",
		Description: "适用于完成选科后的班级，支持行政班、教学班、选科组合和赋分分析。",
	}
}

func defaultPolicies(stageID string) []classroom.Policy {
	rankValue := "班主任可配置"
	bandValue := "已开启"
	rankNote := "当前班级允许展示班级位置，不允许展示年级排名。"
	bandNote := "班主任已开启班级分数段和班级均分对比。"
	if stageID == "pre-selection" {
		rankValue = "关闭"
		bandValue = "关闭"
		rankNote = "未选科阶段默认不展示班级位置。"
		bandNote = "等首次考试稳定后再开放分数段展示。"
	}
	return []classroom.Policy{
		{ID: "policy-parent-class-rank", Title: "家长端班级位置", Value: rankValue, Note: rankNote},
		{ID: "policy-student-band", Title: "学生端分数段", Value: bandValue, Note: bandNote},
		{ID: "policy-sync-trigger", Title: "同步策略", Value: "考试发布后触发", Note: "成绩分析完成后，再统一同步家长、学生与任课老师。"},
	}
}

func compactID(value string) string {
	replacer := strings.NewReplacer("（", "-", "）", "", "(", "-", ")", "", " ", "", "　", "", "/", "-", "\\", "-")
	value = strings.Trim(replacer.Replace(value), "-")
	if value == "" {
		return "new"
	}
	return value
}

func defaultClassID(classID string) string {
	if strings.TrimSpace(classID) == "" {
		return CurrentClassID
	}
	return classID
}

func classOptionsFromWorkspaces(workspaces map[string]classroom.Workspace) []classroom.ClassOption {
	options := []classroom.ClassOption{}
	for _, workspace := range workspaces {
		fields := map[string]string{}
		for _, field := range workspace.BaseFields {
			fields[field.Label] = field.Value
		}
		options = append(options, classroom.ClassOption{
			ID:             workspace.ClassID,
			ClassName:      workspace.ClassName,
			GradeName:      fields["年级"],
			AcademicYear:   fields["学年"],
			RoleLabel:      "班主任",
			PrimaryLabel:   workspace.ClassName,
			SecondaryLabel: fields["学年"],
		})
	}
	return options
}

func defaultWorkspaces() []classroom.Workspace {
	return []classroom.Workspace{
		defaultWorkspace(),
		{
			ClassID:   "admin-class-g1-8",
			ClassName: "高一（8）班",
			Stage: classroom.SelectionStage{
				ID:          "pre-selection",
				Label:       "未选科阶段",
				Description: "适用于高一前期或尚未完成选科分流的班级，以行政班和通用科目分析为主。",
			},
			BaseFields: []classroom.BaseField{
				{Label: "学校", Value: "星河高级中学"},
				{Label: "年级", Value: "高一"},
				{Label: "行政班", Value: "高一（8）班"},
				{Label: "班主任", Value: "李老师"},
				{Label: "学年", Value: "2025-2026 学年"},
				{Label: "班级状态", Value: "正常使用中"},
			},
			Students: []classroom.Student{
				{ID: "student-g240801", StudentNo: "G240801", Name: "沈知夏", Gender: "女", Combination: "", ElectiveSubjects: []string{}, ParentMobile: "138****7101", Status: "已完整", ParentStatus: "已绑定", SelectionStatus: "待确认", ProfileStatus: "missing_selection"},
				{ID: "student-g240802", StudentNo: "G240802", Name: "陆明川", Gender: "男", Combination: "", ElectiveSubjects: []string{}, ParentMobile: "", Status: "待补家长", ParentStatus: "待补充", SelectionStatus: "待确认", ProfileStatus: "missing_parent"},
				{ID: "student-g240803", StudentNo: "G240803", Name: "唐一禾", Gender: "女", Combination: "", ElectiveSubjects: []string{}, ParentMobile: "139****6028", Status: "已完整", ParentStatus: "已绑定", SelectionStatus: "待确认", ProfileStatus: "missing_selection"},
			},
			Teachers: []classroom.TeacherAssignment{
				{ID: "teacher-g18-chinese-lu", Subject: "语文", Teacher: "卢老师", Mobile: "13800002001", Classes: "高一（8）班", SyncStatus: "已同步", AccountStatus: "bound", AccountID: "teacher-account-teacher-g18-chinese-lu", AccountBoundAt: "2026-06-03 09:00", PermissionStatus: "synced", PermissionSyncedAt: "2026-06-03 09:08"},
				{ID: "teacher-g18-math-wang", Subject: "数学", Teacher: "王老师", Mobile: "13800001002", Classes: "高一（8）班、高二（3）班", SyncStatus: "已同步", AccountStatus: "bound", AccountID: "teacher-account-teacher-g18-math-wang", AccountBoundAt: "2026-06-03 09:00", PermissionStatus: "synced", PermissionSyncedAt: "2026-06-03 09:08"},
				{ID: "teacher-g18-english-li", Subject: "英语", Teacher: "李老师", Mobile: "13800001003", Classes: "高一（8）班", SyncStatus: "班主任本人", AccountStatus: "bound", AccountID: "teacher-account-teacher-g18-english-li", AccountBoundAt: "2026-06-03 09:00", PermissionStatus: "synced", PermissionSyncedAt: "2026-06-03 09:08"},
			},
			Policies: defaultPolicies("pre-selection"),
		},
	}
}

func defaultWorkspace() classroom.Workspace {
	return classroom.Workspace{
		ClassID:   CurrentClassID,
		ClassName: "高二（3）班",
		Stage: classroom.SelectionStage{
			ID:          "post-selection",
			Label:       "已选科阶段",
			Description: "适用于完成选科后的班级，支持行政班、教学班、选科组合和赋分分析。",
		},
		BaseFields: []classroom.BaseField{
			{Label: "学校", Value: "星河高级中学"},
			{Label: "年级", Value: "高二"},
			{Label: "行政班", Value: "高二（3）班"},
			{Label: "班主任", Value: "李老师"},
			{Label: "学年", Value: "2025-2026 学年"},
			{Label: "班级状态", Value: "正常使用中"},
		},
		Students: []classroom.Student{
			{ID: "student-g230301", StudentNo: "G230301", Name: "林书言", Gender: "女", Combination: "物化生", ElectiveSubjects: []string{"物理", "化学", "生物"}, ParentMobile: "138****3201", Status: "已完整", ParentStatus: "已绑定", SelectionStatus: "已登记", ProfileStatus: "ready"},
			{ID: "student-g230302", StudentNo: "G230302", Name: "许一诺", Gender: "男", Combination: "物化生", ElectiveSubjects: []string{"物理", "化学", "生物"}, ParentMobile: "139****1882", Status: "已完整", ParentStatus: "已绑定", SelectionStatus: "已登记", ProfileStatus: "ready"},
			{ID: "student-g230317", StudentNo: "G230317", Name: "陈可心", Gender: "女", Combination: "史地政", ElectiveSubjects: []string{"历史", "地理", "政治"}, ParentMobile: "137****5210", Status: "待补家长", ParentStatus: "待补充", SelectionStatus: "已登记", ProfileStatus: "missing_parent"},
			{ID: "student-g230329", StudentNo: "G230329", Name: "赵博文", Gender: "男", Combination: "物化生", ElectiveSubjects: []string{"物理", "化学", "生物"}, ParentMobile: "136****9913", Status: "选科待确认", ParentStatus: "已绑定", SelectionStatus: "待确认", ProfileStatus: "missing_selection"},
		},
		Teachers: []classroom.TeacherAssignment{
			{ID: "teacher-chinese-zhang", Subject: "语文", Teacher: "张老师", Mobile: "13800001001", Classes: "高二（3）班、高二（5）班", SyncStatus: "已同步", AccountStatus: "bound", AccountID: "teacher-account-teacher-chinese-zhang", AccountBoundAt: "2026-06-01 09:00", PermissionStatus: "synced", PermissionSyncedAt: "2026-06-01 09:05"},
			{ID: "teacher-math-wang", Subject: "数学", Teacher: "王老师", Mobile: "13800001002", Classes: "高二（3）班、高一（8）班", SyncStatus: "已同步", AccountStatus: "bound", AccountID: "teacher-account-teacher-math-wang", AccountBoundAt: "2026-06-01 09:00", PermissionStatus: "synced", PermissionSyncedAt: "2026-06-01 09:05"},
			{ID: "teacher-english-li", Subject: "英语", Teacher: "李老师", Mobile: "13800001003", Classes: "高二（3）班", SyncStatus: "班主任本人", AccountStatus: "bound", AccountID: "teacher-account-teacher-english-li", AccountBoundAt: "2026-06-01 09:00", PermissionStatus: "synced", PermissionSyncedAt: "2026-06-01 09:05"},
			{ID: "teacher-chemistry-zhao", Subject: "化学", Teacher: "赵老师", Classes: "高二（3）班教学班", SyncStatus: "待补账号绑定", AccountStatus: "pending", PermissionStatus: "pending"},
		},
		Policies: defaultPolicies("post-selection"),
	}
}
