package eventbus

import (
	"time"
)

const (
	ProjectTasksTopic    = "project.tasks"
	ProjectProjectsTopic = "project.projects"

	LoginDeactivateUserCommandTopic = "login.command.deactivate"

	EmployeeCreatedEventTopic = "employee.event.created"
	EmployeeUpdatedEventTopic = "employee.event.updated"
	EmployeeDeletedEventTopic = "employee.event.deleted"

	EventTypeTaskCreated      	= "task.created"
	EventTypeTaskUpdated       	= "task.updated"
	EventTypeTaskStatusChanged	= "task.status_changed"
	EventTypeTaskDeleted      	= "task.deleted"
	EventTypeTaskAssigned  		= "task.assigned"
	EventTypeTaskCompleted 		= "task.completed"

	EventTypeProjectCreated     = "project.created"
	EventTypeProjectUpdated     = "project.updated"
	EventTypeProjectDeleted   	= "project.deleted"

	EventTypeProjectMemberAdd   = "project.add.member"
	EventTypeProjectMemberDel   = "project.delete.member"
)

type TaskEvent struct {
	TaskID      string     `json:"task_id"`
	ProjectID   string     `json:"project_id"`
	Status      string     `json:"status,omitempty"`
	OldStatus   string     `json:"old_status,omitempty"`
	AssigneeID  string    `json:"assignee_id,omitempty"`
	CreatorID   string     `json:"creator_id,omitempty"`
	Priority    string     `json:"priority,omitempty"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	Timestamp   time.Time  `json:"timestamp"`
}

type ProjectEvent struct {
	ProjectID string     `json:"project_id"`
	ManagerID string     `json:"manager_id,omitempty"`
	MemberID  string	 `json:"member_id,omitempty"`
	Status    string     `json:"status,omitempty"`
	OldStatus string     `json:"old_status,omitempty"`
	DueDate   *time.Time `json:"due_date,omitempty"`
	Timestamp time.Time  `json:"timestamp"`
}

