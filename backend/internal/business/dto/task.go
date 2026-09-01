package dto

import "time"

type TaskDTO struct {
	ID          int64      `json:"id"`
	ProjectID   int64      `json:"project_id"`
	Title       string     `json:"title"`
	Description *string    `json:"description,omitempty"`
	Status      string     `json:"status"`
	Priority    string     `json:"priority"`
	OrderIndex  int        `json:"order_index"`
	AssigneeID  *int64     `json:"assignee_id,omitempty"`
	CreatorID   int64      `json:"creator_id"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type CreateTaskRequest struct {
	ProjectID   int64   `json:"project_id"`
	Title       string  `json:"title"`
	Description *string `json:"description,omitempty"`
	Priority    *int    `json:"priority,omitempty"` // 1=low, 2=med, 3=high, 4=crit
	PriorityStr *string `json:"priority_str,omitempty"`
	AssigneeID  *int64  `json:"assignee_id,omitempty"`
	DueDate     *string `json:"due_date,omitempty"`
}

type UpdateTaskRequest struct {
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
	Status      *string `json:"status,omitempty"`
	Priority    *int    `json:"priority,omitempty"`
	PriorityStr *string `json:"priority_str,omitempty"`
	AssigneeID  *int64  `json:"assignee_id,omitempty"`
	DueDate     *string `json:"due_date,omitempty"`
}

type MoveTaskRequest struct {
	NewStatus     string `json:"new_status"`
	NewOrderIndex int    `json:"new_order_index"`
}

type ListTasksFilter struct {
	ProjectID  int64  `json:"project_id"`
	AssigneeID int64  `json:"assignee_id,omitempty"`
	Status     string `json:"status,omitempty"`
	Priority   string `json:"priority,omitempty"`
}

type ReviewerInfo struct {
	UserID   int64 `json:"user_id"`
	Approved bool  `json:"approved"`
}

type ReviewStatusResponse struct {
	Reviewers []ReviewerInfo `json:"reviewers"`
	IsActive  bool           `json:"is_active"`
}

type TaskCommentDTO struct {
	ID        int64     `json:"id"`
	TaskID    int64     `json:"task_id"`
	UserID    int64     `json:"user_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type TaskHistoryDTO struct {
	ID        int64     `json:"id"`
	TaskID    int64     `json:"task_id"`
	UserID    int64     `json:"user_id"`
	Field     string    `json:"field"`
	OldValue  string    `json:"old_value"`
	NewValue  string    `json:"new_value"`
	ChangedAt time.Time `json:"changed_at"`
}
