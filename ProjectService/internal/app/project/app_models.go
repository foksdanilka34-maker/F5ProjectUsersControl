package employee

import (
	"time"
)

type ProjectStatus int
type TaskStatus int
type TaskPriority int
type ProjectRole int

const (
	ProjectStatusUnspecified ProjectStatus = iota
	ProjectStatusActive
	ProjectStatusOnHold
	ProjectStatusArchived
)

const (
	TaskStatusUnspecified TaskStatus = iota
	TaskStatusTodo
	TaskStatusInProgress
	TaskStatusReview
	TaskStatusDone
)

const (
	PriorityUnspecified TaskPriority = iota
	PriorityLow
	PriorityMedium
	PriorityHigh
	PriorityCritical
)

const (
	RoleUnspecified ProjectRole = iota
	RoleEmployee
	RoleAdmin
	RoleDeveloper
	RoleDirector
	RoleManager
)

func (r ProjectRole) String() string {
	switch r {
	case RoleEmployee:
		return "Employee"
	case RoleAdmin:
		return "Admin"
	case RoleDeveloper:
		return "Developer"
	case RoleManager:
		return "Manager"
	case RoleDirector:
		return "Director"
	default:
		return "Unspecified"
	}
}

func ParseProjectRole(s string) ProjectRole {
	switch s {
	case "Employee":
		return RoleEmployee
	case "Admin":
		return RoleAdmin
	case "Developer":
		return RoleDeveloper
	case "Manager":
		return RoleManager
	case "Directir":
		return RoleDirector
	default:
		return RoleUnspecified
	}
}

type Project struct {
	ID          string
	Name        string
	Description string
	ManagerID   string
	Status      ProjectStatus
	DueDate     *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type CreateProjectRequest struct {
	Name        string
	Description string
	ManagerID   string
	DueDate     *time.Time
}

type UpdateProjectRequest struct {
	ID          string
	Name        *string
	Description *string
	Status      *ProjectStatus
	DueDate     *time.Time
}

type ListProjectsFilter struct {
	PageSize   int32
	PageNumber int32
	ManagerID  *string
	Status     *ProjectStatus
}

type ProjectMember struct {
	UserID   string
	FullName string
	Role     ProjectRole
}

type AddMemberRequest struct {
	ProjectID string
	UserID    string
	Role      ProjectRole
}

type RemoveMemberRequest struct {
	ProjectID string
	UserID    string
}

type Task struct {
	ID          string
	ProjectID   string
	Title       string
	Description string
	Status      TaskStatus
	Priority    TaskPriority
	AssigneeID  *string
	CreatorID   string
	OrderIndex  int32
	DueDate     *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type CreateTaskRequest struct {
	ProjectID   string
	Title       string
	Description string
	Priority    TaskPriority
	AssigneeID  *string
	CreatorID   string
	DueDate     *time.Time
}

type UpdateTaskRequest struct {
	ID          string
	Title       *string
	Description *string
	Status      *TaskStatus
	Priority    *TaskPriority
	AssigneeID  *string
	DueDate     *time.Time
}

type MoveTaskRequest struct {
	TaskID        string
	NewStatus     TaskStatus
	NewOrderIndex int32
}

type AssignTaskRequest struct {
	TaskID     string
	AssigneeID *string
}

type ListTasksFilter struct {
	ProjectID  string
	Status     *TaskStatus
	AssigneeID *string
	Priority   *TaskPriority
}

type ProjectsListResponse struct {
	Projects   []Project
	TotalCount int32
}

type TasksListResponse struct {
	Tasks []Task
}

type ProjectMembersResponse struct {
	Members []ProjectMember
}
