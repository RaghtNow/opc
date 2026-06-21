package auth

type User struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Mobile    string `json:"mobile"`
	Status    string `json:"status"`
	CreatedAt string `json:"createdAt"`
}

type Identity struct {
	ID             string `json:"id"`
	UserID         string `json:"userId"`
	IdentityType   string `json:"identityType"`
	Identifier     string `json:"identifier"`
	Verified       bool   `json:"verified"`
	LastVerifiedAt string `json:"lastVerifiedAt"`
}

type WorkIdentity struct {
	ID             string `json:"id"`
	RoleType       string `json:"roleType"`
	RoleLabel      string `json:"roleLabel"`
	PrimaryLabel   string `json:"primaryLabel"`
	SecondaryLabel string `json:"secondaryLabel"`
	ScopeType      string `json:"scopeType"`
	ScopeID        string `json:"scopeId"`
	Subject        string `json:"subject"`
}

type CurrentUser struct {
	User           User           `json:"user"`
	DefaultRoleID  string         `json:"defaultRoleId"`
	WorkIdentities []WorkIdentity `json:"workIdentities"`
}

type SendSMSCodeRequest struct {
	Mobile string `json:"mobile"`
	Scene  string `json:"scene"`
}

type SendSMSCodeResponse struct {
	Mobile    string `json:"mobile"`
	Scene     string `json:"scene"`
	ExpiresIn int    `json:"expiresIn"`
	DevCode   string `json:"devCode,omitempty"`
	Message   string `json:"message"`
}

type LoginWithSMSRequest struct {
	Mobile string `json:"mobile"`
	Code   string `json:"code"`
	Scene  string `json:"scene"`
}

type LoginResponse struct {
	Token string      `json:"token"`
	Me    CurrentUser `json:"me"`
}
