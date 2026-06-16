package classroom

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/RaghtNow/opc/apps/edu-insight-api/internal/domain/classroom"
)

type Service struct {
	mu        sync.RWMutex
	workspace classroom.Workspace
}

func NewService() *Service {
	service := &Service{}
	service.seed()
	return service
}

func (s *Service) GetWorkspace() classroom.Workspace {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.snapshotLocked()
}

func (s *Service) CreateStudent(req classroom.SaveStudentRequest) classroom.Workspace {
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
	s.refreshRosterInsightsLocked()
	return s.snapshotLocked()
}

func (s *Service) UpdateStudent(id string, req classroom.SaveStudentRequest) (classroom.Workspace, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.workspace.Students {
		if s.workspace.Students[i].ID == id {
			s.workspace.Students[i] = studentFromRequest(id, req)
			s.refreshRosterInsightsLocked()
			return s.snapshotLocked(), true
		}
	}
	return classroom.Workspace{}, false
}

func (s *Service) CreateTeacher(req classroom.SaveTeacherRequest) classroom.Workspace {
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
	s.refreshRosterInsightsLocked()
	return s.snapshotLocked()
}

func (s *Service) UpdateTeacher(id string, req classroom.SaveTeacherRequest) (classroom.Workspace, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.workspace.Teachers {
		if s.workspace.Teachers[i].ID == id {
			s.workspace.Teachers[i] = teacherFromRequest(id, req)
			s.refreshRosterInsightsLocked()
			return s.snapshotLocked(), true
		}
	}
	return classroom.Workspace{}, false
}

func (s *Service) UpdatePolicy(id string, req classroom.SavePolicyRequest) (classroom.Workspace, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.workspace.Policies {
		if s.workspace.Policies[i].ID == id {
			s.workspace.Policies[i].Value = req.Value
			s.workspace.Policies[i].Note = req.Note
			return s.snapshotLocked(), true
		}
	}
	return classroom.Workspace{}, false
}

func (s *Service) snapshotLocked() classroom.Workspace {
	snapshot := s.workspace
	snapshot.BaseFields = append([]classroom.BaseField{}, s.workspace.BaseFields...)
	snapshot.RosterInsights = append([]classroom.RosterInsight{}, s.workspace.RosterInsights...)
	snapshot.Students = append([]classroom.Student{}, s.workspace.Students...)
	snapshot.Teachers = append([]classroom.TeacherAssignment{}, s.workspace.Teachers...)
	snapshot.Policies = append([]classroom.Policy{}, s.workspace.Policies...)
	return snapshot
}

func (s *Service) refreshRosterInsightsLocked() {
	parentReady := 0
	selectionReady := 0
	for _, student := range s.workspace.Students {
		if student.ParentStatus == "已绑定" {
			parentReady++
		}
		if student.SelectionStatus == "已登记" {
			selectionReady++
		}
	}

	teacherReady := 0
	for _, teacher := range s.workspace.Teachers {
		if teacher.AccountStatus == "bound" {
			teacherReady++
		}
	}

	totalStudents := len(s.workspace.Students)
	totalTeachers := len(s.workspace.Teachers)
	s.workspace.RosterInsights = []classroom.RosterInsight{
		{
			Title:  "已完成家长绑定",
			Count:  fmt.Sprintf("%d / %d", parentReady, totalStudents),
			Detail: fmt.Sprintf("%d 位学生缺少有效家长手机号，影响成绩同步。", totalStudents-parentReady),
		},
		{
			Title:  "已完成选科登记",
			Count:  fmt.Sprintf("%d / %d", selectionReady, totalStudents),
			Detail: fmt.Sprintf("%d 位学生选科仍待确认，会影响选考科成绩校验。", totalStudents-selectionReady),
		},
		{
			Title:  "任课老师已绑定账号",
			Count:  fmt.Sprintf("%d / %d", teacherReady, totalTeachers),
			Detail: fmt.Sprintf("%d 位任课老师尚未绑定账号，暂无法自动同步单科分析。", totalTeachers-teacherReady),
		},
	}
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

	syncStatus := "待补账号绑定"
	if accountStatus == "bound" && permissionStatus == "synced" {
		syncStatus = "已同步"
	} else if accountStatus == "bound" {
		syncStatus = "待同步权限"
	}

	return classroom.TeacherAssignment{
		ID:               id,
		Subject:          req.Subject,
		Teacher:          req.Teacher,
		Classes:          req.Classes,
		SyncStatus:       syncStatus,
		AccountStatus:    accountStatus,
		PermissionStatus: permissionStatus,
	}
}

func splitCombination(combination string) []string {
	replacer := strings.NewReplacer("物", "物理,", "化", "化学,", "生", "生物,", "史", "历史,", "地", "地理,", "政", "政治,")
	expanded := strings.TrimSuffix(replacer.Replace(combination), ",")
	if expanded == "" || expanded == combination {
		return []string{}
	}
	return strings.Split(expanded, ",")
}

func (s *Service) seed() {
	s.workspace = classroom.Workspace{
		ClassID:   "admin-class-g2-3",
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
			{ID: "teacher-chinese-zhang", Subject: "语文", Teacher: "张老师", Classes: "高二（3）班、高二（5）班", SyncStatus: "已同步", AccountStatus: "bound", PermissionStatus: "synced"},
			{ID: "teacher-math-wang", Subject: "数学", Teacher: "王老师", Classes: "高二（3）班、高一（8）班", SyncStatus: "已同步", AccountStatus: "bound", PermissionStatus: "synced"},
			{ID: "teacher-english-li", Subject: "英语", Teacher: "李老师", Classes: "高二（3）班", SyncStatus: "班主任本人", AccountStatus: "bound", PermissionStatus: "synced"},
			{ID: "teacher-chemistry-zhao", Subject: "化学", Teacher: "赵老师", Classes: "高二（3）班教学班", SyncStatus: "待补账号绑定", AccountStatus: "pending", PermissionStatus: "pending"},
		},
		Policies: []classroom.Policy{
			{ID: "policy-parent-class-rank", Title: "家长端班级位置", Value: "班主任可配置", Note: "当前班级允许展示班级位置，不允许展示年级排名。"},
			{ID: "policy-student-band", Title: "学生端分数段", Value: "已开启", Note: "班主任已开启班级分数段和班级均分对比。"},
			{ID: "policy-sync-trigger", Title: "同步策略", Value: "考试发布后触发", Note: "成绩分析完成后，再统一同步家长、学生与任课老师。"},
		},
	}
	s.refreshRosterInsightsLocked()
}
