package dto

import "time"

const (
	ExtEventTaskCreated       = "task_created"
	ExtEventTaskStatusChanged = "task_status_changed"
	ExtEventCommentAdded      = "task_comment_added"
)

type ExtensionDTO struct {
	ID              int64     `json:"id"`
	Key             string    `json:"key"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	BaseURL         string    `json:"base_url"`
	SharedSecretEnc []byte    `json:"-"`
	TaskPanelURL    *string   `json:"task_panel_url,omitempty"`
	ProjectTabURL   *string   `json:"project_tab_url,omitempty"`
	ProjectTabLabel *string   `json:"project_tab_label,omitempty"`
	Events          []string  `json:"events"`
	IsActive        bool      `json:"is_active"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type SaveExtensionRequest struct {
	Key             string   `json:"key"`
	Name            string   `json:"name"`
	Description     string   `json:"description"`
	BaseURL         string   `json:"base_url"`
	SharedSecret    string   `json:"shared_secret"`
	TaskPanelURL    string   `json:"task_panel_url,omitempty"`
	ProjectTabURL   string   `json:"project_tab_url,omitempty"`
	ProjectTabLabel string   `json:"project_tab_label,omitempty"`
	Events          []string `json:"events"`
}

type ProjectExtensionDTO struct {
	ExtensionDTO
	Installed bool `json:"installed"`
	Enabled   bool `json:"enabled"`
}

type TaskEntityPropertyDTO struct {
	TaskID      int64     `json:"task_id"`
	ExtensionID int64     `json:"extension_id"`
	Key         string    `json:"key"`
	Value       []byte    `json:"value"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CommentEventPayload struct {
	EventID   string    `json:"event_id"`
	TaskID    int64     `json:"task_id"`
	ProjectID int64     `json:"project_id"`
	CommentID int64     `json:"comment_id"`
	UserID    int64     `json:"user_id"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}
