package events

import (
	"encoding/json"
	"time"
)

type EventType string

const (
	EventTypeTaskCreated       EventType = "task.created"
	EventTypeTaskUpdated       EventType = "task.updated"
	EventTypeTaskStatusChanged EventType = "task.status_changed"
	EventTypeTaskDeleted       EventType = "task.deleted"
	EventTypeProjectCreated    EventType = "project.created"
	EventTypeProjectUpdated    EventType = "project.updated"
	EventTypeProjectDeleted    EventType = "project.deleted"
)

type TaskEvent struct {
	EventType   EventType  `json:"event_type"`
	TaskID      string     `json:"task_id"`
	ProjectID   string     `json:"project_id"`
	Status      string     `json:"status,omitempty"`
	OldStatus   string     `json:"old_status,omitempty"`
	AssigneeID  *string    `json:"assignee_id,omitempty"`
	CreatorID   string     `json:"creator_id,omitempty"`
	Priority    string     `json:"priority,omitempty"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	Timestamp   time.Time  `json:"timestamp"`
}

type ProjectEvent struct {
	EventType EventType  `json:"event_type"`
	ProjectID string     `json:"project_id"`
	ManagerID string     `json:"manager_id,omitempty"`
	Status    string     `json:"status,omitempty"`
	OldStatus string     `json:"old_status,omitempty"`
	DueDate   *time.Time `json:"due_date,omitempty"`
	Timestamp time.Time  `json:"timestamp"`
}

func (e *TaskEvent) Marshal() ([]byte, error) {
	return json.Marshal(e)
}

func (e *ProjectEvent) Marshal() ([]byte, error) {
	return json.Marshal(e)
}
