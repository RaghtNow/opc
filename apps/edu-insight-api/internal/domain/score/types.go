package score

type Exam struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	Type            string   `json:"type"`
	Date            string   `json:"date"`
	SubjectCoverage string   `json:"subjectCoverage"`
	Subjects        []string `json:"subjects"`
	ImportStatus    string   `json:"importStatus"`
}

type ScoreRow struct {
	ID            string            `json:"id"`
	StudentID     string            `json:"studentId"`
	StudentName   string            `json:"studentName"`
	Chinese       string            `json:"chinese,omitempty"`
	Math          string            `json:"math,omitempty"`
	English       string            `json:"english,omitempty"`
	ElectiveLabel string            `json:"electiveLabel,omitempty"`
	ElectiveScore string            `json:"electiveScore,omitempty"`
	SubjectScores map[string]string `json:"subjectScores"`
	Total         string            `json:"total"`
}

type ImportIssue struct {
	ID         string `json:"id"`
	Row        string `json:"row"`
	Student    string `json:"student"`
	Issue      string `json:"issue"`
	Suggestion string `json:"suggestion"`
	Status     string `json:"status"`
}

type ExamDetail struct {
	Exam   Exam          `json:"exam"`
	Scores []ScoreRow    `json:"scores"`
	Issues []ImportIssue `json:"issues"`
}

type ImportRequest struct {
	Name            string        `json:"name"`
	Type            string        `json:"type"`
	Date            string        `json:"date"`
	Subjects        []string      `json:"subjects"`
	SubjectCoverage string        `json:"subjectCoverage"`
	FileName        string        `json:"fileName"`
	Scores          []ScoreRow    `json:"scores"`
	Issues          []ImportIssue `json:"issues"`
}

type ValidationSummaryItem struct {
	Field  string `json:"field"`
	Result string `json:"result"`
	Note   string `json:"note"`
}

type ValidationMetrics struct {
	Total   int `json:"total"`
	Matched int `json:"matched"`
	Issues  int `json:"issues"`
}

type ImportValidationResult struct {
	OK                bool                    `json:"ok"`
	Rows              []map[string]string     `json:"rows"`
	Headers           []string                `json:"headers"`
	Issues            []ImportIssue           `json:"issues"`
	Metrics           ValidationMetrics       `json:"metrics"`
	ValidationSummary []ValidationSummaryItem `json:"validationSummary"`
	ScoreRows         []ScoreRow              `json:"scoreRows"`
	Error             string                  `json:"error,omitempty"`
}

type UpdateScoreRequest struct {
	Chinese       string            `json:"chinese,omitempty"`
	Math          string            `json:"math,omitempty"`
	English       string            `json:"english,omitempty"`
	ElectiveLabel string            `json:"electiveLabel,omitempty"`
	ElectiveScore string            `json:"electiveScore,omitempty"`
	SubjectScores map[string]string `json:"subjectScores"`
	Total         string            `json:"total"`
}
