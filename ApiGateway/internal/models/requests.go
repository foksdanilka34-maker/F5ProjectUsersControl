package models

import "time"

type LoginRequest struct {
	Login    string `json:"login" binding:"required,min=3"`
	Password string `json:"password" binding:"required,min=6"`
}

type CreateCredentialsRequest struct {
	UserID   string `json:"user_id" binding:"required"`
	Login    string `json:"login" binding:"required,min=3"`
	Password string `json:"password" binding:"required,min=6"`
	Role     string `json:"role" binding:"required,oneof=specialist manager admin director"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// Employee Profile Requests
type CreateProfileRequest struct {
	FirstName    string    `json:"first_name" binding:"required"`
	LastName     string    `json:"last_name" binding:"required"`
	PositionID   string    `json:"position_id" binding:"required"`
	Email        string    `json:"email" binding:"required,email"`
	DepartmentID *string   `json:"department_id,omitempty"`
	HireDate     time.Time `json:"hire_date" binding:"required"`
	Login        string    `json:"login" binding:"required,min=3"`
	Password     string    `json:"password" binding:"required,min=6"`
	Role         string    `json:"role" binding:"required,oneof=specialist manager admin director"`
}

type UpdateProfileRequest struct {
	FirstName    *string `json:"first_name,omitempty"`
	LastName     *string `json:"last_name,omitempty"`
	PositionID   *string `json:"position_id,omitempty"`
	Email        *string `json:"email,omitempty" binding:"omitempty,email"`
	DepartmentID *string `json:"department_id,omitempty"`
	AvatarURL    *string `json:"avatar_url,omitempty"`
}

type ListProfilesRequest struct {
	PageSize     int32  `form:"page_size" binding:"omitempty,min=1,max=100"`
	PageNumber   int32  `form:"page_number" binding:"omitempty,min=1"`
	DepartmentID string `form:"department_id"`
	PositionID   string `form:"position_id"`
}

type ChangeUserStatusRequest struct {
	Status bool `json:"status" binding:"required"`
}

// Department Requests
type CreateDepartmentRequest struct {
	Name string `json:"name" binding:"required,min=2"`
}

type UpdateDepartmentRequest struct {
	Name string `json:"name" binding:"required,min=2"`
}

// Position Requests
type CreatePositionRequest struct {
	Name string `json:"name" binding:"required,min=2"`
}

type UpdatePositionRequest struct {
	Name string `json:"name" binding:"required,min=2"`
}

// Skill Requests
type CreateSkillRequest struct {
	Name string `json:"name" binding:"required,min=2"`
}

type AddSkillToEmployeeRequest struct {
	SkillID string `json:"skill_id" binding:"required"`
}

type RemoveSkillFromEmployeeRequest struct {
	SkillID string `json:"skill_id" binding:"required"`
}

// Project Requests
type CreateProjectRequest struct {
	Name        string     `json:"name" binding:"required,min=2"`
	Description *string    `json:"description,omitempty"`
	ManagerID   string     `json:"manager_id" binding:"required"`
	DueDate     *time.Time `json:"due_date,omitempty"`
}

type UpdateProjectRequest struct {
	Name        *string    `json:"name,omitempty" binding:"omitempty,min=2"`
	Description *string    `json:"description,omitempty"`
	Status      *int32     `json:"status,omitempty" binding:"omitempty,min=0,max=3"`
	DueDate     *time.Time `json:"due_date,omitempty"`
}

type ListProjectsRequest struct {
	PageSize   int32   `form:"page_size" binding:"omitempty,min=1,max=100"`
	PageNumber int32   `form:"page_number" binding:"omitempty,min=1"`
	ManagerID  *string `form:"manager_id,omitempty"`
	Status     *int32  `form:"status,omitempty" binding:"omitempty,min=0,max=3"`
}

// Task Requests
type CreateTaskRequest struct {
	ProjectID   string    `json:"project_id" binding:"required"`
	Title       string    `json:"title" binding:"required,min=2"`
	Description string    `json:"description"`
	Priority    *int32    `json:"priority,omitempty" binding:"omitempty,min=0,max=4"`
	AssigneeID  *string   `json:"assignee_id,omitempty"`
	DueDate     time.Time `json:"due_date" binding:"required"`
}

type UpdateTaskRequest struct {
	Title       *string    `json:"title,omitempty" binding:"omitempty,min=2"`
	Description *string    `json:"description,omitempty"`
	Status      *int32     `json:"status,omitempty" binding:"omitempty,min=0,max=4"`
	Priority    *int32     `json:"priority,omitempty" binding:"omitempty,min=0,max=4"`
	AssigneeID  *string    `json:"assignee_id,omitempty"`
	DueDate     *time.Time `json:"due_date,omitempty"`
}

type MoveTaskRequest struct {
	NewStatus     int32 `json:"new_status" binding:"required,min=1,max=4"`
	NewOrderIndex int32 `json:"new_order_index" binding:"required,min=0"`
}

type AssignTaskRequest struct {
	AssigneeID string `json:"assignee_id" binding:"required"`
}

type ListTasksByProjectRequest struct {
	Status     *int32  `form:"status,omitempty" binding:"omitempty,min=0,max=4"`
	AssigneeID *string `form:"assignee_id,omitempty"`
	Priority   *int32  `form:"priority,omitempty" binding:"omitempty,min=0,max=4"`
}

// Project Member Requests
type AddMemberToProjectRequest struct {
	UserID string `json:"user_id" binding:"required"`
}

type RemoveMemberFromProjectRequest struct {
	UserID string `json:"user_id" binding:"required"`
}

// Task History Requests
type GetTaskStatusHistoryRequest struct {
	PageSize   int32 `form:"page_size" binding:"omitempty,min=1,max=100"`
	PageNumber int32 `form:"page_number" binding:"omitempty,min=1"`
}

// Project Metrics Requests
type GetProjectMetricsRequest struct {
	StartDate *time.Time `form:"start_date,omitempty"`
	EndDate   *time.Time `form:"end_date,omitempty"`
}
