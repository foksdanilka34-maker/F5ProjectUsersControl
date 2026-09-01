package dto

import "time"

type OutboxStatus string

const (
	OutboxPending   OutboxStatus = "PENDING"
	OutboxPublished OutboxStatus = "PUBLISHED"
	OutboxFailed    OutboxStatus = "FAILED"
)

type OutboxRecord struct {
	ID           string       `json:"id"`
	EventType    string       `json:"event_type"`
	Payload      []byte       `json:"payload"`
	Status       OutboxStatus `json:"status"`
	RetryCount   int          `json:"retry_count"`
	ErrorMessage *string      `json:"error_message,omitempty"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
	ProcessedAt  *time.Time   `json:"processed_at,omitempty"`
}

type TaskEventPayload struct {
	EventID    string    `json:"event_id"`
	TaskID     int64     `json:"task_id"`
	ProjectID  int64     `json:"project_id"`
	Title      string    `json:"title"`
	Status     string    `json:"status"`
	AssigneeID *int64    `json:"assignee_id,omitempty"`
	Timestamp  time.Time `json:"timestamp"`
}

type ProjectEventPayload struct {
	EventID   string    `json:"event_id"`
	ProjectID int64     `json:"project_id"`
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}
