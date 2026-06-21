package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	domainauth "github.com/RaghtNow/opc/apps/edu-insight-api/internal/domain/auth"
)

const (
	defaultScene = "login"
	codeTTL      = 5 * time.Minute
	sessionTTL   = 24 * time.Hour
)

type Service interface {
	SendSMSCode(req domainauth.SendSMSCodeRequest) (domainauth.SendSMSCodeResponse, error)
	LoginWithSMS(req domainauth.LoginWithSMSRequest) (domainauth.LoginResponse, error)
	CurrentUser(token string) (domainauth.CurrentUser, bool, error)
}

type Repository interface {
	SaveSMSCode(record SMSCodeRecord) error
	LatestValidSMSCode(mobile, scene string, now time.Time) (SMSCodeRecord, bool, error)
	MarkSMSCodeUsed(id string, usedAt time.Time) error
	FindUserByMobile(mobile string) (domainauth.User, bool, error)
	CreateUserWithMobile(user domainauth.User, mobile string, verifiedAt time.Time) error
	SaveSession(session SessionRecord) error
	FindSession(tokenHash string, now time.Time) (SessionRecord, bool, error)
	ListWorkIdentities(userID, mobile string) ([]domainauth.WorkIdentity, error)
}

type SMSCodeRecord struct {
	ID        string
	Mobile    string
	Scene     string
	CodeHash  string
	ExpiresAt time.Time
	UsedAt    sql.NullTime
	CreatedAt time.Time
}

type SessionRecord struct {
	ID        string
	UserID    string
	TokenHash string
	ExpiresAt time.Time
	CreatedAt time.Time
}

type AuthService struct {
	repo    Repository
	appEnv  string
	nowFunc func() time.Time
}

func NewService(repo Repository, appEnv string) *AuthService {
	return &AuthService{repo: repo, appEnv: appEnv, nowFunc: time.Now}
}

func (s *AuthService) SendSMSCode(req domainauth.SendSMSCodeRequest) (domainauth.SendSMSCodeResponse, error) {
	mobile, err := normalizeMobile(req.Mobile)
	if err != nil {
		return domainauth.SendSMSCodeResponse{}, err
	}
	scene := normalizeScene(req.Scene)
	code, err := randomDigits(6)
	if err != nil {
		return domainauth.SendSMSCodeResponse{}, err
	}
	now := s.nowFunc()
	record := SMSCodeRecord{
		ID:        newID("sms"),
		Mobile:    mobile,
		Scene:     scene,
		CodeHash:  hashCode(mobile, scene, code),
		ExpiresAt: now.Add(codeTTL),
		CreatedAt: now,
	}
	if err := s.repo.SaveSMSCode(record); err != nil {
		return domainauth.SendSMSCodeResponse{}, err
	}

	resp := domainauth.SendSMSCodeResponse{
		Mobile:    mobile,
		Scene:     scene,
		ExpiresIn: int(codeTTL.Seconds()),
		Message:   "验证码已发送",
	}
	if s.appEnv != "production" {
		resp.DevCode = code
		resp.Message = "开发环境验证码已生成"
	}
	return resp, nil
}

func (s *AuthService) LoginWithSMS(req domainauth.LoginWithSMSRequest) (domainauth.LoginResponse, error) {
	mobile, err := normalizeMobile(req.Mobile)
	if err != nil {
		return domainauth.LoginResponse{}, err
	}
	scene := normalizeScene(req.Scene)
	code := strings.TrimSpace(req.Code)
	if code == "" {
		return domainauth.LoginResponse{}, fmt.Errorf("请输入验证码")
	}

	now := s.nowFunc()
	record, ok, err := s.repo.LatestValidSMSCode(mobile, scene, now)
	if err != nil {
		return domainauth.LoginResponse{}, err
	}
	if !ok || record.CodeHash != hashCode(mobile, scene, code) {
		return domainauth.LoginResponse{}, fmt.Errorf("验证码无效或已过期")
	}
	if err := s.repo.MarkSMSCodeUsed(record.ID, now); err != nil {
		return domainauth.LoginResponse{}, err
	}

	user, ok, err := s.repo.FindUserByMobile(mobile)
	if err != nil {
		return domainauth.LoginResponse{}, err
	}
	if !ok {
		user = domainauth.User{
			ID:        newID("user"),
			Name:      defaultNameFromMobile(mobile),
			Mobile:    mobile,
			Status:    "active",
			CreatedAt: now.Format("2006-01-02 15:04:05"),
		}
		if err := s.repo.CreateUserWithMobile(user, mobile, now); err != nil {
			return domainauth.LoginResponse{}, err
		}
	}

	token, tokenHash, err := newToken()
	if err != nil {
		return domainauth.LoginResponse{}, err
	}
	if err := s.repo.SaveSession(SessionRecord{
		ID:        newID("session"),
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: now.Add(sessionTTL),
		CreatedAt: now,
	}); err != nil {
		return domainauth.LoginResponse{}, err
	}

	me, err := s.currentUserFor(user)
	if err != nil {
		return domainauth.LoginResponse{}, err
	}
	return domainauth.LoginResponse{Token: token, Me: me}, nil
}

func (s *AuthService) CurrentUser(token string) (domainauth.CurrentUser, bool, error) {
	token = strings.TrimSpace(strings.TrimPrefix(token, "Bearer "))
	if token == "" {
		return domainauth.CurrentUser{}, false, nil
	}
	session, ok, err := s.repo.FindSession(hashToken(token), s.nowFunc())
	if err != nil || !ok {
		return domainauth.CurrentUser{}, ok, err
	}
	user, ok, err := s.findUserByID(session.UserID)
	if err != nil || !ok {
		return domainauth.CurrentUser{}, ok, err
	}
	me, err := s.currentUserFor(user)
	return me, err == nil, err
}

func (s *AuthService) findUserByID(userID string) (domainauth.User, bool, error) {
	if repo, ok := s.repo.(interface {
		FindUserByID(userID string) (domainauth.User, bool, error)
	}); ok {
		return repo.FindUserByID(userID)
	}
	return domainauth.User{}, false, fmt.Errorf("repository does not support FindUserByID")
}

func (s *AuthService) currentUserFor(user domainauth.User) (domainauth.CurrentUser, error) {
	identities, err := s.repo.ListWorkIdentities(user.ID, user.Mobile)
	if err != nil {
		return domainauth.CurrentUser{}, err
	}
	defaultRoleID := ""
	if len(identities) > 0 {
		defaultRoleID = identities[0].ID
	}
	return domainauth.CurrentUser{User: user, DefaultRoleID: defaultRoleID, WorkIdentities: identities}, nil
}

type MemoryRepository struct {
	mu       sync.RWMutex
	users    map[string]domainauth.User
	byMobile map[string]string
	codes    []SMSCodeRecord
	sessions map[string]SessionRecord
}

func NewMemoryRepository() *MemoryRepository {
	repo := &MemoryRepository{
		users:    map[string]domainauth.User{},
		byMobile: map[string]string{},
		codes:    []SMSCodeRecord{},
		sessions: map[string]SessionRecord{},
	}
	for _, seed := range []domainauth.User{
		{ID: "user-teacher-li", Name: "李老师", Mobile: "13800001003", Status: "active", CreatedAt: "2026-06-01 09:00:00"},
		{ID: "user-teacher-wang", Name: "王老师", Mobile: "13800001002", Status: "active", CreatedAt: "2026-06-01 09:00:00"},
		{ID: "user-teacher-zhang", Name: "张老师", Mobile: "13800001001", Status: "active", CreatedAt: "2026-06-01 09:00:00"},
	} {
		repo.users[seed.ID] = seed
		repo.byMobile[seed.Mobile] = seed.ID
	}
	return repo
}

func (r *MemoryRepository) SaveSMSCode(record SMSCodeRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.codes = append(r.codes, record)
	return nil
}

func (r *MemoryRepository) LatestValidSMSCode(mobile, scene string, now time.Time) (SMSCodeRecord, bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for i := len(r.codes) - 1; i >= 0; i-- {
		record := r.codes[i]
		if record.Mobile == mobile && record.Scene == scene && record.UsedAt.Time.IsZero() && now.Before(record.ExpiresAt) {
			return record, true, nil
		}
	}
	return SMSCodeRecord{}, false, nil
}

func (r *MemoryRepository) MarkSMSCodeUsed(id string, usedAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.codes {
		if r.codes[i].ID == id {
			r.codes[i].UsedAt = sql.NullTime{Time: usedAt, Valid: true}
			return nil
		}
	}
	return nil
}

func (r *MemoryRepository) FindUserByMobile(mobile string) (domainauth.User, bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	userID, ok := r.byMobile[mobile]
	if !ok {
		return domainauth.User{}, false, nil
	}
	user, ok := r.users[userID]
	return user, ok, nil
}

func (r *MemoryRepository) FindUserByID(userID string) (domainauth.User, bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	user, ok := r.users[userID]
	return user, ok, nil
}

func (r *MemoryRepository) CreateUserWithMobile(user domainauth.User, mobile string, _ time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.users[user.ID] = user
	r.byMobile[mobile] = user.ID
	return nil
}

func (r *MemoryRepository) SaveSession(session SessionRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sessions[session.TokenHash] = session
	return nil
}

func (r *MemoryRepository) FindSession(tokenHash string, now time.Time) (SessionRecord, bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	session, ok := r.sessions[tokenHash]
	if !ok || !now.Before(session.ExpiresAt) {
		return SessionRecord{}, false, nil
	}
	return session, true, nil
}

func (r *MemoryRepository) ListWorkIdentities(_ string, mobile string) ([]domainauth.WorkIdentity, error) {
	if mobile == "13800001003" {
		return []domainauth.WorkIdentity{
			{ID: "homeroom", RoleType: "homeroom_teacher", RoleLabel: "班主任", PrimaryLabel: "班级管理", SecondaryLabel: "高二（3）班 / 高一（8）班", ScopeType: "admin_class_group", ScopeID: "homeroom-li"},
			{ID: "subject-english", RoleType: "subject_teacher", RoleLabel: "任课老师", PrimaryLabel: "英语任课分析", SecondaryLabel: "高二（3）班 / 高一（8）班", ScopeType: "teaching_scope", ScopeID: "teacher-english-li", Subject: "英语"},
		}, nil
	}
	if mobile == "13800001002" {
		return []domainauth.WorkIdentity{
			{ID: "subject-math", RoleType: "subject_teacher", RoleLabel: "任课老师", PrimaryLabel: "数学任课分析", SecondaryLabel: "高二（3）班 / 高一（8）班", ScopeType: "teaching_scope", ScopeID: "teacher-math-wang", Subject: "数学"},
		}, nil
	}
	return []domainauth.WorkIdentity{
		{ID: "trial-homeroom", RoleType: "homeroom_teacher", RoleLabel: "班主任", PrimaryLabel: "个人试用班级", SecondaryLabel: "可创建班级并导入成绩", ScopeType: "personal_workspace", ScopeID: "trial"},
	}, nil
}

func normalizeMobile(mobile string) (string, error) {
	mobile = strings.TrimSpace(mobile)
	if len(mobile) != 11 {
		return "", fmt.Errorf("请输入 11 位手机号")
	}
	return mobile, nil
}

func normalizeScene(scene string) string {
	scene = strings.TrimSpace(scene)
	if scene == "" {
		return defaultScene
	}
	return scene
}

func randomDigits(length int) (string, error) {
	var builder strings.Builder
	for builder.Len() < length {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		builder.WriteString(n.String())
	}
	return builder.String(), nil
}

func newToken() (string, string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", "", err
	}
	token := hex.EncodeToString(raw)
	return token, hashToken(token), nil
}

func hashCode(mobile, scene, code string) string {
	sum := sha256.Sum256([]byte(mobile + ":" + scene + ":" + code))
	return hex.EncodeToString(sum[:])
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func newID(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}

func defaultNameFromMobile(mobile string) string {
	if len(mobile) >= 4 {
		return "用户" + mobile[len(mobile)-4:]
	}
	return "新用户"
}
