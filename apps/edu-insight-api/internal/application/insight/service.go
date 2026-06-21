package insight

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	appclassroom "github.com/RaghtNow/opc/apps/edu-insight-api/internal/application/classroom"
	appscore "github.com/RaghtNow/opc/apps/edu-insight-api/internal/application/score"
	domainclassroom "github.com/RaghtNow/opc/apps/edu-insight-api/internal/domain/classroom"
	"github.com/RaghtNow/opc/apps/edu-insight-api/internal/domain/insight"
	domainscore "github.com/RaghtNow/opc/apps/edu-insight-api/internal/domain/score"
)

type Service interface {
	GetDashboard(query DashboardQuery) (insight.Dashboard, error)
	PublishLatestExam(classID string) (insight.Dashboard, error)
}

type DashboardQuery struct {
	ClassID  string
	Scope    string
	ClassIDs []string
	ExamID   string
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

func (s *service) GetDashboard(query DashboardQuery) (insight.Dashboard, error) {
	classKey := defaultClassID(query.ClassID)
	if query.Scope == "overall" {
		classIDs := normalizedClassIDs(query.ClassIDs, classKey)
		return s.getOverallDashboard(classIDs)
	}
	workspace, err := s.classroom.GetWorkspace(classKey)
	if err != nil {
		return insight.Dashboard{}, err
	}
	exams, err := s.score.ListExams(classKey)
	if err != nil {
		return insight.Dashboard{}, err
	}
	details := []domainscore.ExamDetail{}
	detail := domainscore.ExamDetail{}
	for index, exam := range exams {
		item, ok, err := s.score.GetExamDetail(classKey, exam.ID)
		if err != nil {
			return insight.Dashboard{}, err
		}
		if ok {
			details = append(details, item)
		}
		if query.ExamID != "" && exam.ID == query.ExamID && ok {
			detail = item
		}
		if query.ExamID == "" && index == 0 && ok {
			detail = item
		}
	}
	records, err := s.repo.ListSyncRecords(classKey)
	if err != nil {
		return insight.Dashboard{}, err
	}
	return buildDashboard(workspace, detail, details, records, "single", workspace.ClassName, []string{classKey}), nil
}

func (s *service) PublishLatestExam(classID string) (insight.Dashboard, error) {
	classKey := defaultClassID(classID)
	dashboard, err := s.GetDashboard(DashboardQuery{ClassID: classKey, Scope: "single"})
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
	if err := s.repo.SaveSyncRecords(classKey, records); err != nil {
		return insight.Dashboard{}, err
	}
	return s.GetDashboard(DashboardQuery{ClassID: classKey, Scope: "single"})
}

func (s *service) getOverallDashboard(classIDs []string) (insight.Dashboard, error) {
	workspaces := []domainclassroom.Workspace{}
	latestDetails := []domainscore.ExamDetail{}
	allDetails := []domainscore.ExamDetail{}
	records := []insight.SyncRecord{}

	for _, classID := range classIDs {
		workspace, err := s.classroom.GetWorkspace(classID)
		if err != nil {
			return insight.Dashboard{}, err
		}
		exams, err := s.score.ListExams(classID)
		if err != nil {
			return insight.Dashboard{}, err
		}
		for index, exam := range exams {
			detail, ok, err := s.score.GetExamDetail(classID, exam.ID)
			if err != nil {
				return insight.Dashboard{}, err
			}
			if !ok {
				continue
			}
			prefixed := prefixExamDetail(classID, workspace.ClassName, detail)
			allDetails = append(allDetails, prefixed)
			if index == 0 {
				latestDetails = append(latestDetails, prefixed)
			}
		}
		classRecords, err := s.repo.ListSyncRecords(classID)
		if err != nil {
			return insight.Dashboard{}, err
		}
		records = append(records, classRecords...)
		workspaces = append(workspaces, workspace)
	}

	workspace := aggregateWorkspace(workspaces)
	latest := aggregateLatestDetail(latestDetails)
	details := aggregateDetailsByExamName(allDetails)
	dashboard := buildDashboard(workspace, latest, details, records, "overall", workspace.ClassName, classIDs)
	dashboard.Analysis.ClassComparisons = buildClassComparisons(workspaces, latestDetails)
	return dashboard, nil
}

func defaultClassID(classID string) string {
	if classID == "" {
		return appclassroom.CurrentClassID
	}
	return classID
}

func normalizedClassIDs(classIDs []string, fallback string) []string {
	seen := map[string]bool{}
	ids := []string{}
	for _, raw := range classIDs {
		id := strings.TrimSpace(raw)
		if id == "" || seen[id] {
			continue
		}
		ids = append(ids, id)
		seen[id] = true
	}
	if len(ids) == 0 {
		ids = append(ids, fallback)
	}
	return ids
}

func prefixExamDetail(classID, className string, detail domainscore.ExamDetail) domainscore.ExamDetail {
	prefixed := detail
	prefixed.Exam.ID = fmt.Sprintf("%s:%s", classID, detail.Exam.ID)
	prefixed.Exam.Subjects = append([]string{}, detail.Exam.Subjects...)
	prefixed.Scores = append([]domainscore.ScoreRow{}, detail.Scores...)
	prefixed.Issues = append([]domainscore.ImportIssue{}, detail.Issues...)
	for index := range prefixed.Scores {
		prefixed.Scores[index].ID = fmt.Sprintf("%s:%s", classID, prefixed.Scores[index].ID)
		prefixed.Scores[index].StudentID = fmt.Sprintf("%s:%s", classID, prefixed.Scores[index].StudentID)
		prefixed.Scores[index].StudentName = fmt.Sprintf("%s · %s", className, prefixed.Scores[index].StudentName)
		if prefixed.Scores[index].SubjectScores != nil {
			prefixed.Scores[index].SubjectScores = copySubjectScores(prefixed.Scores[index].SubjectScores)
		}
	}
	return prefixed
}

func copySubjectScores(source map[string]string) map[string]string {
	target := map[string]string{}
	for key, value := range source {
		target[key] = value
	}
	return target
}

func aggregateWorkspace(workspaces []domainclassroom.Workspace) domainclassroom.Workspace {
	if len(workspaces) == 0 {
		return domainclassroom.Workspace{ClassID: "overall", ClassName: "整体范围"}
	}
	base := workspaces[0]
	base.ClassID = "overall"
	base.ClassName = fmt.Sprintf("%s等 %d 个班级", workspaces[0].ClassName, len(workspaces))
	base.Students = []domainclassroom.Student{}
	base.Teachers = []domainclassroom.TeacherAssignment{}
	base.ClassOptions = []domainclassroom.ClassOption{}
	for _, workspace := range workspaces {
		base.Students = append(base.Students, workspace.Students...)
		base.Teachers = append(base.Teachers, workspace.Teachers...)
		base.ClassOptions = append(base.ClassOptions, workspace.ClassOptions...)
	}
	base.RosterInsights = []domainclassroom.RosterInsight{
		{Title: "覆盖班级", Count: fmt.Sprintf("%d 个", len(workspaces)), Detail: "整体范围聚合当前身份授权内的班级。"},
		{Title: "学生总数", Count: fmt.Sprintf("%d 人", len(base.Students)), Detail: "用于整体成绩分析和同步准备度判断。"},
	}
	return base
}

func aggregateLatestDetail(details []domainscore.ExamDetail) domainscore.ExamDetail {
	if len(details) == 0 {
		return domainscore.ExamDetail{}
	}
	return mergeExamDetails("overall-latest", "各班最近考试整体", maxExamDate(details), details)
}

func aggregateDetailsByExamName(details []domainscore.ExamDetail) []domainscore.ExamDetail {
	groups := map[string][]domainscore.ExamDetail{}
	keys := []string{}
	for _, detail := range details {
		key := detail.Exam.Name
		if key == "" {
			key = detail.Exam.Date
		}
		if _, ok := groups[key]; !ok {
			keys = append(keys, key)
		}
		groups[key] = append(groups[key], detail)
	}
	sort.Slice(keys, func(i, j int) bool {
		return maxExamDate(groups[keys[i]]) > maxExamDate(groups[keys[j]])
	})
	merged := []domainscore.ExamDetail{}
	for _, key := range keys {
		merged = append(merged, mergeExamDetails("overall-"+key, key, maxExamDate(groups[key]), groups[key]))
	}
	return merged
}

func mergeExamDetails(id, name, date string, details []domainscore.ExamDetail) domainscore.ExamDetail {
	subjectSet := map[string]bool{}
	scores := []domainscore.ScoreRow{}
	issues := []domainscore.ImportIssue{}
	for _, detail := range details {
		for _, subject := range detail.Exam.Subjects {
			subjectSet[subject] = true
		}
		scores = append(scores, detail.Scores...)
		issues = append(issues, detail.Issues...)
	}
	return domainscore.ExamDetail{
		Exam: domainscore.Exam{
			ID:              id,
			Name:            name,
			Type:            "整体分析",
			Date:            date,
			SubjectCoverage: joinLabels(sortedKeys(subjectSet)),
			Subjects:        sortedKeys(subjectSet),
			ImportStatus:    "聚合生成",
		},
		Scores: scores,
		Issues: issues,
	}
}

func buildClassComparisons(workspaces []domainclassroom.Workspace, latestDetails []domainscore.ExamDetail) []insight.ClassComparison {
	nameByClass := map[string]string{}
	for _, workspace := range workspaces {
		nameByClass[workspace.ClassID] = workspace.ClassName
	}
	items := []insight.ClassComparison{}
	for _, detail := range latestDetails {
		classID := strings.Split(detail.Exam.ID, ":")[0]
		rows := sortedScoreRows(detail.Scores)
		average := averageTotal(rows)
		highest, lowest := highestLowest(rows)
		riskLine := maxFloat(0, average-30)
		riskCount := 0
		for _, row := range rows {
			if parseScore(row.Total) < riskLine {
				riskCount++
			}
		}
		items = append(items, insight.ClassComparison{
			ClassID:      classID,
			ClassName:    defaultString(nameByClass[classID], classID),
			StudentCount: len(rows),
			Average:      average,
			Highest:      highest,
			Lowest:       lowest,
			RiskCount:    riskCount,
			ExamName:     detail.Exam.Name,
		})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Average > items[j].Average })
	return items
}

func maxExamDate(details []domainscore.ExamDetail) string {
	maxDate := ""
	for _, detail := range details {
		if detail.Exam.Date > maxDate {
			maxDate = detail.Exam.Date
		}
	}
	return maxDate
}

func buildDashboard(workspace domainclassroom.Workspace, detail domainscore.ExamDetail, details []domainscore.ExamDetail, records []insight.SyncRecord, scope string, scopeLabel string, sourceClassIDs []string) insight.Dashboard {
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
		Scope:          scope,
		ScopeLabel:     scopeLabel,
		SourceClassIDs: sourceClassIDs,
		SummaryMetrics: []insight.SummaryMetric{
			{Label: "当前班级", Value: workspace.ClassName, Note: fmt.Sprintf("%d 人 / %s", totalStudents, workspace.Stage.Label)},
			{Label: "最近考试", Value: latestExamName, Note: fmt.Sprintf("%d 条成绩", len(scoreRows))},
			{Label: "班级均分", Value: formatFloat(average), Note: "基于最近一次考试总分"},
			{Label: "重点预警", Value: fmt.Sprintf("%d 人", len(alerts)), Note: "低于班级均分或单科偏弱"},
		},
		StudentTrends:     buildStudentTrends(scoreRows, average),
		CohortInsights:    buildCohorts(scoreRows, average),
		AlertItems:        alerts,
		Analysis:          buildAnalysisDashboard(detail, details),
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

func buildAnalysisDashboard(latest domainscore.ExamDetail, details []domainscore.ExamDetail) insight.AnalysisDashboard {
	rows := sortedScoreRows(latest.Scores)
	average := averageTotal(rows)
	median := medianTotal(rows)
	highest, lowest := highestLowest(rows)
	excellenceLine := average + 30
	riskLine := maxFloat(0, average-30)
	highCount, middleCount, riskCount := layerCounts(rows, excellenceLine, riskLine)

	subjectDiagnostics := buildSubjectDiagnostics(latest.Exam.Subjects, rows)
	scoreBands := buildScoreBands(len(rows), highCount, middleCount, riskCount, excellenceLine, riskLine)
	layerGroups := buildLayerGroups(rows, excellenceLine, riskLine)
	riskStudents := buildRiskStudents(latest.Exam.Subjects, rows, average)

	return insight.AnalysisDashboard{
		ExamMetrics: []insight.ExamMetric{
			{Label: "班级均分", Value: formatFloat(average), Note: fmt.Sprintf("%s / %d 人", defaultString(latest.Exam.Name, "当前考试"), len(rows))},
			{Label: "中位分", Value: formatFloat(median), Note: fmt.Sprintf("最高 %s / 最低 %s", formatFloat(highest), formatFloat(lowest))},
			{Label: "高分层", Value: fmt.Sprintf("%d 人", highCount), Note: "高于均分 30 分以上"},
			{Label: "预警层", Value: fmt.Sprintf("%d 人", riskCount), Note: "低于均分 30 分以上"},
		},
		SubjectDiagnostics: subjectDiagnostics,
		ScoreBands:         scoreBands,
		LayerGroups:        layerGroups,
		RiskStudents:       riskStudents,
		TeachingActions:    buildTeachingActions(subjectDiagnostics, scoreBands, riskStudents, highest-lowest),
		ClassTrend:         buildClassTrend(details),
		SubjectTrends:      buildSubjectTrends(details),
		StudentAnalyses:    buildStudentAnalyses(details),
	}
}

func sortedScoreRows(rows []domainscore.ScoreRow) []domainscore.ScoreRow {
	items := append([]domainscore.ScoreRow{}, rows...)
	sort.Slice(items, func(i, j int) bool { return parseScore(items[i].Total) > parseScore(items[j].Total) })
	return items
}

func medianTotal(rows []domainscore.ScoreRow) float64 {
	if len(rows) == 0 {
		return 0
	}
	values := []float64{}
	for _, row := range rows {
		values = append(values, parseScore(row.Total))
	}
	sort.Float64s(values)
	mid := len(values) / 2
	if len(values)%2 == 0 {
		return (values[mid-1] + values[mid]) / 2
	}
	return values[mid]
}

func highestLowest(rows []domainscore.ScoreRow) (float64, float64) {
	if len(rows) == 0 {
		return 0, 0
	}
	highest := parseScore(rows[0].Total)
	lowest := highest
	for _, row := range rows {
		value := parseScore(row.Total)
		if value > highest {
			highest = value
		}
		if value < lowest {
			lowest = value
		}
	}
	return highest, lowest
}

func layerCounts(rows []domainscore.ScoreRow, excellenceLine, riskLine float64) (int, int, int) {
	high, middle, risk := 0, 0, 0
	for _, row := range rows {
		total := parseScore(row.Total)
		if total >= excellenceLine {
			high++
		} else if total < riskLine {
			risk++
		} else {
			middle++
		}
	}
	return high, middle, risk
}

func buildSubjectDiagnostics(subjects []string, rows []domainscore.ScoreRow) []insight.SubjectDiagnostic {
	items := []insight.SubjectDiagnostic{}
	for _, subject := range subjects {
		values := []float64{}
		for _, row := range rows {
			if value := parseScore(row.SubjectScores[subject]); value > 0 {
				values = append(values, value)
			}
		}
		avg := average(values)
		highest, lowest := highestLowestValues(values)
		lowCount := countBy(values, func(value float64) bool { return value < 60 })
		excellentCount := countBy(values, func(value float64) bool { return value >= 90 })
		passCount := countBy(values, func(value float64) bool { return value >= 60 })
		riskLabel := "暂无低分预警"
		if lowCount > 0 {
			riskLabel = fmt.Sprintf("%d 人低于 60", lowCount)
		}
		items = append(items, insight.SubjectDiagnostic{
			Subject:        subject,
			Average:        avg,
			Highest:        highest,
			Lowest:         lowest,
			ExcellentCount: excellentCount,
			PassCount:      passCount,
			LowCount:       lowCount,
			BarWidth:       clamp((avg/maxFloat(highest, 150))*100, 8, 100),
			RiskLabel:      riskLabel,
		})
	}
	return items
}

func buildScoreBands(total, high, middle, risk int, excellenceLine, riskLine float64) []insight.ScoreBand {
	denominator := maxInt(total, 1)
	return []insight.ScoreBand{
		scoreBand("高分突破", fmt.Sprintf("≥ %s", formatFloat(excellenceLine)), high, denominator, "strong"),
		scoreBand("稳定中坚", fmt.Sprintf("%s - %s", formatFloat(riskLine), formatFloat(excellenceLine)), middle, denominator, "steady"),
		scoreBand("重点帮扶", fmt.Sprintf("< %s", formatFloat(riskLine)), risk, denominator, "risk"),
	}
}

func scoreBand(label, scoreRange string, count, total int, tone string) insight.ScoreBand {
	percent := int(float64(count) / float64(total) * 100)
	width := maxInt(8, percent)
	return insight.ScoreBand{Label: label, Range: scoreRange, Count: count, Percent: percent, Width: width, Tone: tone}
}

func buildLayerGroups(rows []domainscore.ScoreRow, excellenceLine, riskLine float64) []insight.LayerGroup {
	high, middle, risk := []domainscore.ScoreRow{}, []domainscore.ScoreRow{}, []domainscore.ScoreRow{}
	for _, row := range rows {
		total := parseScore(row.Total)
		if total >= excellenceLine {
			high = append(high, row)
		} else if total < riskLine {
			risk = append(risk, row)
		} else {
			middle = append(middle, row)
		}
	}
	return []insight.LayerGroup{
		{Title: "高分突破组", Count: fmt.Sprintf("%d 人", len(high)), Students: namesOf(high), Goal: "保持优势科，安排压轴题和高阶迁移训练。"},
		{Title: "临界提升组", Count: fmt.Sprintf("%d 人", len(middle)), Students: namesOf(middle), Goal: "找出最短板学科，优先做 2 周小目标提升。"},
		{Title: "重点帮扶组", Count: fmt.Sprintf("%d 人", len(risk)), Students: namesOf(risk), Goal: "班主任、单科老师、家长同步跟进，先止跌再补弱。"},
	}
}

func buildRiskStudents(subjects []string, rows []domainscore.ScoreRow, average float64) []insight.RiskStudent {
	items := []insight.RiskStudent{}
	for _, row := range rows {
		weakSubjects := []string{}
		for _, subject := range subjects {
			if value := parseScore(row.SubjectScores[subject]); value > 0 && value < 60 {
				weakSubjects = append(weakSubjects, subject)
			}
		}
		gap := average - parseScore(row.Total)
		level := "观察"
		if gap >= 45 || len(weakSubjects) >= 2 {
			level = "高"
		} else if gap >= 20 || len(weakSubjects) == 1 {
			level = "中"
		}
		if level == "观察" {
			continue
		}
		reason := fmt.Sprintf("低于均分 %s 分", formatFloat(gap))
		if gap < 0 {
			reason = fmt.Sprintf("高于均分 %s 分", formatFloat(-gap))
		}
		weakLabel := "暂无明显低分科"
		if len(weakSubjects) > 0 {
			weakLabel = joinLabels(weakSubjects)
		}
		items = append(items, insight.RiskStudent{Name: row.StudentName, Total: row.Total, Gap: gap, WeakSubjects: weakLabel, Level: level, Reason: reason})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Gap > items[j].Gap })
	return items
}

func buildTeachingActions(subjects []insight.SubjectDiagnostic, bands []insight.ScoreBand, risks []insight.RiskStudent, spread float64) []insight.TeachingAction {
	weakest := insight.SubjectDiagnostic{}
	for index, subject := range subjects {
		if index == 0 || subject.Average < weakest.Average {
			weakest = subject
		}
	}
	riskCount := 0
	if len(bands) >= 3 {
		riskCount = bands[2].Count
	}
	actions := []insight.TeachingAction{
		{Title: "先抓班级最弱学科", Detail: "暂无学科成绩数据。", Tag: "待导入"},
		{Title: "建立动态小组", Detail: fmt.Sprintf("按高分突破、临界提升、重点帮扶三组推进，当前重点帮扶 %d 人。", riskCount), Tag: "分层教学"},
		{Title: "优先跟进高风险学生", Detail: "当前没有明显高风险学生。", Tag: "稳定"},
		{Title: "下次考试验证目标", Detail: fmt.Sprintf("建议把班级均分提升 %d 分、重点帮扶组减少 1 人作为下一次阶段目标。", maxInt(3, int(spread*0.04))), Tag: "闭环验证"},
	}
	if weakest.Subject != "" {
		actions[0] = insight.TeachingAction{Title: "先抓班级最弱学科", Detail: fmt.Sprintf("%s 均分 %s，%s，建议先做错因归类和小测回收。", weakest.Subject, formatFloat(weakest.Average), weakest.RiskLabel), Tag: weakest.Subject}
	}
	if len(risks) > 0 {
		actions[2] = insight.TeachingAction{Title: "优先跟进高风险学生", Detail: fmt.Sprintf("%s %s，薄弱点：%s。", risks[0].Name, risks[0].Reason, risks[0].WeakSubjects), Tag: risks[0].Level}
	}
	return actions
}

func buildClassTrend(details []domainscore.ExamDetail) []insight.TrendPoint {
	items := []insight.TrendPoint{}
	for i := len(details) - 1; i >= 0; i-- {
		detail := details[i]
		items = append(items, insight.TrendPoint{ExamID: detail.Exam.ID, ExamName: detail.Exam.Name, Date: detail.Exam.Date, Value: averageTotal(detail.Scores)})
	}
	return items
}

func buildSubjectTrends(details []domainscore.ExamDetail) []insight.SubjectTrend {
	subjectSet := map[string]bool{}
	for _, detail := range details {
		for _, subject := range detail.Exam.Subjects {
			subjectSet[subject] = true
		}
	}
	subjects := sortedKeys(subjectSet)
	trends := []insight.SubjectTrend{}
	for _, subject := range subjects {
		points := []insight.TrendPoint{}
		for i := len(details) - 1; i >= 0; i-- {
			detail := details[i]
			values := []float64{}
			for _, row := range detail.Scores {
				if value := parseScore(row.SubjectScores[subject]); value > 0 {
					values = append(values, value)
				}
			}
			if len(values) > 0 {
				points = append(points, insight.TrendPoint{ExamID: detail.Exam.ID, ExamName: detail.Exam.Name, Date: detail.Exam.Date, Value: average(values)})
			}
		}
		trends = append(trends, insight.SubjectTrend{Subject: subject, Points: points})
	}
	return trends
}

func buildStudentAnalyses(details []domainscore.ExamDetail) []insight.StudentAnalysis {
	byStudent := map[string]*insight.StudentAnalysis{}
	for i := len(details) - 1; i >= 0; i-- {
		detail := details[i]
		ranked := sortedScoreRows(detail.Scores)
		for rank, row := range ranked {
			key := row.StudentID
			if key == "" {
				key = row.StudentName
			}
			item := byStudent[key]
			if item == nil {
				item = &insight.StudentAnalysis{StudentID: row.StudentID, StudentName: row.StudentName}
				byStudent[key] = item
			}
			item.TotalTrend = append(item.TotalTrend, insight.TrendPoint{ExamID: detail.Exam.ID, ExamName: detail.Exam.Name, Date: detail.Exam.Date, Value: parseScore(row.Total)})
			item.RankTrend = append(item.RankTrend, insight.TrendPoint{ExamID: detail.Exam.ID, ExamName: detail.Exam.Name, Date: detail.Exam.Date, Value: float64(rank + 1)})
			if i == 0 {
				item.LatestRank = rank + 1
				item.LatestTotal = row.Total
				item.WeakSubjects = weakSubjectsLabel(detail.Exam.Subjects, row)
				item.Recommendation = studentRecommendation(item.WeakSubjects)
			}
			item.SubjectTrends = appendStudentSubjectTrends(item.SubjectTrends, detail, row)
		}
	}
	items := []insight.StudentAnalysis{}
	for _, item := range byStudent {
		items = append(items, *item)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].LatestRank < items[j].LatestRank })
	return items
}

func appendStudentSubjectTrends(existing []insight.SubjectTrend, detail domainscore.ExamDetail, row domainscore.ScoreRow) []insight.SubjectTrend {
	indexBySubject := map[string]int{}
	for index, item := range existing {
		indexBySubject[item.Subject] = index
	}
	for _, subject := range detail.Exam.Subjects {
		value := parseScore(row.SubjectScores[subject])
		if value <= 0 {
			continue
		}
		point := insight.TrendPoint{ExamID: detail.Exam.ID, ExamName: detail.Exam.Name, Date: detail.Exam.Date, Value: value}
		if index, ok := indexBySubject[subject]; ok {
			existing[index].Points = append(existing[index].Points, point)
		} else {
			existing = append(existing, insight.SubjectTrend{Subject: subject, Points: []insight.TrendPoint{point}})
		}
	}
	return existing
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

func average(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	total := 0.0
	for _, value := range values {
		total += value
	}
	return total / float64(len(values))
}

func parseScore(value string) float64 {
	number, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0
	}
	return number
}

func highestLowestValues(values []float64) (float64, float64) {
	if len(values) == 0 {
		return 0, 0
	}
	highest, lowest := values[0], values[0]
	for _, value := range values {
		if value > highest {
			highest = value
		}
		if value < lowest {
			lowest = value
		}
	}
	return highest, lowest
}

func countBy(values []float64, predicate func(float64) bool) int {
	count := 0
	for _, value := range values {
		if predicate(value) {
			count++
		}
	}
	return count
}

func namesOf(rows []domainscore.ScoreRow) string {
	if len(rows) == 0 {
		return "暂无学生"
	}
	limit := min(len(rows), 6)
	names := []string{}
	for i := 0; i < limit; i++ {
		names = append(names, rows[i].StudentName)
	}
	return joinLabels(names)
}

func weakSubjectsLabel(subjects []string, row domainscore.ScoreRow) string {
	items := []string{}
	for _, subject := range subjects {
		if value := parseScore(row.SubjectScores[subject]); value > 0 && value < 60 {
			items = append(items, subject)
		}
	}
	if len(items) == 0 {
		return "暂无明显低分科"
	}
	return joinLabels(items)
}

func studentRecommendation(weakSubjects string) string {
	if weakSubjects == "" || weakSubjects == "暂无明显低分科" {
		return "保持当前优势，重点观察排名稳定性。"
	}
	return "优先处理薄弱学科：" + weakSubjects + "，建议设置两周可验证目标。"
}

func formatFloat(value float64) string {
	if value == float64(int64(value)) {
		return fmt.Sprintf("%d", int64(value))
	}
	return strconv.FormatFloat(value, 'f', 1, 64)
}

func defaultString(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func joinLabels(values []string) string {
	result := ""
	for index, value := range values {
		if index > 0 {
			result += "、"
		}
		result += value
	}
	return result
}

func sortedKeys(values map[string]bool) []string {
	keys := []string{}
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func clamp(value, minValue, maxValue float64) float64 {
	return minFloat(maxFloat(value, minValue), maxValue)
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
