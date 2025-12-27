package repo

import (
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TaskRepo struct {
	db *pgxpool.Pool
}

func NewTaskRepo(db *pgxpool.Pool) *TaskRepo {
	return &TaskRepo{db: db}
}

func (r *TaskRepo) Create(ctx context.Context, t *Task) (int64, error) {
	var id int64
	query := `
		INSERT INTO tasks (project_id, title, description, status, priority, assignee_id, creator_id, due_date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id
	`
	err := r.db.QueryRow(ctx, query, t.ProjectID, t.Title, t.Description, t.Status, t.Priority, t.AssigneeID, t.CreatorID, t.DueDate, t.CreatedAt, t.UpdatedAt).Scan(&id)
	return id, err
}

func (r *TaskRepo) GetByID(ctx context.Context, id int64) (*Task, error) {
	query := `
		SELECT t.id, t.project_id, t.title, t.description, t.status, t.priority, t.assignee_id, t.creator_id, t.due_date, t.created_at, t.updated_at,
			p.name as project_name
		FROM tasks t
		JOIN projects p ON t.project_id = p.id
		WHERE t.id = $1
	`
	var t Task
	err := r.db.QueryRow(ctx, query, id).Scan(
		&t.ID, &t.ProjectID, &t.Title, &t.Description, &t.Status, &t.Priority, &t.AssigneeID, &t.CreatorID, &t.DueDate, &t.CreatedAt, &t.UpdatedAt,
		&t.ProjectName,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TaskRepo) List(ctx context.Context, pageSize, offset int, projectID, assigneeID int64, status, priority string) ([]*Task, int, error) {
	// Count query
	countBuilder := psql.Select("COUNT(*)").From("tasks t")
	if projectID > 0 {
		countBuilder = countBuilder.Where(sq.Eq{"t.project_id": projectID})
	}
	if assigneeID > 0 {
		countBuilder = countBuilder.Where(sq.Eq{"t.assignee_id": assigneeID})
	}
	if status != "" {
		countBuilder = countBuilder.Where(sq.Expr("UPPER(t.status) = UPPER(?)", status))
	}
	if priority != "" {
		countBuilder = countBuilder.Where(sq.Expr("LOWER(t.priority) = LOWER(?)", priority))
	}

	countQuery, countArgs, err := countBuilder.ToSql()
	if err != nil {
		return nil, 0, err
	}

	var total int
	if err := r.db.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Select query
	selectBuilder := psql.Select(
		"t.id", "t.project_id", "t.title", "t.description", "t.status", "t.priority",
		"t.assignee_id", "t.creator_id", "t.due_date", "t.created_at", "t.updated_at",
		"p.name as project_name",
	).From("tasks t").
		Join("projects p ON t.project_id = p.id").
		OrderBy("t.created_at DESC").
		Limit(uint64(pageSize)).
		Offset(uint64(offset))

	if projectID > 0 {
		selectBuilder = selectBuilder.Where(sq.Eq{"t.project_id": projectID})
	}
	if assigneeID > 0 {
		selectBuilder = selectBuilder.Where(sq.Eq{"t.assignee_id": assigneeID})
	}
	if status != "" {
		selectBuilder = selectBuilder.Where(sq.Expr("UPPER(t.status) = UPPER(?)", status))
	}
	if priority != "" {
		selectBuilder = selectBuilder.Where(sq.Expr("LOWER(t.priority) = LOWER(?)", priority))
	}

	query, args, err := selectBuilder.ToSql()
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var tasks []*Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(
			&t.ID, &t.ProjectID, &t.Title, &t.Description, &t.Status, &t.Priority,
			&t.AssigneeID, &t.CreatorID, &t.DueDate, &t.CreatedAt, &t.UpdatedAt,
			&t.ProjectName,
		); err != nil {
			return nil, 0, err
		}
		tasks = append(tasks, &t)
	}
	return tasks, total, nil
}

func (r *TaskRepo) Update(ctx context.Context, id int64, title, description, status, priority *string, assigneeID *int64, dueDate *time.Time) error {
	builder := psql.Update("tasks").Where(sq.Eq{"id": id})

	if title != nil {
		builder = builder.Set("title", *title)
	}
	if description != nil {
		builder = builder.Set("description", *description)
	}
	if status != nil {
		builder = builder.Set("status", *status)
		// Устанавливаем completed_at когда задача завершена
		if *status == "DONE" {
			builder = builder.Set("completed_at", time.Now())
		} else {
			// Сбрасываем completed_at если задача возвращена в работу
			builder = builder.Set("completed_at", nil)
		}
	}
	if priority != nil {
		builder = builder.Set("priority", *priority)
	}
	if assigneeID != nil {
		if *assigneeID == 0 {
			builder = builder.Set("assignee_id", nil)
		} else {
			builder = builder.Set("assignee_id", *assigneeID)
		}
	}
	if dueDate != nil {
		builder = builder.Set("due_date", *dueDate)
	}

	builder = builder.Set("updated_at", time.Now())

	query, args, err := builder.ToSql()
	if err != nil {
		return err
	}

	_, err = r.db.Exec(ctx, query, args...)
	return err
}

func (r *TaskRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.db.Exec(ctx, "DELETE FROM tasks WHERE id = $1", id)
	return err
}

func (r *TaskRepo) GetUserTasks(ctx context.Context, userID int64, status string) ([]*Task, error) {
	query := `
		SELECT t.id, t.project_id, t.title, t.description, t.status, t.priority, t.assignee_id, t.creator_id, t.due_date, t.created_at, t.updated_at,
			p.name as project_name
		FROM tasks t
		JOIN projects p ON t.project_id = p.id
		WHERE t.assignee_id = $1
	`
	args := []interface{}{userID}
	if status != "" {
		query += " AND UPPER(t.status) = UPPER($2)"
		args = append(args, status)
	}
	query += " ORDER BY t.due_date ASC NULLS LAST, t.priority DESC"

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(
			&t.ID, &t.ProjectID, &t.Title, &t.Description, &t.Status, &t.Priority, &t.AssigneeID, &t.CreatorID, &t.DueDate, &t.CreatedAt, &t.UpdatedAt,
			&t.ProjectName,
		); err != nil {
			return nil, err
		}
		tasks = append(tasks, &t)
	}
	return tasks, nil
}

// Task Comments
func (r *TaskRepo) CreateComment(ctx context.Context, c *TaskComment) (int64, error) {
	var id int64
	query := `INSERT INTO task_comments (task_id, user_id, content, created_at) VALUES ($1, $2, $3, $4) RETURNING id`
	err := r.db.QueryRow(ctx, query, c.TaskID, c.UserID, c.Content, c.CreatedAt).Scan(&id)
	return id, err
}

func (r *TaskRepo) GetComments(ctx context.Context, taskID int64) ([]*TaskComment, error) {
	query := `SELECT id, task_id, user_id, content, created_at FROM task_comments WHERE task_id = $1 ORDER BY created_at ASC`
	rows, err := r.db.Query(ctx, query, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []*TaskComment
	for rows.Next() {
		var c TaskComment
		if err := rows.Scan(&c.ID, &c.TaskID, &c.UserID, &c.Content, &c.CreatedAt); err != nil {
			return nil, err
		}
		comments = append(comments, &c)
	}
	return comments, nil
}

func (r *TaskRepo) DeleteComment(ctx context.Context, id int64) error {
	_, err := r.db.Exec(ctx, "DELETE FROM task_comments WHERE id = $1", id)
	return err
}

// Task History
func (r *TaskRepo) AddHistory(ctx context.Context, h *TaskHistory) error {
	query := `INSERT INTO task_history (task_id, user_id, field, old_value, new_value, changed_at) VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.Exec(ctx, query, h.TaskID, h.UserID, h.Field, h.OldValue, h.NewValue, h.ChangedAt)
	return err
}

func (r *TaskRepo) GetHistory(ctx context.Context, taskID int64) ([]*TaskHistory, error) {
	query := `SELECT id, task_id, user_id, field, old_value, new_value, changed_at FROM task_history WHERE task_id = $1 ORDER BY changed_at DESC`
	rows, err := r.db.Query(ctx, query, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []*TaskHistory
	for rows.Next() {
		var h TaskHistory
		if err := rows.Scan(&h.ID, &h.TaskID, &h.UserID, &h.Field, &h.OldValue, &h.NewValue, &h.ChangedAt); err != nil {
			return nil, err
		}
		history = append(history, &h)
	}
	return history, nil
}
