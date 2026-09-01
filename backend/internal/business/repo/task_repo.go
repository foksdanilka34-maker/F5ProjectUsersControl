package repo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/business/dto"
)

type TaskRepo struct {
	db DBExecutor
}

func NewTaskRepo(db DBExecutor) *TaskRepo {
	return &TaskRepo{db: db}
}

func (r *TaskRepo) Create(ctx context.Context, t *dto.TaskDTO) (int64, error) {
	query := `
		INSERT INTO business.tasks (project_id, title, description, status, priority, order_index, assignee_id, creator_id, due_date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
		RETURNING id
	`
	var id int64
	err := r.db.QueryRow(ctx, query,
		t.ProjectID, t.Title, t.Description, t.Status, t.Priority, t.OrderIndex, t.AssigneeID, t.CreatorID, t.DueDate,
	).Scan(&id)
	return id, err
}

func (r *TaskRepo) GetByID(ctx context.Context, id int64) (*dto.TaskDTO, error) {
	query := `
		SELECT id, project_id, title, description, status, priority, order_index, assignee_id, creator_id,
		       due_date, started_at, completed_at, created_at, updated_at
		FROM business.tasks
		WHERE id = $1
	`
	var t dto.TaskDTO
	err := r.db.QueryRow(ctx, query, id).Scan(
		&t.ID, &t.ProjectID, &t.Title, &t.Description, &t.Status, &t.Priority, &t.OrderIndex, &t.AssigneeID, &t.CreatorID,
		&t.DueDate, &t.StartedAt, &t.CompletedAt, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TaskRepo) List(ctx context.Context, filter dto.ListTasksFilter) ([]dto.TaskDTO, error) {
	whereClause := "WHERE project_id = $1"
	args := []interface{}{filter.ProjectID}
	argIdx := 2

	if filter.AssigneeID > 0 {
		whereClause += fmt.Sprintf(" AND assignee_id = $%d", argIdx)
		args = append(args, filter.AssigneeID)
		argIdx++
	}
	if filter.Status != "" {
		whereClause += fmt.Sprintf(" AND UPPER(status) = UPPER($%d)", argIdx)
		args = append(args, filter.Status)
		argIdx++
	}
	if filter.Priority != "" {
		whereClause += fmt.Sprintf(" AND LOWER(priority) = LOWER($%d)", argIdx)
		args = append(args, filter.Priority)
		argIdx++
	}

	query := fmt.Sprintf(`
		SELECT id, project_id, title, description, status, priority, order_index, assignee_id, creator_id,
		       due_date, started_at, completed_at, created_at, updated_at
		FROM business.tasks
		%s
		ORDER BY order_index ASC, id ASC
	`, whereClause)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []dto.TaskDTO
	for rows.Next() {
		var t dto.TaskDTO
		if err := rows.Scan(
			&t.ID, &t.ProjectID, &t.Title, &t.Description, &t.Status, &t.Priority, &t.OrderIndex, &t.AssigneeID, &t.CreatorID,
			&t.DueDate, &t.StartedAt, &t.CompletedAt, &t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}

	return tasks, nil
}

func (r *TaskRepo) Update(ctx context.Context, id int64, req dto.UpdateTaskRequest) error {
	var completedAt *time.Time
	if req.Status != nil && strings.ToUpper(*req.Status) == "DONE" {
		now := time.Now()
		completedAt = &now
	}

	query := `
		UPDATE business.tasks
		SET title = COALESCE($1, title),
		    description = COALESCE($2, description),
		    status = COALESCE($3, status),
		    priority = COALESCE($4, priority),
		    assignee_id = COALESCE($5, assignee_id),
		    due_date = COALESCE($6, due_date),
		    completed_at = CASE WHEN $3 = 'DONE' THEN NOW() ELSE completed_at END,
		    updated_at = NOW()
		WHERE id = $7
	`
	_, err := r.db.Exec(ctx, query, req.Title, req.Description, req.Status, req.PriorityStr, req.AssigneeID, req.DueDate, id)
	_ = completedAt
	return err
}

func (r *TaskRepo) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM business.tasks WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *TaskRepo) CreateComment(ctx context.Context, c *dto.TaskCommentDTO) (int64, error) {
	query := `
		INSERT INTO business.task_comments (task_id, user_id, content, created_at)
		VALUES ($1, $2, $3, NOW())
		RETURNING id
	`
	var id int64
	err := r.db.QueryRow(ctx, query, c.TaskID, c.UserID, c.Content).Scan(&id)
	return id, err
}

func (r *TaskRepo) GetComments(ctx context.Context, taskID int64) ([]dto.TaskCommentDTO, error) {
	query := `
		SELECT id, task_id, user_id, content, created_at
		FROM business.task_comments
		WHERE task_id = $1
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(ctx, query, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []dto.TaskCommentDTO
	for rows.Next() {
		var c dto.TaskCommentDTO
		if err := rows.Scan(&c.ID, &c.TaskID, &c.UserID, &c.Content, &c.CreatedAt); err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}
	return comments, nil
}

func (r *TaskRepo) DeleteComment(ctx context.Context, id int64) error {
	query := `DELETE FROM business.task_comments WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *TaskRepo) AddHistory(ctx context.Context, h *dto.TaskHistoryDTO) error {
	query := `
		INSERT INTO business.task_history (task_id, user_id, field, old_value, new_value, changed_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
	`
	_, err := r.db.Exec(ctx, query, h.TaskID, h.UserID, h.Field, h.OldValue, h.NewValue)
	return err
}

func (r *TaskRepo) GetHistory(ctx context.Context, taskID int64) ([]dto.TaskHistoryDTO, error) {
	query := `
		SELECT id, task_id, user_id, field, old_value, new_value, changed_at
		FROM business.task_history
		WHERE task_id = $1
		ORDER BY changed_at DESC
	`
	rows, err := r.db.Query(ctx, query, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []dto.TaskHistoryDTO
	for rows.Next() {
		var h dto.TaskHistoryDTO
		if err := rows.Scan(&h.ID, &h.TaskID, &h.UserID, &h.Field, &h.OldValue, &h.NewValue, &h.ChangedAt); err != nil {
			return nil, err
		}
		history = append(history, h)
	}
	return history, nil
}
