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
	GetWorkspace() (classroom.Workspace, error)
	CreateStudent(req classroom.SaveStudentRequest) (classroom.Workspace, error)
	UpdateStudent(id string, req classroom.SaveStudentRequest) (classroom.Workspace, bool, error)
	ImportStudents(fileName string, content []byte) (classroom.Workspace, classroom.ImportSummary, error)
	CreateTeacher(req classroom.SaveTeacherRequest) (classroom.Workspace, error)
	UpdateTeacher(id string, req classroom.SaveTeacherRequest) (classroom.Workspace, bool, error)
	ImportTeachers(fileName string, content []byte) (classroom.Workspace, classroom.ImportSummary, error)
	BindTeacherAccount(id string) (classroom.Workspace, bool, error)
	SyncTeacherPermission(id string) (classroom.Workspace, bool, error)
	UpdatePolicy(id string, req classroom.SavePolicyRequest) (classroom.Workspace, bool, error)
}

type Repository interface {
	GetWorkspace(classID string) (classroom.Workspace, error)
	SaveStudent(classID string, student classroom.Student) error
	SaveTeacher(classID string, teacher classroom.TeacherAssignment) error
	SavePolicy(classID string, policy classroom.Policy) error
	SeedIfEmpty(workspace classroom.Workspace) error
}

type PersistentService struct {
	repo    Repository
	classID string
}

func NewPersistentService(repo Repository) (*PersistentService, error) {
	service := &PersistentService{repo: repo, classID: CurrentClassID}
	if err := repo.SeedIfEmpty(defaultWorkspace()); err != nil {
		return nil, err
	}
	return service, nil
}

func (s *PersistentService) GetWorkspace() (classroom.Workspace, error) {
	return s.repo.GetWorkspace(s.classID)
}

func (s *PersistentService) CreateStudent(req classroom.SaveStudentRequest) (classroom.Workspace, error) {
	student := studentFromRequest(fmt.Sprintf("student-%d", time.Now().UnixNano()), req)
	if student.StudentNo == "" {
		workspace, err := s.repo.GetWorkspace(s.classID)
		if err != nil {
			return classroom.Workspace{}, err
		}
		student.StudentNo = fmt.Sprintf("NEW%03d", len(workspace.Students)+1)
	}
	if student.Name == "" {
		student.Name = "新学生"
	}
	if err := s.repo.SaveStudent(s.classID, student); err != nil {
		return classroom.Workspace{}, err
	}
	return s.repo.GetWorkspace(s.classID)
}

func (s *PersistentService) UpdateStudent(id string, req classroom.SaveStudentRequest) (classroom.Workspace, bool, error) {
	workspace, err := s.repo.GetWorkspace(s.classID)
	if err != nil {
		return classroom.Workspace{}, false, err
	}
	if !studentExists(workspace.Students, id) {
		return classroom.Workspace{}, false, nil
	}
	if err := s.repo.SaveStudent(s.classID, studentFromRequest(id, req)); err != nil {
		return classroom.Workspace{}, false, err
	}
	workspace, err = s.repo.GetWorkspace(s.classID)
	return workspace, true, err
}

func (s *PersistentService) ImportStudents(fileName string, content []byte) (classroom.Workspace, classroom.ImportSummary, error) {
	rows, err := parseCSVRows(fileName, content)
	if err != nil {
		return classroom.Workspace{}, classroom.ImportSummary{}, err
	}
	workspace, err := s.repo.GetWorkspace(s.classID)
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
		if err := s.repo.SaveStudent(s.classID, student); err != nil {
			return classroom.Workspace{}, classroom.ImportSummary{}, err
		}
	}
	workspace, err = s.repo.GetWorkspace(s.classID)
	return workspace, summary, err
}

func (s *PersistentService) CreateTeacher(req classroom.SaveTeacherRequest) (classroom.Workspace, error) {
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
	if err := s.repo.SaveTeacher(s.classID, teacher); err != nil {
		return classroom.Workspace{}, err
	}
	return s.repo.GetWorkspace(s.classID)
}

func (s *PersistentService) UpdateTeacher(id string, req classroom.SaveTeacherRequest) (classroom.Workspace, bool, error) {
	workspace, err := s.repo.GetWorkspace(s.classID)
	if err != nil {
		return classroom.Workspace{}, false, err
	}
	for _, teacher := range workspace.Teachers {
		if teacher.ID == id {
			teacher.Subject = req.Subject
			teacher.Teacher = req.Teacher
			teacher.Mobile = req.Mobile
			teacher.Classes = req.Classes
			if err := s.repo.SaveTeacher(s.classID, normalizeTeacherState(teacher)); err != nil {
				return classroom.Workspace{}, false, err
			}
			workspace, err = s.repo.GetWorkspace(s.classID)
			return workspace, true, err
		}
	}
	return classroom.Workspace{}, false, nil
}

func (s *PersistentService) ImportTeachers(fileName string, content []byte) (classroom.Workspace, classroom.ImportSummary, error) {
	rows, err := parseCSVRows(fileName, content)
	if err != nil {
		return classroom.Workspace{}, classroom.ImportSummary{}, err
	}
	workspace, err := s.repo.GetWorkspace(s.classID)
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
		if err := s.repo.SaveTeacher(s.classID, teacher); err != nil {
			return classroom.Workspace{}, classroom.ImportSummary{}, err
		}
	}
	workspace, err = s.repo.GetWorkspace(s.classID)
	return workspace, summary, err
}

func (s *PersistentService) BindTeacherAccount(id string) (classroom.Workspace, bool, error) {
	return s.updateTeacherState(id, func(teacher classroom.TeacherAssignment) (classroom.TeacherAssignment, error) {
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

func (s *PersistentService) SyncTeacherPermission(id string) (classroom.Workspace, bool, error) {
	return s.updateTeacherState(id, func(teacher classroom.TeacherAssignment) (classroom.TeacherAssignment, error) {
		if teacher.AccountStatus != "bound" {
			return teacher, fmt.Errorf("请先绑定教师账号，再同步权限")
		}
		teacher.PermissionStatus = "synced"
		teacher.PermissionSyncedAt = time.Now().Format("2006-01-02 15:04")
		return teacher, nil
	})
}

func (s *PersistentService) updateTeacherState(id string, mutate func(classroom.TeacherAssignment) (classroom.TeacherAssignment, error)) (classroom.Workspace, bool, error) {
	workspace, err := s.repo.GetWorkspace(s.classID)
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
			if err := s.repo.SaveTeacher(s.classID, teacher); err != nil {
				return classroom.Workspace{}, false, err
			}
			workspace, err = s.repo.GetWorkspace(s.classID)
			return workspace, true, err
		}
	}
	return classroom.Workspace{}, false, nil
}

func (s *PersistentService) UpdatePolicy(id string, req classroom.SavePolicyRequest) (classroom.Workspace, bool, error) {
	workspace, err := s.repo.GetWorkspace(s.classID)
	if err != nil {
		return classroom.Workspace{}, false, err
	}
	for _, policy := range workspace.Policies {
		if policy.ID == id {
			policy.Value = req.Value
			policy.Note = req.Note
			if err := s.repo.SavePolicy(s.classID, policy); err != nil {
				return classroom.Workspace{}, false, err
			}
			workspace, err = s.repo.GetWorkspace(s.classID)
			return workspace, true, err
		}
	}
	return classroom.Workspace{}, false, nil
}

type MemoryService struct {
	mu        sync.RWMutex
	workspace classroom.Workspace
}

func NewMemoryService() *MemoryService {
	return &MemoryService{workspace: withRosterInsights(defaultWorkspace())}
}

func (s *MemoryService) GetWorkspace() (classroom.Workspace, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return snapshot(s.workspace), nil
}

func (s *MemoryService) CreateStudent(req classroom.SaveStudentRequest) (classroom.Workspace, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	student := studentFromRequest(fmt.Sprintf("student-%d", time.Now().UnixNano()), req)
	if student.StudentNo == "" {
		student.StudentNo = fmt.Sprintf("NEW%03d", len(s.workspace.Students)+1)
	}
	if student.Name == "" {
		student.Name = "新学生"
	}
	s.workspace.Students = append([]classroom.Student{student}, s.workspace.Students...)
	s.workspace = withRosterInsights(s.workspace)
	return snapshot(s.workspace), nil
}

func (s *MemoryService) UpdateStudent(id string, req classroom.SaveStudentRequest) (classroom.Workspace, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.workspace.Students {
		if s.workspace.Students[i].ID == id {
			s.workspace.Students[i] = studentFromRequest(id, req)
			s.workspace = withRosterInsights(s.workspace)
			return snapshot(s.workspace), true, nil
		}
	}
	return classroom.Workspace{}, false, nil
}

func (s *MemoryService) ImportStudents(fileName string, content []byte) (classroom.Workspace, classroom.ImportSummary, error) {
	rows, err := parseCSVRows(fileName, content)
	if err != nil {
		return classroom.Workspace{}, classroom.ImportSummary{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	byNo := map[string]int{}
	for index, student := range s.workspace.Students {
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
			student.ID = s.workspace.Students[existingIndex].ID
			s.workspace.Students[existingIndex] = student
			summary.Updated++
		} else {
			s.workspace.Students = append([]classroom.Student{student}, s.workspace.Students...)
			summary.Created++
		}
	}
	s.workspace = withRosterInsights(s.workspace)
	return snapshot(s.workspace), summary, nil
}

func (s *MemoryService) CreateTeacher(req classroom.SaveTeacherRequest) (classroom.Workspace, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

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
	s.workspace.Teachers = append([]classroom.TeacherAssignment{teacher}, s.workspace.Teachers...)
	s.workspace = withRosterInsights(s.workspace)
	return snapshot(s.workspace), nil
}

func (s *MemoryService) UpdateTeacher(id string, req classroom.SaveTeacherRequest) (classroom.Workspace, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.workspace.Teachers {
		if s.workspace.Teachers[i].ID == id {
			s.workspace.Teachers[i].Subject = req.Subject
			s.workspace.Teachers[i].Teacher = req.Teacher
			s.workspace.Teachers[i].Mobile = req.Mobile
			s.workspace.Teachers[i].Classes = req.Classes
			s.workspace.Teachers[i] = normalizeTeacherState(s.workspace.Teachers[i])
			s.workspace = withRosterInsights(s.workspace)
			return snapshot(s.workspace), true, nil
		}
	}
	return classroom.Workspace{}, false, nil
}

func (s *MemoryService) ImportTeachers(fileName string, content []byte) (classroom.Workspace, classroom.ImportSummary, error) {
	rows, err := parseCSVRows(fileName, content)
	if err != nil {
		return classroom.Workspace{}, classroom.ImportSummary{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	byKey := map[string]int{}
	for index, teacher := range s.workspace.Teachers {
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
			teacher.ID = s.workspace.Teachers[existingIndex].ID
			teacher.AccountStatus = s.workspace.Teachers[existingIndex].AccountStatus
			teacher.AccountID = s.workspace.Teachers[existingIndex].AccountID
			teacher.AccountBoundAt = s.workspace.Teachers[existingIndex].AccountBoundAt
			teacher.PermissionStatus = s.workspace.Teachers[existingIndex].PermissionStatus
			teacher.PermissionSyncedAt = s.workspace.Teachers[existingIndex].PermissionSyncedAt
			s.workspace.Teachers[existingIndex] = teacher
			summary.Updated++
		} else {
			s.workspace.Teachers = append([]classroom.TeacherAssignment{teacher}, s.workspace.Teachers...)
			summary.Created++
		}
	}
	s.workspace = withRosterInsights(s.workspace)
	return snapshot(s.workspace), summary, nil
}

func (s *MemoryService) BindTeacherAccount(id string) (classroom.Workspace, bool, error) {
	return s.updateTeacherState(id, func(teacher classroom.TeacherAssignment) (classroom.TeacherAssignment, error) {
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

func (s *MemoryService) SyncTeacherPermission(id string) (classroom.Workspace, bool, error) {
	return s.updateTeacherState(id, func(teacher classroom.TeacherAssignment) (classroom.TeacherAssignment, error) {
		if teacher.AccountStatus != "bound" {
			return teacher, fmt.Errorf("请先绑定教师账号，再同步权限")
		}
		teacher.PermissionStatus = "synced"
		teacher.PermissionSyncedAt = time.Now().Format("2006-01-02 15:04")
		return teacher, nil
	})
}

func (s *MemoryService) updateTeacherState(id string, mutate func(classroom.TeacherAssignment) (classroom.TeacherAssignment, error)) (classroom.Workspace, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.workspace.Teachers {
		if s.workspace.Teachers[i].ID == id {
			teacher, err := mutate(s.workspace.Teachers[i])
			if err != nil {
				return classroom.Workspace{}, false, err
			}
			s.workspace.Teachers[i] = normalizeTeacherState(teacher)
			s.workspace = withRosterInsights(s.workspace)
			return snapshot(s.workspace), true, nil
		}
	}
	return classroom.Workspace{}, false, nil
}

func (s *MemoryService) UpdatePolicy(id string, req classroom.SavePolicyRequest) (classroom.Workspace, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.workspace.Policies {
		if s.workspace.Policies[i].ID == id {
			s.workspace.Policies[i].Value = req.Value
			s.workspace.Policies[i].Note = req.Note
			return snapshot(s.workspace), true, nil
		}
	}
	return classroom.Workspace{}, false, nil
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
		Policies: []classroom.Policy{
			{ID: "policy-parent-class-rank", Title: "家长端班级位置", Value: "班主任可配置", Note: "当前班级允许展示班级位置，不允许展示年级排名。"},
			{ID: "policy-student-band", Title: "学生端分数段", Value: "已开启", Note: "班主任已开启班级分数段和班级均分对比。"},
			{ID: "policy-sync-trigger", Title: "同步策略", Value: "考试发布后触发", Note: "成绩分析完成后，再统一同步家长、学生与任课老师。"},
		},
	}
}
