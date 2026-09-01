package dto

import "time"

type LoginRequest struct {
	Login     string `json:"login"`
	Password  string `json:"password"`
	UserAgent string `json:"user_agent,omitempty"`
	IPAddress string `json:"ip_address,omitempty"`
}

type UserInfo struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	FullName  string `json:"full_name"`
	Role      string `json:"role"`
	AvatarURL string `json:"avatar_url,omitempty"`
}

type LoginResponse struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token,omitempty"`
	User         UserInfo `json:"user"`
}

type RefreshResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	User         *UserInfo `json:"user,omitempty"`
}

type ChangePasswordRequest struct {
	UserID      int64  `json:"user_id"`
	NewPassword string `json:"new_password"`
}

type ChangeUserStatusRequest struct {
	UserID   int64 `json:"user_id"`
	IsActive bool  `json:"is_active"`
}

type Session struct {
	ID           int64
	UserID       int64
	RefreshToken string
	UserAgent    string
	IPAddress    string
	ExpiresAt    time.Time
	CreatedAt    time.Time
}
