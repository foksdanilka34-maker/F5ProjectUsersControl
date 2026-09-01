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

type EmployeeEventPayload struct {
	EventID   string  `json:"event_id"`
	UserID    int64   `json:"user_id"`
	FullName  string  `json:"full_name"`
	PhotoURL  *string `json:"photo_url,omitempty"`
	Role      string  `json:"role"`
	Timestamp int64   `json:"timestamp"`
}
