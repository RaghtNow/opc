package auth

import (
	"database/sql"
	"time"

	domainauth "github.com/RaghtNow/opc/apps/edu-insight-api/internal/domain/auth"
)

type MySQLRepository struct {
	db *sql.DB
}

func NewMySQLRepository(db *sql.DB) *MySQLRepository {
	return &MySQLRepository{db: db}
}

func (r *MySQLRepository) SaveSMSCode(record SMSCodeRecord) error {
	_, err := r.db.Exec(`
		INSERT INTO auth_sms_codes (id, mobile, scene, code_hash, expires_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, record.ID, record.Mobile, record.Scene, record.CodeHash, record.ExpiresAt, record.CreatedAt)
	return err
}

func (r *MySQLRepository) LatestValidSMSCode(mobile, scene string, now time.Time) (SMSCodeRecord, bool, error) {
	var record SMSCodeRecord
	err := r.db.QueryRow(`
		SELECT id, mobile, scene, code_hash, expires_at, used_at, created_at
		FROM auth_sms_codes
		WHERE mobile = ? AND scene = ? AND used_at IS NULL AND expires_at > ?
		ORDER BY created_at DESC
		LIMIT 1
	`, mobile, scene, now).Scan(&record.ID, &record.Mobile, &record.Scene, &record.CodeHash, &record.ExpiresAt, &record.UsedAt, &record.CreatedAt)
	if err == sql.ErrNoRows {
		return SMSCodeRecord{}, false, nil
	}
	return record, err == nil, err
}

func (r *MySQLRepository) MarkSMSCodeUsed(id string, usedAt time.Time) error {
	_, err := r.db.Exec("UPDATE auth_sms_codes SET used_at = ? WHERE id = ?", usedAt, id)
	return err
}

func (r *MySQLRepository) FindUserByMobile(mobile string) (domainauth.User, bool, error) {
	var user domainauth.User
	err := r.db.QueryRow(`
		SELECT u.id, u.name, COALESCE(ui.identifier, ''), u.status, DATE_FORMAT(u.created_at, '%Y-%m-%d %H:%i:%s')
		FROM auth_users u
		JOIN auth_user_identities ui ON ui.user_id = u.id AND ui.identity_type = 'mobile'
		WHERE ui.identifier = ?
		LIMIT 1
	`, mobile).Scan(&user.ID, &user.Name, &user.Mobile, &user.Status, &user.CreatedAt)
	if err == sql.ErrNoRows {
		return domainauth.User{}, false, nil
	}
	return user, err == nil, err
}

func (r *MySQLRepository) FindUserByID(userID string) (domainauth.User, bool, error) {
	var user domainauth.User
	err := r.db.QueryRow(`
		SELECT u.id, u.name, COALESCE(ui.identifier, ''), u.status, DATE_FORMAT(u.created_at, '%Y-%m-%d %H:%i:%s')
		FROM auth_users u
		LEFT JOIN auth_user_identities ui ON ui.user_id = u.id AND ui.identity_type = 'mobile'
		WHERE u.id = ?
		LIMIT 1
	`, userID).Scan(&user.ID, &user.Name, &user.Mobile, &user.Status, &user.CreatedAt)
	if err == sql.ErrNoRows {
		return domainauth.User{}, false, nil
	}
	return user, err == nil, err
}

func (r *MySQLRepository) CreateUserWithMobile(user domainauth.User, mobile string, verifiedAt time.Time) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`
		INSERT INTO auth_users (id, name, status, created_at)
		VALUES (?, ?, ?, ?)
	`, user.ID, user.Name, user.Status, verifiedAt); err != nil {
		return err
	}
	if _, err := tx.Exec(`
		INSERT INTO auth_user_identities (id, user_id, identity_type, identifier, verified, last_verified_at)
		VALUES (?, ?, 'mobile', ?, 1, ?)
	`, newID("identity"), user.ID, mobile, verifiedAt); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *MySQLRepository) SaveSession(session SessionRecord) error {
	_, err := r.db.Exec(`
		INSERT INTO auth_sessions (id, user_id, token_hash, expires_at, created_at)
		VALUES (?, ?, ?, ?, ?)
	`, session.ID, session.UserID, session.TokenHash, session.ExpiresAt, session.CreatedAt)
	return err
}

func (r *MySQLRepository) FindSession(tokenHash string, now time.Time) (SessionRecord, bool, error) {
	var session SessionRecord
	err := r.db.QueryRow(`
		SELECT id, user_id, token_hash, expires_at, created_at
		FROM auth_sessions
		WHERE token_hash = ? AND expires_at > ? AND revoked_at IS NULL
		LIMIT 1
	`, tokenHash, now).Scan(&session.ID, &session.UserID, &session.TokenHash, &session.ExpiresAt, &session.CreatedAt)
	if err == sql.ErrNoRows {
		return SessionRecord{}, false, nil
	}
	return session, err == nil, err
}

func (r *MySQLRepository) ListWorkIdentities(userID, mobile string) ([]domainauth.WorkIdentity, error) {
	rows, err := r.db.Query(`
		SELECT id, role_type, role_label, primary_label, secondary_label, scope_type, scope_id, subject
		FROM auth_work_identities
		WHERE user_id = ? AND status = 'active'
		ORDER BY sort_order, id
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	identities := []domainauth.WorkIdentity{}
	for rows.Next() {
		var identity domainauth.WorkIdentity
		if err := rows.Scan(&identity.ID, &identity.RoleType, &identity.RoleLabel, &identity.PrimaryLabel, &identity.SecondaryLabel, &identity.ScopeType, &identity.ScopeID, &identity.Subject); err != nil {
			return nil, err
		}
		identities = append(identities, identity)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(identities) > 0 {
		return identities, nil
	}
	return fallbackWorkIdentities(mobile), nil
}

func fallbackWorkIdentities(mobile string) []domainauth.WorkIdentity {
	if mobile == "13800001003" {
		return []domainauth.WorkIdentity{
			{ID: "homeroom", RoleType: "homeroom_teacher", RoleLabel: "班主任", PrimaryLabel: "班级管理", SecondaryLabel: "高二（3）班 / 高一（8）班", ScopeType: "admin_class_group", ScopeID: "homeroom-li"},
			{ID: "subject-english", RoleType: "subject_teacher", RoleLabel: "任课老师", PrimaryLabel: "英语任课分析", SecondaryLabel: "高二（3）班 / 高一（8）班", ScopeType: "teaching_scope", ScopeID: "teacher-english-li", Subject: "英语"},
		}
	}
	if mobile == "13800001002" {
		return []domainauth.WorkIdentity{
			{ID: "subject-math", RoleType: "subject_teacher", RoleLabel: "任课老师", PrimaryLabel: "数学任课分析", SecondaryLabel: "高二（3）班 / 高一（8）班", ScopeType: "teaching_scope", ScopeID: "teacher-math-wang", Subject: "数学"},
		}
	}
	return []domainauth.WorkIdentity{
		{ID: "trial-homeroom", RoleType: "homeroom_teacher", RoleLabel: "班主任", PrimaryLabel: "个人试用班级", SecondaryLabel: "可创建班级并导入成绩", ScopeType: "personal_workspace", ScopeID: "trial"},
	}
}
