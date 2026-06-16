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
	ID            string `json:"id"`
	StudentID     string `json:"studentId"`
	StudentName   string `json:"studentName"`
	Chinese       string `json:"chinese"`
	Math          string `json:"math"`
	English       string `json:"english"`
	ElectiveLabel string `json:"electiveLabel"`
	ElectiveScore string `json:"electiveScore"`
	Total         string `json:"total"`
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
	Exam   Exam         `json:"exam"`
	Scores []ScoreRow   `json:"scores"`
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

type UpdateScoreRequest struct {
	Chinese       string `json:"chinese"`
	Math          string `json:"math"`
	English       string `json:"english"`
	ElectiveLabel string `json:"electiveLabel"`
	ElectiveScore string `json:"electiveScore"`
	Total         string `json:"total"`
}
