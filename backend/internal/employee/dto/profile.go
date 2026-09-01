package dto

import "time"

type DepartmentDTO struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type PositionDTO struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type SkillDTO struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type ProfileDTO struct {
	ID         int64          `json:"id"`
	FirstName  string         `json:"first_name"`
	LastName   string         `json:"last_name"`
	PositionID int64          `json:"position_id"`
	Email      string         `json:"email"`
	AvatarURL  *string        `json:"avatar_url,omitempty"`
	HireDate   string         `json:"hire_date"`
	Department *DepartmentDTO `json:"department,omitempty"`
	Skills     []SkillDTO     `json:"skills,omitempty"`
	Login      string         `json:"login"`
	Role       string         `json:"role"`
	IsActive   bool           `json:"is_active"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}

type CreateProfileRequest struct {
	FirstName    string  `json:"first_name"`
	LastName     string  `json:"last_name"`
	PositionID   int64   `json:"position_id"`
	Email        string  `json:"email"`
	DepartmentID *int64  `json:"department_id,omitempty"`
	HireDate     string  `json:"hire_date"`
	Login        string  `json:"login"`
	Password     string  `json:"password"`
	Role         string  `json:"role"`
	AvatarURL    *string `json:"avatar_url,omitempty"`
}

type UpdateProfileRequest struct {
	FirstName    *string `json:"first_name,omitempty"`
	LastName     *string `json:"last_name,omitempty"`
	PositionID   *int64  `json:"position_id,omitempty"`
	DepartmentID *int64  `json:"department_id,omitempty"`
	Email        *string `json:"email,omitempty"`
	AvatarURL    *string `json:"avatar_url,omitempty"`
}

type ListProfilesFilter struct {
	PageSize     int   `json:"page_size"`
	PageNumber   int   `json:"page_number"`
	DepartmentID int64 `json:"department_id,omitempty"`
	PositionID   int64 `json:"position_id,omitempty"`
}

type ListProfilesResponse struct {
	Profiles   []ProfileDTO `json:"profiles"`
	TotalCount int          `json:"total_count"`
}
