package score

import (
	"fmt"
	"sync"
	"time"

	"github.com/RaghtNow/opc/apps/edu-insight-api/internal/domain/score"
)

type Service struct {
	mu     sync.RWMutex
	exams  []score.Exam
	detail map[string]score.ExamDetail
}

func NewService() *Service {
	service := &Service{
		detail: map[string]score.ExamDetail{},
	}

	service.seed()
	return service
}

func (s *Service) ListExams() []score.Exam {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]score.Exam{}, s.exams...)
}

func (s *Service) GetExamDetail(examID string) (score.ExamDetail, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	detail, ok := s.detail[examID]
	return detail, ok
}

func (s *Service) ImportExam(req score.ImportRequest) score.ExamDetail {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := fmt.Sprintf("exam-%d", time.Now().UnixNano())
	exam := score.Exam{
		ID:              id,
		Name:            req.Name,
		Type:            req.Type,
		Date:            req.Date,
		SubjectCoverage: req.SubjectCoverage,
		Subjects:        req.Subjects,
		ImportStatus:    "待校验",
	}

	detail := score.ExamDetail{
		Exam:   exam,
		Scores: req.Scores,
		Issues: req.Issues,
	}

	s.exams = append([]score.Exam{exam}, s.exams...)
	s.detail[id] = detail
	return detail
}

func (s *Service) UpdateScore(examID, scoreID string, req score.UpdateScoreRequest) (score.ExamDetail, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	detail, ok := s.detail[examID]
	if !ok {
		return score.ExamDetail{}, false
	}

	for i := range detail.Scores {
		if detail.Scores[i].ID == scoreID {
			detail.Scores[i].Chinese = req.Chinese
			detail.Scores[i].Math = req.Math
			detail.Scores[i].English = req.English
			detail.Scores[i].ElectiveLabel = req.ElectiveLabel
			detail.Scores[i].ElectiveScore = req.ElectiveScore
			detail.Scores[i].Total = req.Total
			s.detail[examID] = detail
			return detail, true
		}
	}

	return score.ExamDetail{}, false
}

func (s *Service) seed() {
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

	s.exams = []score.Exam{exam1, exam2}
	s.detail[exam1.ID] = score.ExamDetail{
		Exam: exam1,
		Scores: []score.ScoreRow{
			{ID: "score-row-1", StudentID: "student-g230301", StudentName: "林书言", Chinese: "121", Math: "128", English: "136", ElectiveLabel: "物化生", ElectiveScore: "278", Total: "663"},
			{ID: "score-row-2", StudentID: "student-g230302", StudentName: "许一诺", Chinese: "116", Math: "122", English: "129", ElectiveLabel: "物化生", ElectiveScore: "266", Total: "633"},
			{ID: "score-row-3", StudentID: "student-g230317", StudentName: "陈可心", Chinese: "113", Math: "104", English: "125", ElectiveLabel: "史地政", ElectiveScore: "249", Total: "591"},
			{ID: "score-row-4", StudentID: "student-g230329", StudentName: "赵博文", Chinese: "108", Math: "96", English: "118", ElectiveLabel: "物化生", ElectiveScore: "241", Total: "563"},
		},
		Issues: []score.ImportIssue{
			{ID: "issue-24", Row: "24", Student: "周子昂 / 生物", Issue: "赋分缺失", Suggestion: "若本次为选考统计，请补充赋分字段", Status: "待处理"},
		},
	}
	s.detail[exam2.ID] = score.ExamDetail{Exam: exam2, Scores: []score.ScoreRow{}, Issues: []score.ImportIssue{}}
}
