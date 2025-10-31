package project

import (
	"strings"
	"time"
)

type ProjectStatus string

const (
	ProjectStatusUnspecified ProjectStatus = "PROJECT_STATUS_UNSPECIFIED"
	ProjectStatusActive      ProjectStatus = "ACTIVE"
	ProjectStatusOnHold      ProjectStatus = "ON_HOLD"
	ProjectStatusArchived    ProjectStatus = "ARCHIVED"
)

func (s ProjectStatus) String() string {
	if s == "" {
		return string(ProjectStatusUnspecified)
	}
	return string(s)
}

func (s ProjectStatus) ProtoValue() int {
	switch s {
	case ProjectStatusActive:
		return 1
	case ProjectStatusOnHold:
		return 2
	case ProjectStatusArchived:
		return 3
	case ProjectStatusUnspecified:
		fallthrough
	default:
		return 0
	}
}

func ProjectStatusFromProtoValue(value int32) ProjectStatus {
	switch value {
	case 1:
		return ProjectStatusActive
	case 2:
		return ProjectStatusOnHold
	case 3:
		return ProjectStatusArchived
	default:
		return ProjectStatusUnspecified
	}
}

func ProjectStatusFromDB(value string) ProjectStatus {
	switch strings.ToUpper(value) {
	case string(ProjectStatusActive):
		return ProjectStatusActive
	case string(ProjectStatusOnHold):
		return ProjectStatusOnHold
	case string(ProjectStatusArchived):
		return ProjectStatusArchived
	case "", string(ProjectStatusUnspecified):
		fallthrough
	default:
		return ProjectStatusUnspecified
	}
}

func (s ProjectStatus) DBValue() string {
	if s == "" {
		return string(ProjectStatusUnspecified)
	}
	return string(s)
}

type TaskStatus string

const (
	TaskStatusUnspecified TaskStatus = "TASK_STATUS_UNSPECIFIED"
	TaskStatusTodo        TaskStatus = "TODO"
	TaskStatusInProgress  TaskStatus = "IN_PROGRESS"
	TaskStatusReview      TaskStatus = "REVIEW"
	TaskStatusDone        TaskStatus = "DONE"
)

func (s TaskStatus) String() string {
	if s == "" {
		return string(TaskStatusUnspecified)
	}
	return string(s)
}

func (s TaskStatus) ProtoValue() int32 {
	switch s {
	case TaskStatusTodo:
		return 1
	case TaskStatusInProgress:
		return 2
	case TaskStatusReview:
		return 3
	case TaskStatusDone:
		return 4
	case TaskStatusUnspecified:
		fallthrough
	default:
		return 0
	}
}

func TaskStatusFromProtoValue(value int32) TaskStatus {
	switch value {
	case 1:
		return TaskStatusTodo
	case 2:
		return TaskStatusInProgress
	case 3:
		return TaskStatusReview
	case 4:
		return TaskStatusDone
	default:
		return TaskStatusUnspecified
	}
}

func TaskStatusFromDB(value string) TaskStatus {
	switch strings.ToUpper(value) {
	case string(TaskStatusTodo):
		return TaskStatusTodo
	case string(TaskStatusInProgress):
		return TaskStatusInProgress
	case string(TaskStatusReview):
		return TaskStatusReview
	case string(TaskStatusDone):
		return TaskStatusDone
	case "", string(TaskStatusUnspecified):
		fallthrough
	default:
		return TaskStatusUnspecified
	}
}

func (s TaskStatus) DBValue() string {
	if s == "" {
		return string(TaskStatusUnspecified)
	}
	return string(s)
}

type TaskPriority string

const (
	TaskPriorityUnspecified TaskPriority = "PRIORITY_UNSPECIFIED"
	TaskPriorityLow         TaskPriority = "LOW"
	TaskPriorityMedium      TaskPriority = "MEDIUM"
	TaskPriorityHigh        TaskPriority = "HIGH"
	TaskPriorityCritical    TaskPriority = "CRITICAL"
)

func (p TaskPriority) String() string {
	if p == "" {
		return string(TaskPriorityUnspecified)
	}
	return string(p)
}

func (p TaskPriority) ProtoValue() int32 {
	switch p {
	case TaskPriorityLow:
		return 1
	case TaskPriorityMedium:
		return 2
	case TaskPriorityHigh:
		return 3
	case TaskPriorityCritical:
		return 4
	case TaskPriorityUnspecified:
		fallthrough
	default:
		return 0
	}
}

func TaskPriorityFromProtoValue(value int32) TaskPriority {
	switch value {
	case 1:
		return TaskPriorityLow
	case 2:
		return TaskPriorityMedium
	case 3:
		return TaskPriorityHigh
	case 4:
		return TaskPriorityCritical
	default:
		return TaskPriorityUnspecified
	}
}

func TaskPriorityFromDB(value string) TaskPriority {
	switch strings.ToUpper(value) {
	case string(TaskPriorityLow):
		return TaskPriorityLow
	case string(TaskPriorityMedium):
		return TaskPriorityMedium
	case string(TaskPriorityHigh):
		return TaskPriorityHigh
	case string(TaskPriorityCritical):
		return TaskPriorityCritical
	case "", string(TaskPriorityUnspecified):
		fallthrough
	default:
		return TaskPriorityUnspecified
	}
}

func (p TaskPriority) DBValue() string {
	if p == "" {
		return string(TaskPriorityUnspecified)
	}
	return string(p)
}

type ProjectRole string

const (
	ProjectRoleUnspecified ProjectRole = "UNSPECIFIED"
	ProjectRoleMember      ProjectRole = "MEMBER"
	ProjectRoleManager     ProjectRole = "MANAGER"
	ProjectRoleOwner       ProjectRole = "OWNER"
)

func ParseProjectRole(value string) ProjectRole {
	switch strings.ToUpper(value) {
	case string(ProjectRoleMember):
		return ProjectRoleMember
	case string(ProjectRoleManager):
		return ProjectRoleManager
	case string(ProjectRoleOwner):
		return ProjectRoleOwner
	case "", string(ProjectRoleUnspecified):
		fallthrough
	default:
		return ProjectRoleUnspecified
	}
}

func (r ProjectRole) String() string {
	if r == "" {
		return string(ProjectRoleUnspecified)
	}
	return string(r)
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

type Task struct {
	ID          string
	ProjectID   string
	TaskName       string
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

type ProjectMember struct {
	ProjectID string
	UserID    string
	FullName  string
	Role      ProjectRole
	AddedAt   time.Time
}

type CreateProjectRequest struct {
	Name        string
	Description *string
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

type ProjectsListResponse struct {
	Projects   []*Project
	TotalCount int32
}

type CreateTaskRequest struct {
	ProjectID   string
	TaskName    string
	Description string
	Status      TaskStatus
	Priority    TaskPriority
	AssigneeID  *string
	CreatorID   string
	DueDate     time.Time
}

type UpdateTaskRequest struct {
	ID          string
	TaskName       *string
	Description *string
	Status      *TaskStatus
	Priority    *TaskPriority
	AssigneeID  *string
	DueDate     *time.Time
	OrderIndex  *int32
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

type TasksListResponse struct {
	Tasks []*Task
}

type ProjectMembersResponse struct {
	Members []*ProjectMember
}
