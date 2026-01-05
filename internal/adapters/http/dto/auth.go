package dto

import "time"

type AuthRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
	Remember    bool   `json:"remember"`
}

type AuthResponse struct {
	AccessToken string       `json:"access_token"`
	TokenType   string       `json:"token_type"`
	ExpiresIn   int64        `json:"expires_in"`
	User        UserResponse `json:"user"`
}

type UserResponse struct {
	ID              string     `json:"id"`
	Email           string     `json:"email"`
	DisplayName     string     `json:"display_name"`
	AvatarURL       string     `json:"avatar_url,omitempty"`
	EmailVerifiedAt *time.Time `json:"email_verified_at,omitempty"`
}

type AuthSessionResponse struct {
	ID         string    `json:"id"`
	CreatedAt  time.Time `json:"created_at"`
	ExpiresAt  time.Time `json:"expires_at"`
	UserAgent  string    `json:"user_agent,omitempty"`
	IP         string    `json:"ip,omitempty"`
	DeviceName string    `json:"device_name,omitempty"`
	IsCurrent  bool      `json:"is_current"`
}
