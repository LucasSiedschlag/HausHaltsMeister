package auth

import "time"

type User struct {
	ID              string
	Email           string
	DisplayName     string
	AvatarURL       string
	EmailVerifiedAt *time.Time
	IsActive        bool
	CreatedAt       time.Time
	UpdatedAt       *time.Time
}

type AuthSession struct {
	ID        string
	UserID    string
	ExpiresAt time.Time
	RevokedAt *time.Time
	IsPersistent bool
}

type OAuthState struct {
	ID          string
	Provider    string
	State       string
	CodeVerifier string
	RedirectURI string
	ExpiresAt   time.Time
	UsedAt      *time.Time
}

type AuthIdentity struct {
	ID             string
	UserID         string
	Provider       string
	ProviderUserID string
	Email          string
	DisplayName    string
	AvatarURL      string
	EmailVerified  bool
	LastLoginAt    *time.Time
}

type AuthTokens struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
	SessionID    string
}

type AuthResult struct {
	User   User
	Tokens AuthTokens
	Session AuthSession
}
