package score

import (
	"database/sql"
	"encoding/json"

	"github.com/RaghtNow/opc/apps/edu-insight-api/internal/domain/score"
)

type MySQLRepository struct {
	db *sql.DB
}

func NewMySQLRepository(db *sql.DB) *MySQLRepository {
	return &MySQLRepository{db: db}
}

func (r *MySQLRepository) ListExams(classID string) ([]score.Exam, error) {
	rows, err := r.db.Query(`
		SELECT id, name, type, exam_date, subject_coverage, subjects, import_status
		FROM score_exams
		WHERE class_id = ?
		ORDER BY exam_date DESC, created_at DESC
	`, classID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	exams := []score.Exam{}
	for rows.Next() {
		var exam score.Exam
		var subjectsRaw string
		if err := rows.Scan(&exam.ID, &exam.Name, &exam.Type, &exam.Date, &exam.SubjectCoverage, &subjectsRaw, &exam.ImportStatus); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(subjectsRaw), &exam.Subjects); err != nil {
			exam.Subjects = []string{}
		}
		exams = append(exams, exam)
	}
	return exams, rows.Err()
}

func (r *MySQLRepository) GetExamDetail(classID, examID string) (score.ExamDetail, bool, error) {
	var detail score.ExamDetail
	var subjectsRaw string
	err := r.db.QueryRow(`
		SELECT id, name, type, exam_date, subject_coverage, subjects, import_status
		FROM score_exams
		WHERE class_id = ? AND id = ?
	`, classID, examID).Scan(&detail.Exam.ID, &detail.Exam.Name, &detail.Exam.Type, &detail.Exam.Date, &detail.Exam.SubjectCoverage, &subjectsRaw, &detail.Exam.ImportStatus)
	if err == sql.ErrNoRows {
		return score.ExamDetail{}, false, nil
	}
	if err != nil {
		return score.ExamDetail{}, false, err
	}
	if err := json.Unmarshal([]byte(subjectsRaw), &detail.Exam.Subjects); err != nil {
		detail.Exam.Subjects = []string{}
	}

	rows, err := r.db.Query(`
		SELECT id, student_id, student_name, subject_scores, chinese, math, english, elective_label, elective_score, total
		FROM score_rows
		WHERE exam_id = ?
		ORDER BY student_id
	`, examID)
	if err != nil {
		return score.ExamDetail{}, false, err
	}
	for rows.Next() {
		var row score.ScoreRow
		var subjectScoresRaw sql.NullString
		if err := rows.Scan(&row.ID, &row.StudentID, &row.StudentName, &subjectScoresRaw, &row.Chinese, &row.Math, &row.English, &row.ElectiveLabel, &row.ElectiveScore, &row.Total); err != nil {
			rows.Close()
			return score.ExamDetail{}, false, err
		}
		row.SubjectScores = map[string]string{}
		if subjectScoresRaw.Valid && subjectScoresRaw.String != "" {
			if err := json.Unmarshal([]byte(subjectScoresRaw.String), &row.SubjectScores); err != nil {
				row.SubjectScores = map[string]string{}
			}
		}
		if len(row.SubjectScores) == 0 {
			row.SubjectScores = legacySubjectScores(detail.Exam.Subjects, row)
		}
		detail.Scores = append(detail.Scores, row)
	}
	if err := rows.Close(); err != nil {
		return score.ExamDetail{}, false, err
	}

	issueRows, err := r.db.Query(`
		SELECT id, row_no, student, issue, suggestion, status
		FROM score_import_issues
		WHERE exam_id = ?
		ORDER BY row_no, id
	`, examID)
	if err != nil {
		return score.ExamDetail{}, false, err
	}
	defer issueRows.Close()
	for issueRows.Next() {
		var issue score.ImportIssue
		if err := issueRows.Scan(&issue.ID, &issue.Row, &issue.Student, &issue.Issue, &issue.Suggestion, &issue.Status); err != nil {
			return score.ExamDetail{}, false, err
		}
		detail.Issues = append(detail.Issues, issue)
	}
	return detail, true, issueRows.Err()
}

func (r *MySQLRepository) SaveExamDetail(classID string, detail score.ExamDetail, fileName string) error {
	subjects, err := json.Marshal(detail.Exam.Subjects)
	if err != nil {
		return err
	}
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`
		INSERT INTO score_exams (
			id, class_id, name, type, exam_date, subject_coverage, subjects, import_status, file_name
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			name = VALUES(name),
			type = VALUES(type),
			exam_date = VALUES(exam_date),
			subject_coverage = VALUES(subject_coverage),
			subjects = VALUES(subjects),
			import_status = VALUES(import_status),
			file_name = VALUES(file_name)
	`, detail.Exam.ID, classID, detail.Exam.Name, detail.Exam.Type, detail.Exam.Date, detail.Exam.SubjectCoverage, string(subjects), detail.Exam.ImportStatus, fileName); err != nil {
		return err
	}

	if _, err := tx.Exec("DELETE FROM score_rows WHERE exam_id = ?", detail.Exam.ID); err != nil {
		return err
	}
	for _, row := range detail.Scores {
		subjectScores, err := json.Marshal(row.SubjectScores)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(`
			INSERT INTO score_rows (
				id, exam_id, student_id, student_name, subject_scores, chinese, math, english, elective_label, elective_score, total
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, row.ID, detail.Exam.ID, row.StudentID, row.StudentName, string(subjectScores), row.Chinese, row.Math, row.English, row.ElectiveLabel, row.ElectiveScore, row.Total); err != nil {
			return err
		}
	}

	if _, err := tx.Exec("DELETE FROM score_import_issues WHERE exam_id = ?", detail.Exam.ID); err != nil {
		return err
	}
	for _, issue := range detail.Issues {
		if _, err := tx.Exec(`
			INSERT INTO score_import_issues (
				id, exam_id, row_no, student, issue, suggestion, status
			) VALUES (?, ?, ?, ?, ?, ?, ?)
		`, issue.ID, detail.Exam.ID, issue.Row, issue.Student, issue.Issue, issue.Suggestion, issue.Status); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *MySQLRepository) UpdateScore(examID, scoreID string, row score.ScoreRow) error {
	subjectScores, err := json.Marshal(row.SubjectScores)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(`
		UPDATE score_rows
		SET subject_scores = ?, chinese = ?, math = ?, english = ?, elective_label = ?, elective_score = ?, total = ?
		WHERE exam_id = ? AND id = ?
	`, string(subjectScores), row.Chinese, row.Math, row.English, row.ElectiveLabel, row.ElectiveScore, row.Total, examID, scoreID)
	return err
}

func (r *MySQLRepository) SeedIfEmpty(classID string, details []score.ExamDetail) error {
	var count int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM score_exams WHERE class_id = ?", classID).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	for _, detail := range details {
		if err := r.SaveExamDetail(classID, detail, "seed"); err != nil {
			return err
		}
	}
	return nil
}

func legacySubjectScores(subjects []string, row score.ScoreRow) map[string]string {
	values := map[string]string{}
	for _, subject := range subjects {
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
	return values
}
