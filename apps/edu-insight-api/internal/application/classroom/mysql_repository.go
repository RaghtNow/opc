package classroom

import (
	"database/sql"
	"encoding/json"

	"github.com/RaghtNow/opc/apps/edu-insight-api/internal/domain/classroom"
)

type MySQLRepository struct {
	db *sql.DB
}

func NewMySQLRepository(db *sql.DB) *MySQLRepository {
	return &MySQLRepository{db: db}
}

func (r *MySQLRepository) GetWorkspace(classID string) (classroom.Workspace, error) {
	var workspace classroom.Workspace
	var schoolName, gradeName, homeroomTeacher, academicYear, classStatus string
	err := r.db.QueryRow(`
		SELECT id, school_name, grade_name, class_name, homeroom_teacher, academic_year, class_status,
		       stage_id, stage_label, stage_description
		FROM classroom_classes
		WHERE id = ?
	`, classID).Scan(
		&workspace.ClassID,
		&schoolName,
		&gradeName,
		&workspace.ClassName,
		&homeroomTeacher,
		&academicYear,
		&classStatus,
		&workspace.Stage.ID,
		&workspace.Stage.Label,
		&workspace.Stage.Description,
	)
	if err != nil {
		return classroom.Workspace{}, err
	}
	workspace.BaseFields = []classroom.BaseField{
		{Label: "学校", Value: schoolName},
		{Label: "年级", Value: gradeName},
		{Label: "行政班", Value: workspace.ClassName},
		{Label: "班主任", Value: homeroomTeacher},
		{Label: "学年", Value: academicYear},
		{Label: "班级状态", Value: classStatus},
	}

	students, err := r.listStudents(classID)
	if err != nil {
		return classroom.Workspace{}, err
	}
	teachers, err := r.listTeachers(classID)
	if err != nil {
		return classroom.Workspace{}, err
	}
	policies, err := r.listPolicies(classID)
	if err != nil {
		return classroom.Workspace{}, err
	}
	workspace.Students = students
	workspace.Teachers = teachers
	workspace.Policies = policies
	return withRosterInsights(workspace), nil
}

func (r *MySQLRepository) ListClassOptions() ([]classroom.ClassOption, error) {
	rows, err := r.db.Query(`
		SELECT id, grade_name, class_name, academic_year
		FROM classroom_classes
		WHERE class_status <> '已归档'
		ORDER BY academic_year DESC, grade_name, class_name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	options := []classroom.ClassOption{}
	for rows.Next() {
		var option classroom.ClassOption
		if err := rows.Scan(&option.ID, &option.GradeName, &option.ClassName, &option.AcademicYear); err != nil {
			return nil, err
		}
		option.RoleLabel = "班主任"
		option.PrimaryLabel = option.ClassName
		option.SecondaryLabel = option.AcademicYear
		options = append(options, option)
	}
	return options, rows.Err()
}

func (r *MySQLRepository) SaveStudent(classID string, student classroom.Student) error {
	subjects, err := json.Marshal(student.ElectiveSubjects)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(`
		INSERT INTO classroom_students (
			id, class_id, student_no, name, gender, combination, elective_subjects,
			parent_mobile, status, parent_status, selection_status, profile_status
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			student_no = VALUES(student_no),
			name = VALUES(name),
			gender = VALUES(gender),
			combination = VALUES(combination),
			elective_subjects = VALUES(elective_subjects),
			parent_mobile = VALUES(parent_mobile),
			status = VALUES(status),
			parent_status = VALUES(parent_status),
			selection_status = VALUES(selection_status),
			profile_status = VALUES(profile_status)
	`, student.ID, classID, student.StudentNo, student.Name, student.Gender, student.Combination, string(subjects), student.ParentMobile, student.Status, student.ParentStatus, student.SelectionStatus, student.ProfileStatus)
	return err
}

func (r *MySQLRepository) SaveClass(workspace classroom.Workspace) error {
	fields := map[string]string{}
	for _, field := range workspace.BaseFields {
		fields[field.Label] = field.Value
	}

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`
		INSERT INTO classroom_classes (
			id, school_name, grade_name, class_name, homeroom_teacher, academic_year, class_status,
			stage_id, stage_label, stage_description
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			school_name = VALUES(school_name),
			grade_name = VALUES(grade_name),
			class_name = VALUES(class_name),
			homeroom_teacher = VALUES(homeroom_teacher),
			academic_year = VALUES(academic_year),
			class_status = VALUES(class_status),
			stage_id = VALUES(stage_id),
			stage_label = VALUES(stage_label),
			stage_description = VALUES(stage_description)
	`, workspace.ClassID, fields["学校"], fields["年级"], workspace.ClassName, fields["班主任"], fields["学年"], fields["班级状态"], workspace.Stage.ID, workspace.Stage.Label, workspace.Stage.Description); err != nil {
		return err
	}

	for _, policy := range workspace.Policies {
		if _, err := tx.Exec(`
			INSERT INTO classroom_policies (id, class_id, title, value, note)
			VALUES (?, ?, ?, ?, ?)
			ON DUPLICATE KEY UPDATE
				title = VALUES(title),
				value = VALUES(value),
				note = VALUES(note)
		`, policy.ID, workspace.ClassID, policy.Title, policy.Value, policy.Note); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *MySQLRepository) ArchiveStudent(classID, studentID string) error {
	_, err := r.db.Exec(`
		UPDATE classroom_students
		SET status = '已删除', parent_status = '已停用', selection_status = '已停用', profile_status = 'archived'
		WHERE class_id = ? AND id = ?
	`, classID, studentID)
	return err
}

func (r *MySQLRepository) SaveTeacher(classID string, teacher classroom.TeacherAssignment) error {
	_, err := r.db.Exec(`
		INSERT INTO classroom_teachers (
			id, class_id, subject, teacher, mobile, classes, sync_status, account_status,
			account_id, account_bound_at, permission_status, permission_synced_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			subject = VALUES(subject),
			teacher = VALUES(teacher),
			mobile = VALUES(mobile),
			classes = VALUES(classes),
			sync_status = VALUES(sync_status),
			account_status = VALUES(account_status),
			account_id = VALUES(account_id),
			account_bound_at = VALUES(account_bound_at),
			permission_status = VALUES(permission_status),
			permission_synced_at = VALUES(permission_synced_at)
	`, teacher.ID, classID, teacher.Subject, teacher.Teacher, teacher.Mobile, teacher.Classes, teacher.SyncStatus, teacher.AccountStatus, teacher.AccountID, teacher.AccountBoundAt, teacher.PermissionStatus, teacher.PermissionSyncedAt)
	return err
}

func (r *MySQLRepository) DeleteTeacher(classID, teacherID string) error {
	_, err := r.db.Exec(`
		DELETE FROM classroom_teachers
		WHERE class_id = ? AND id = ?
	`, classID, teacherID)
	return err
}

func (r *MySQLRepository) SavePolicy(classID string, policy classroom.Policy) error {
	_, err := r.db.Exec(`
		INSERT INTO classroom_policies (id, class_id, title, value, note)
		VALUES (?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			title = VALUES(title),
			value = VALUES(value),
			note = VALUES(note)
	`, policy.ID, classID, policy.Title, policy.Value, policy.Note)
	return err
}

func (r *MySQLRepository) SeedIfEmpty(workspace classroom.Workspace) error {
	var count int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM classroom_classes WHERE id = ?", workspace.ClassID).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	if err := r.SaveClass(workspace); err != nil {
		return err
	}

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, student := range workspace.Students {
		subjects, err := json.Marshal(student.ElectiveSubjects)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(`
			INSERT INTO classroom_students (
				id, class_id, student_no, name, gender, combination, elective_subjects,
				parent_mobile, status, parent_status, selection_status, profile_status
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, student.ID, workspace.ClassID, student.StudentNo, student.Name, student.Gender, student.Combination, string(subjects), student.ParentMobile, student.Status, student.ParentStatus, student.SelectionStatus, student.ProfileStatus); err != nil {
			return err
		}
	}

	for _, teacher := range workspace.Teachers {
		if _, err := tx.Exec(`
			INSERT INTO classroom_teachers (
				id, class_id, subject, teacher, mobile, classes, sync_status, account_status,
				account_id, account_bound_at, permission_status, permission_synced_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, teacher.ID, workspace.ClassID, teacher.Subject, teacher.Teacher, teacher.Mobile, teacher.Classes, teacher.SyncStatus, teacher.AccountStatus, teacher.AccountID, teacher.AccountBoundAt, teacher.PermissionStatus, teacher.PermissionSyncedAt); err != nil {
			return err
		}
	}

	for _, policy := range workspace.Policies {
		if _, err := tx.Exec(`
			INSERT INTO classroom_policies (id, class_id, title, value, note)
			VALUES (?, ?, ?, ?, ?)
			ON DUPLICATE KEY UPDATE
				title = VALUES(title),
				value = VALUES(value),
				note = VALUES(note)
		`, policy.ID, workspace.ClassID, policy.Title, policy.Value, policy.Note); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *MySQLRepository) listStudents(classID string) ([]classroom.Student, error) {
	rows, err := r.db.Query(`
		SELECT id, student_no, name, gender, combination, elective_subjects,
		       parent_mobile, status, parent_status, selection_status, profile_status
		FROM classroom_students
		WHERE class_id = ?
		  AND profile_status <> 'archived'
		ORDER BY student_no
	`, classID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	students := []classroom.Student{}
	for rows.Next() {
		var student classroom.Student
		var subjectsRaw string
		if err := rows.Scan(&student.ID, &student.StudentNo, &student.Name, &student.Gender, &student.Combination, &subjectsRaw, &student.ParentMobile, &student.Status, &student.ParentStatus, &student.SelectionStatus, &student.ProfileStatus); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(subjectsRaw), &student.ElectiveSubjects); err != nil {
			student.ElectiveSubjects = []string{}
		}
		students = append(students, student)
	}
	return students, rows.Err()
}

func (r *MySQLRepository) listTeachers(classID string) ([]classroom.TeacherAssignment, error) {
	rows, err := r.db.Query(`
		SELECT id, subject, teacher, mobile, classes, sync_status, account_status,
		       account_id, account_bound_at, permission_status, permission_synced_at
		FROM classroom_teachers
		WHERE class_id = ?
		ORDER BY subject, teacher
	`, classID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	teachers := []classroom.TeacherAssignment{}
	for rows.Next() {
		var teacher classroom.TeacherAssignment
		if err := rows.Scan(&teacher.ID, &teacher.Subject, &teacher.Teacher, &teacher.Mobile, &teacher.Classes, &teacher.SyncStatus, &teacher.AccountStatus, &teacher.AccountID, &teacher.AccountBoundAt, &teacher.PermissionStatus, &teacher.PermissionSyncedAt); err != nil {
			return nil, err
		}
		teachers = append(teachers, normalizeTeacherState(teacher))
	}
	return teachers, rows.Err()
}

func (r *MySQLRepository) listPolicies(classID string) ([]classroom.Policy, error) {
	rows, err := r.db.Query(`
		SELECT id, title, value, note
		FROM classroom_policies
		WHERE class_id = ?
		ORDER BY id
	`, classID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	policies := []classroom.Policy{}
	for rows.Next() {
		var policy classroom.Policy
		if err := rows.Scan(&policy.ID, &policy.Title, &policy.Value, &policy.Note); err != nil {
			return nil, err
		}
		policies = append(policies, policy)
	}
	return policies, rows.Err()
}
