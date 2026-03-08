package repo

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Credential struct {
	UserID       int64     `json:"user_id"`
	Login        string    `json:"login"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type RefreshSession struct {
	ID           int64     `json:"id"`
	UserID       int64     `json:"user_id"`
	RefreshToken string    `json:"refresh_token"`
	UserAgent    string    `json:"user_agent"`
	IPAddress    string    `json:"ip_address"`
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
}

type CustomClaims struct {
	UserID int64  `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type Profile struct {
	ID           int64       `json:"id"`
	FirstName    string      `json:"first_name"`
	LastName     string      `json:"last_name"`
	PositionID   int64       `json:"position_id"`
	DepartmentID *int64      `json:"department_id,omitempty"`
	Email        string      `json:"email"`
	AvatarURL    string      `json:"avatar_url,omitempty"`
	Login        string      `json:"login"`
	Role         string      `json:"role"`
	IsActive     bool        `json:"is_active"`
	HireDate     time.Time   `json:"hire_date"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
	Department   *Department `json:"department,omitempty"`
	Skills       []Skill     `json:"skills,omitempty"`
}

type Department struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type Position struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type Skill struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}


