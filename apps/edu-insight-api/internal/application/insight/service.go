package insight

import (
	"fmt"
	"sort"
	"strconv"
	"time"

	appclassroom "github.com/RaghtNow/opc/apps/edu-insight-api/internal/application/classroom"
	appscore "github.com/RaghtNow/opc/apps/edu-insight-api/internal/application/score"
	domainclassroom "github.com/RaghtNow/opc/apps/edu-insight-api/internal/domain/classroom"
	"github.com/RaghtNow/opc/apps/edu-insight-api/internal/domain/insight"
	domainscore "github.com/RaghtNow/opc/apps/edu-insight-api/internal/domain/score"
)

type Service interface {
	GetDashboard() (insight.Dashboard, error)
	PublishLatestExam() (insight.Dashboard, error)
}

type Repository interface {
	ListSyncRecords(classID string) ([]insight.SyncRecord, error)
	SaveSyncRecords(classID string, records []insight.SyncRecord) error
}

type service struct {
	classroom appclassroom.Service
	score     appscore.Service
	repo      Repository
	classID   string
}

func NewService(classroomService appclassroom.Service, scoreService appscore.Service, repo Repository) Service {
	return &service{classroom: classroomService, score: scoreService, repo: repo, classID: appclassroom.CurrentClassID}
}

func (s *service) GetDashboard() (insight.Dashboard, error) {
	workspace, err := s.classroom.GetWorkspace()
	if err != nil {
		return insight.Dashboard{}, err
	}
	exams, err := s.score.ListExams()
	if err != nil {
		return insight.Dashboard{}, err
	}
	detail := domainscore.ExamDetail{}
	if len(exams) > 0 {
		if item, ok, err := s.score.GetExamDetail(exams[0].ID); err != nil {
			return insight.Dashboard{}, err
		} else if ok {
			detail = item
		}
	}
	records, err := s.repo.ListSyncRecords(s.classID)
	if err != nil {
		return insight.Dashboard{}, err
	}
	return buildDashboard(workspace, detail, records), nil
}

func (s *service) PublishLatestExam() (insight.Dashboard, error) {
	dashboard, err := s.GetDashboard()
	if err != nil {
		return insight.Dashboard{}, err
	}
	if !dashboard.CanPublish {
		return dashboard, fmt.Errorf("当前不满足发布条件：%v", dashboard.PublishBlockers)
	}
	now := time.Now().Format("2006-01-02 15:04")
	records := []insight.SyncRecord{
		{Target: "家长（当前班级）", Channel: "微信小程序订阅消息", Status: "已创建发布任务", Time: now},
		{Target: "学生（当前班级）", Channel: "微信小程序", Status: "已创建发布任务", Time: now},
		{Target: "任课老师", Channel: "站内消息", Status: "已创建发布任务", Time: now},
	}
	if err := s.repo.SaveSyncRecords(s.classID, records); err != nil {
		return insight.Dashboard{}, err
	}
	return s.GetDashboard()
}

func buildDashboard(workspace domainclassroom.Workspace, detail domainscore.ExamDetail, records []insight.SyncRecord) insight.Dashboard {
	totalStudents := len(workspace.Students)
	parentReady := 0
	for _, student := range workspace.Students {
		if student.ParentStatus == "已绑定" {
			parentReady++
		}
	}
	teacherSynced := 0
	for _, teacher := range workspace.Teachers {
		if teacher.PermissionStatus == "synced" {
			teacherSynced++
		}
	}

	scoreRows := append([]domainscore.ScoreRow{}, detail.Scores...)
	sort.Slice(scoreRows, func(i, j int) bool { return parseScore(scoreRows[i].Total) > parseScore(scoreRows[j].Total) })
	average := averageTotal(scoreRows)
	alerts := buildAlerts(scoreRows, average)
	latestExamName := detail.Exam.Name
	if latestExamName == "" {
		latestExamName = "暂无考试"
	}
	blockers := []string{}
	if len(scoreRows) == 0 {
		blockers = append(blockers, "暂无可发布考试成绩")
	}
	if parentReady == 0 {
		blockers = append(blockers, "家长触达信息未就绪")
	}

	return insight.Dashboard{
		SummaryMetrics: []insight.SummaryMetric{
			{Label: "当前班级", Value: workspace.ClassName, Note: fmt.Sprintf("%d 人 / %s", totalStudents, workspace.Stage.Label)},
			{Label: "最近考试", Value: latestExamName, Note: fmt.Sprintf("%d 条成绩", len(scoreRows))},
			{Label: "班级均分", Value: formatFloat(average), Note: "基于最近一次考试总分"},
			{Label: "重点预警", Value: fmt.Sprintf("%d 人", len(alerts)), Note: "低于班级均分或单科偏弱"},
		},
		StudentTrends:     buildStudentTrends(scoreRows, average),
		CohortInsights:    buildCohorts(scoreRows, average),
		AlertItems:        alerts,
		SyncAudienceCards: buildSyncCards(totalStudents, parentReady, len(workspace.Teachers), teacherSynced),
		SyncRecords:       records,
		LatestExamName:    latestExamName,
		CanPublish:        len(blockers) == 0,
		PublishBlockers:   blockers,
	}
}

func buildStudentTrends(rows []domainscore.ScoreRow, average float64) []insight.StudentTrend {
	limit := min(len(rows), 4)
	items := []insight.StudentTrend{}
	for i := 0; i < limit; i++ {
		total := parseScore(rows[i].Total)
		tag := "稳定"
		delta := "持平"
		if total >= average+30 {
			tag = "高分保持"
			delta = "高于均分"
		} else if total < average-30 {
			tag = "重点关注"
			delta = "低于均分"
		}
		items = append(items, insight.StudentTrend{Name: rows[i].StudentName, TotalScore: rows[i].Total, Delta: delta, Tag: tag})
	}
	return items
}

func buildCohorts(rows []domainscore.ScoreRow, average float64) []insight.CohortInsight {
	high, middle, support := 0, 0, 0
	for _, row := range rows {
		total := parseScore(row.Total)
		if total >= average+30 {
			high++
		} else if total < average-30 {
			support++
		} else {
			middle++
		}
	}
	return []insight.CohortInsight{
		{Title: "高分稳定群体", Students: fmt.Sprintf("%d 人", high), Insight: "总分显著高于班级均分，适合继续做拔高跟踪。"},
		{Title: "临界波动群体", Students: fmt.Sprintf("%d 人", middle), Insight: "接近班级均分，是班主任和单科老师联合提升的核心人群。"},
		{Title: "重点帮扶群体", Students: fmt.Sprintf("%d 人", support), Insight: "低于班级均分较多，建议进入预警和家校沟通链路。"},
	}
}

func buildAlerts(rows []domainscore.ScoreRow, average float64) []insight.AlertItem {
	alerts := []insight.AlertItem{}
	for _, row := range rows {
		total := parseScore(row.Total)
		if total < average-30 {
			alerts = append(alerts, insight.AlertItem{Student: row.StudentName, Subject: "总分", Level: "高", Detail: fmt.Sprintf("总分 %s，低于班级均分 %s", row.Total, formatFloat(average))})
		}
		for subject, raw := range row.SubjectScores {
			if parseScore(raw) > 0 && parseScore(raw) < 60 {
				alerts = append(alerts, insight.AlertItem{Student: row.StudentName, Subject: subject, Level: "中", Detail: fmt.Sprintf("%s 分数为 %s，低于预警线", subject, raw)})
			}
		}
		if len(alerts) >= 6 {
			return alerts
		}
	}
	return alerts
}

func buildSyncCards(totalStudents, parentReady, totalTeachers, teacherSynced int) []insight.SyncAudienceCard {
	return []insight.SyncAudienceCard{
		{Audience: "家长", Readiness: fmt.Sprintf("%d / %d 已可触达", parentReady, totalStudents), Note: "只向已绑定家长发送，未绑定学生会保留待补状态。"},
		{Audience: "学生", Readiness: fmt.Sprintf("%d / %d 已可查看", totalStudents, totalStudents), Note: "学生端查看能力由后续小程序身份绑定控制。"},
		{Audience: "任课老师", Readiness: fmt.Sprintf("%d / %d 已授权", teacherSynced, totalTeachers), Note: "任课老师需完成账号绑定和权限同步后可查看授权范围。"},
	}
}

func averageTotal(rows []domainscore.ScoreRow) float64 {
	if len(rows) == 0 {
		return 0
	}
	total := 0.0
	for _, row := range rows {
		total += parseScore(row.Total)
	}
	return total / float64(len(rows))
}

func parseScore(value string) float64 {
	number, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0
	}
	return number
}

func formatFloat(value float64) string {
	if value == float64(int64(value)) {
		return fmt.Sprintf("%d", int64(value))
	}
	return strconv.FormatFloat(value, 'f', 1, 64)
}
