package score

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/RaghtNow/opc/apps/edu-insight-api/internal/application/classroom"
	domainclassroom "github.com/RaghtNow/opc/apps/edu-insight-api/internal/domain/classroom"
	"github.com/RaghtNow/opc/apps/edu-insight-api/internal/domain/score"
	"github.com/xuri/excelize/v2"
)

type Service interface {
	ListExams() ([]score.Exam, error)
	GetExamDetail(examID string) (score.ExamDetail, bool, error)
	ValidateImport(fileName string, content []byte, subjects []string, workspace domainclassroom.Workspace) (score.ImportValidationResult, error)
	ImportExam(req score.ImportRequest) (score.ExamDetail, error)
	UpdateScore(examID, scoreID string, req score.UpdateScoreRequest) (score.ExamDetail, bool, error)
}

type Repository interface {
	ListExams(classID string) ([]score.Exam, error)
	GetExamDetail(classID, examID string) (score.ExamDetail, bool, error)
	SaveExamDetail(classID string, detail score.ExamDetail, fileName string) error
	UpdateScore(examID, scoreID string, row score.ScoreRow) error
	SeedIfEmpty(classID string, details []score.ExamDetail) error
}

type PersistentService struct {
	repo    Repository
	classID string
}

func NewPersistentService(repo Repository) (*PersistentService, error) {
	service := &PersistentService{repo: repo, classID: classroom.CurrentClassID}
	if err := repo.SeedIfEmpty(service.classID, defaultExamDetails()); err != nil {
		return nil, err
	}
	return service, nil
}

func (s *PersistentService) ListExams() ([]score.Exam, error) {
	return s.repo.ListExams(s.classID)
}

func (s *PersistentService) GetExamDetail(examID string) (score.ExamDetail, bool, error) {
	return s.repo.GetExamDetail(s.classID, examID)
}

func (s *PersistentService) ValidateImport(fileName string, content []byte, subjects []string, workspace domainclassroom.Workspace) (score.ImportValidationResult, error) {
	return validateImportContent(fileName, content, subjects, workspace)
}

func (s *PersistentService) ImportExam(req score.ImportRequest) (score.ExamDetail, error) {
	id := fmt.Sprintf("exam-%d", time.Now().UnixNano())
	detail := score.ExamDetail{
		Exam: score.Exam{
			ID:              id,
			Name:            req.Name,
			Type:            req.Type,
			Date:            req.Date,
			SubjectCoverage: req.SubjectCoverage,
			Subjects:        req.Subjects,
			ImportStatus:    importStatus(req.Issues),
		},
		Scores: normalizeScoreRows(req.Subjects, req.Scores),
		Issues: normalizeIssueIDs(req.Issues),
	}
	if err := s.repo.SaveExamDetail(s.classID, detail, req.FileName); err != nil {
		return score.ExamDetail{}, err
	}
	return detail, nil
}

func (s *PersistentService) UpdateScore(examID, scoreID string, req score.UpdateScoreRequest) (score.ExamDetail, bool, error) {
	detail, ok, err := s.repo.GetExamDetail(s.classID, examID)
	if err != nil || !ok {
		return score.ExamDetail{}, ok, err
	}
	for _, row := range detail.Scores {
		if row.ID == scoreID {
			row.Chinese = req.Chinese
			row.Math = req.Math
			row.English = req.English
			row.ElectiveLabel = req.ElectiveLabel
			row.ElectiveScore = req.ElectiveScore
			row.SubjectScores = normalizeSubjectScores(detail.Exam.Subjects, score.ScoreRow{
				Chinese:       req.Chinese,
				Math:          req.Math,
				English:       req.English,
				ElectiveLabel: req.ElectiveLabel,
				ElectiveScore: req.ElectiveScore,
				SubjectScores: req.SubjectScores,
			})
			row.Total = normalizeTotal(row.SubjectScores, req.Total)
			if err := s.repo.UpdateScore(examID, scoreID, row); err != nil {
				return score.ExamDetail{}, false, err
			}
			return s.repo.GetExamDetail(s.classID, examID)
		}
	}
	return score.ExamDetail{}, false, nil
}

type MemoryService struct {
	mu     sync.RWMutex
	exams  []score.Exam
	detail map[string]score.ExamDetail
}

func NewMemoryService() *MemoryService {
	service := &MemoryService{detail: map[string]score.ExamDetail{}}
	for _, detail := range defaultExamDetails() {
		service.exams = append(service.exams, detail.Exam)
		service.detail[detail.Exam.ID] = detail
	}
	return service
}

func (s *MemoryService) ListExams() ([]score.Exam, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]score.Exam{}, s.exams...), nil
}

func (s *MemoryService) GetExamDetail(examID string) (score.ExamDetail, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	detail, ok := s.detail[examID]
	return detail, ok, nil
}

func (s *MemoryService) ValidateImport(fileName string, content []byte, subjects []string, workspace domainclassroom.Workspace) (score.ImportValidationResult, error) {
	return validateImportContent(fileName, content, subjects, workspace)
}

func (s *MemoryService) ImportExam(req score.ImportRequest) (score.ExamDetail, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := fmt.Sprintf("exam-%d", time.Now().UnixNano())
	detail := score.ExamDetail{
		Exam: score.Exam{
			ID:              id,
			Name:            req.Name,
			Type:            req.Type,
			Date:            req.Date,
			SubjectCoverage: req.SubjectCoverage,
			Subjects:        req.Subjects,
			ImportStatus:    importStatus(req.Issues),
		},
		Scores: normalizeScoreRows(req.Subjects, req.Scores),
		Issues: normalizeIssueIDs(req.Issues),
	}

	s.exams = append([]score.Exam{detail.Exam}, s.exams...)
	s.detail[id] = detail
	return detail, nil
}

func (s *MemoryService) UpdateScore(examID, scoreID string, req score.UpdateScoreRequest) (score.ExamDetail, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	detail, ok := s.detail[examID]
	if !ok {
		return score.ExamDetail{}, false, nil
	}
	for i := range detail.Scores {
		if detail.Scores[i].ID == scoreID {
			detail.Scores[i].Chinese = req.Chinese
			detail.Scores[i].Math = req.Math
			detail.Scores[i].English = req.English
			detail.Scores[i].ElectiveLabel = req.ElectiveLabel
			detail.Scores[i].ElectiveScore = req.ElectiveScore
			detail.Scores[i].SubjectScores = normalizeSubjectScores(detail.Exam.Subjects, score.ScoreRow{
				Chinese:       req.Chinese,
				Math:          req.Math,
				English:       req.English,
				ElectiveLabel: req.ElectiveLabel,
				ElectiveScore: req.ElectiveScore,
				SubjectScores: req.SubjectScores,
			})
			detail.Scores[i].Total = normalizeTotal(detail.Scores[i].SubjectScores, req.Total)
			s.detail[examID] = detail
			return detail, true, nil
		}
	}
	return score.ExamDetail{}, false, nil
}

func importStatus(issues []score.ImportIssue) string {
	if len(issues) > 0 {
		return "待校验"
	}
	return "已完成"
}

func normalizeScoreRows(subjects []string, rows []score.ScoreRow) []score.ScoreRow {
	normalized := append([]score.ScoreRow{}, rows...)
	for i := range normalized {
		if normalized[i].ID == "" {
			normalized[i].ID = fmt.Sprintf("score-row-%d-%d", time.Now().UnixNano(), i)
		}
		normalized[i].SubjectScores = normalizeSubjectScores(subjects, normalized[i])
		normalized[i].Total = normalizeTotal(normalized[i].SubjectScores, normalized[i].Total)
	}
	return normalized
}

func normalizeSubjectScores(subjects []string, row score.ScoreRow) map[string]string {
	values := map[string]string{}
	for _, subject := range subjects {
		if row.SubjectScores != nil {
			if value, ok := row.SubjectScores[subject]; ok {
				values[subject] = value
				continue
			}
		}
		switch subject {
		case "语文":
			values[subject] = row.Chinese
		case "数学":
			values[subject] = row.Math
		case "英语":
			values[subject] = row.English
		default:
			values[subject] = ""
		}
	}
	if len(values) == 0 && row.SubjectScores != nil {
		for subject, value := range row.SubjectScores {
			values[subject] = value
		}
	}
	return values
}

func normalizeTotal(subjectScores map[string]string, fallback string) string {
	var total float64
	hasNumeric := false
	for _, raw := range subjectScores {
		if raw == "" || raw == "-" {
			continue
		}
		value, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			continue
		}
		total += value
		hasNumeric = true
	}
	if !hasNumeric {
		return fallback
	}
	if total == float64(int64(total)) {
		return fmt.Sprintf("%d", int64(total))
	}
	return strconv.FormatFloat(total, 'f', 1, 64)
}

func normalizeIssueIDs(issues []score.ImportIssue) []score.ImportIssue {
	normalized := append([]score.ImportIssue{}, issues...)
	for i := range normalized {
		if normalized[i].ID == "" {
			normalized[i].ID = fmt.Sprintf("issue-%d-%d", time.Now().UnixNano(), i)
		}
	}
	return normalized
}

func validateImportContent(fileName string, content []byte, subjects []string, workspace domainclassroom.Workspace) (score.ImportValidationResult, error) {
	rows, headers, err := parseImportRows(fileName, content)
	if err != nil {
		return score.ImportValidationResult{
			OK:                false,
			Rows:              []map[string]string{},
			Headers:           []string{},
			Issues:            []score.ImportIssue{},
			Metrics:           score.ValidationMetrics{},
			ValidationSummary: []score.ValidationSummaryItem{},
			ScoreRows:         []score.ScoreRow{},
			Error:             err.Error(),
		}, nil
	}
	if len(rows) == 0 {
		return score.ImportValidationResult{
			OK:                false,
			Rows:              []map[string]string{},
			Headers:           headers,
			Issues:            []score.ImportIssue{},
			Metrics:           score.ValidationMetrics{},
			ValidationSummary: []score.ValidationSummaryItem{},
			ScoreRows:         []score.ScoreRow{},
			Error:             "文件内容为空、格式无效，或当前格式暂不支持",
		}, nil
	}

	requiredBase := []string{"学号", "姓名"}
	missingBase := missingHeaders(headers, requiredBase)
	missingSubjects := missingHeaders(headers, subjects)
	knownStudentIDs := map[string]bool{}
	knownStudentNames := map[string]bool{}
	for _, student := range workspace.Students {
		knownStudentIDs[student.StudentNo] = true
		knownStudentNames[student.Name] = true
	}

	issues := []score.ImportIssue{}
	scoreRows := []score.ScoreRow{}
	for index, row := range rows {
		studentNo := row["学号"]
		studentName := row["姓名"]
		if studentName == "" {
			studentName = fmt.Sprintf("第%d行学生", index+2)
		}
		if !knownStudentIDs[studentNo] && !knownStudentNames[studentName] {
			issues = append(issues, score.ImportIssue{
				ID:         fmt.Sprintf("issue-student-%d", index),
				Row:        fmt.Sprintf("%d", index+2),
				Student:    studentName,
				Issue:      "学生未匹配到班级档案",
				Suggestion: "检查学号或姓名是否与班级档案一致",
				Status:     "待处理",
			})
		}

		subjectScores := map[string]string{}
		for _, subject := range subjects {
			subjectScores[subject] = row[subject]
		}
		scoreRows = append(scoreRows, score.ScoreRow{
			ID:            fmt.Sprintf("score-%s-%d", safeID(studentNo), index),
			StudentID:     studentNo,
			StudentName:   studentName,
			Chinese:       subjectScores["语文"],
			Math:          subjectScores["数学"],
			English:       subjectScores["英语"],
			ElectiveLabel: strings.Join(nonCoreSubjects(subjects), ""),
			SubjectScores: subjectScores,
			Total:         normalizeTotal(subjectScores, ""),
		})
	}

	validationSummary := []score.ValidationSummaryItem{
		{
			Field:  "基础字段",
			Result: resultLabel(len(missingBase) == 0),
			Note:   summaryNote(missingBase, "学号、姓名字段齐全"),
		},
		{
			Field:  "学科字段",
			Result: resultLabel(len(missingSubjects) == 0),
			Note:   summaryNote(missingSubjects, "已勾选学科全部存在"),
		},
		{
			Field:  "学生匹配",
			Result: fmt.Sprintf("%d / %d 通过", len(rows)-len(issues), len(rows)),
			Note: func() string {
				if len(issues) == 0 {
					return "全部学生已成功匹配"
				}
				return fmt.Sprintf("%d 条学生档案未匹配", len(issues))
			}(),
		},
	}

	return score.ImportValidationResult{
		OK:                len(missingBase) == 0 && len(missingSubjects) == 0,
		Rows:              rows,
		Headers:           headers,
		Issues:            issues,
		Metrics:           score.ValidationMetrics{Total: len(rows), Matched: len(rows) - len(issues), Issues: len(issues)},
		ValidationSummary: validationSummary,
		ScoreRows:         scoreRows,
	}, nil
}

func parseImportRows(fileName string, content []byte) ([]map[string]string, []string, error) {
	lowerName := strings.ToLower(fileName)
	switch {
	case strings.HasSuffix(lowerName, ".csv"):
		reader := csv.NewReader(bytes.NewReader(content))
		reader.TrimLeadingSpace = true
		records, err := reader.ReadAll()
		if err != nil {
			return nil, nil, fmt.Errorf("CSV 解析失败：%w", err)
		}
		return recordsToRows(records), headersFromRecords(records), nil
	case strings.HasSuffix(lowerName, ".xlsx") || strings.HasSuffix(lowerName, ".xls"):
		workbook, err := excelize.OpenReader(bytes.NewReader(content))
		if err != nil {
			return nil, nil, fmt.Errorf("Excel 解析失败：%w", err)
		}
		defer workbook.Close()
		sheets := workbook.GetSheetList()
		if len(sheets) == 0 {
			return nil, nil, nil
		}
		records, err := workbook.GetRows(sheets[0])
		if err != nil && err != io.EOF {
			return nil, nil, fmt.Errorf("Excel 读取失败：%w", err)
		}
		return recordsToRows(records), headersFromRecords(records), nil
	default:
		return nil, nil, fmt.Errorf("文件内容为空、格式无效，或当前格式暂不支持")
	}
}

func headersFromRecords(records [][]string) []string {
	if len(records) == 0 {
		return []string{}
	}
	headers := make([]string, 0, len(records[0]))
	for _, header := range records[0] {
		headers = append(headers, strings.TrimSpace(header))
	}
	return headers
}

func recordsToRows(records [][]string) []map[string]string {
	headers := headersFromRecords(records)
	rows := []map[string]string{}
	for _, record := range records[1:] {
		empty := true
		row := map[string]string{}
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
	return rows
}

func missingHeaders(headers []string, required []string) []string {
	headerSet := map[string]bool{}
	for _, header := range headers {
		headerSet[header] = true
	}
	missing := []string{}
	for _, item := range required {
		if !headerSet[item] {
			missing = append(missing, item)
		}
	}
	return missing
}

func resultLabel(ok bool) string {
	if ok {
		return "通过"
	}
	return "缺失"
}

func summaryNote(missing []string, okNote string) string {
	if len(missing) == 0 {
		return okNote
	}
	return "缺少：" + strings.Join(missing, "、")
}

func nonCoreSubjects(subjects []string) []string {
	items := []string{}
	for _, subject := range subjects {
		if subject != "语文" && subject != "数学" && subject != "英语" {
			items = append(items, subject)
		}
	}
	return items
}

func safeID(value string) string {
	if value == "" {
		return "unknown"
	}
	return strings.NewReplacer(" ", "-", "/", "-", "\\", "-").Replace(value)
}

func defaultExamDetails() []score.ExamDetail {
	exam1 := score.Exam{
		ID:              "exam-june-monthly",
		Name:            "2026 年 6 月月考",
		Type:            "月考",
		Date:            "2026-06-08",
		SubjectCoverage: "语文 / 数学 / 英语 / 物理 / 化学 / 生物",
		Subjects:        []string{"语文", "数学", "英语", "物理", "化学", "生物"},
		ImportStatus:    "已完成",
	}
	exam2 := score.Exam{
		ID:              "exam-midterm",
		Name:            "2026 年期中考试",
		Type:            "期中",
		Date:            "2026-05-12",
		SubjectCoverage: "全科",
		Subjects:        []string{"语文", "数学", "英语", "物理", "化学", "生物", "历史", "地理", "政治"},
		ImportStatus:    "已完成",
	}
	return []score.ExamDetail{
		{
			Exam: exam1,
			Scores: []score.ScoreRow{
				{ID: "score-row-1", StudentID: "student-g230301", StudentName: "林书言", Chinese: "121", Math: "128", English: "136", ElectiveLabel: "物化生", ElectiveScore: "278", SubjectScores: map[string]string{"语文": "121", "数学": "128", "英语": "136", "物理": "94", "化学": "91", "生物": "93"}, Total: "663"},
				{ID: "score-row-2", StudentID: "student-g230302", StudentName: "许一诺", Chinese: "116", Math: "122", English: "129", ElectiveLabel: "物化生", ElectiveScore: "266", SubjectScores: map[string]string{"语文": "116", "数学": "122", "英语": "129", "物理": "89", "化学": "88", "生物": "89"}, Total: "633"},
				{ID: "score-row-3", StudentID: "student-g230317", StudentName: "陈可心", Chinese: "113", Math: "104", English: "125", ElectiveLabel: "史地政", ElectiveScore: "249", SubjectScores: map[string]string{"语文": "113", "数学": "104", "英语": "125", "历史": "84", "地理": "82", "政治": "83"}, Total: "591"},
				{ID: "score-row-4", StudentID: "student-g230329", StudentName: "赵博文", Chinese: "108", Math: "96", English: "118", ElectiveLabel: "物化生", ElectiveScore: "241", SubjectScores: map[string]string{"语文": "108", "数学": "96", "英语": "118", "物理": "82", "化学": "80", "生物": "79"}, Total: "563"},
			},
			Issues: []score.ImportIssue{
				{ID: "issue-24", Row: "24", Student: "周子昂 / 生物", Issue: "赋分缺失", Suggestion: "若本次为选考统计，请补充赋分字段", Status: "待处理"},
			},
		},
		{Exam: exam2, Scores: []score.ScoreRow{}, Issues: []score.ImportIssue{}},
	}
}
