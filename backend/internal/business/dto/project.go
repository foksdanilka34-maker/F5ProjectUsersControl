package dto

import "time"

type ProjectDTO struct {
	ID          int64      `json:"id"`
	Name        string     `json:"name"`
	Description *string    `json:"description,omitempty"`
	ManagerID   int64      `json:"manager_id"`
	Status      string     `json:"status"`
	StartDate   *time.Time `json:"start_date,omitempty"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type CreateProjectRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	ManagerID   int64   `json:"manager_id"`
	DueDate     *string `json:"due_date,omitempty"`
}

type UpdateProjectRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Status      *string `json:"status,omitempty"`
	ManagerID   *int64  `json:"manager_id,omitempty"`
	DueDate     *string `json:"due_date,omitempty"`
}

type ListProjectsFilter struct {
	PageSize   int    `json:"page_size"`
	PageNumber int    `json:"page_number"`
	ManagerID  int64  `json:"manager_id,omitempty"`
	MemberID   int64  `json:"member_id,omitempty"`
	Status     string `json:"status,omitempty"`
}

type ListProjectsResponse struct {
	Projects   []ProjectDTO `json:"projects"`
	TotalCount int          `json:"total_count"`
}

type ProjectMemberDTO struct {
	UserID   int64  `json:"user_id"`
	FullName string `json:"full_name"`
	Role     string `json:"role"`
	PhotoURL *string `json:"photo_url,omitempty"`
}
