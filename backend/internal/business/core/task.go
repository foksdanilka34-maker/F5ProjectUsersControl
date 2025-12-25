package core

import (
	"context"
	"time"

	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/business/repo"
)

// TaskRepository - интерфейс репозитория задач
type TaskRepository interface {
	Create(ctx context.Context, t *repo.Task) (int64, error)
	GetByID(ctx context.Context, id int64) (*repo.Task, error)
	List(ctx context.Context, pageSize, offset int, projectID, assigneeID int64, status, priority string) ([]*repo.Task, int, error)
	Update(ctx context.Context, id int64, title, description, status, priority *string, assigneeID *int64, dueDate *time.Time) error
	Delete(ctx context.Context, id int64) error
	GetUserTasks(ctx context.Context, userID int64, status string) ([]*repo.Task, error)

	CreateComment(ctx context.Context, c *repo.TaskComment) (int64, error)
	GetComments(ctx context.Context, taskID int64) ([]*repo.TaskComment, error)
	DeleteComment(ctx context.Context, id int64) error

	AddHistory(ctx context.Context, h *repo.TaskHistory) error
	GetHistory(ctx context.Context, taskID int64) ([]*repo.TaskHistory, error)
}

// TaskService - сервис задач
type TaskService struct {
	repo TaskRepository
}

func NewTaskService(repo TaskRepository) *TaskService {
	return &TaskService{repo: repo}
}

type CreateTaskRequest struct {
	ProjectID   int64
	Title       string
	Description string
	Priority    string
	AssigneeID  *int64
	CreatorID   int64
	DueDate     *time.Time
}

func (s *TaskService) CreateTask(ctx context.Context, req *CreateTaskRequest) (*repo.Task, error) {
	priority := req.Priority
	if priority == "" {
		priority = "medium"
	}

	task := &repo.Task{
		ProjectID:   req.ProjectID,
		Title:       req.Title,
		Description: req.Description,
		Status:      "todo",
		Priority:    priority,
		AssigneeID:  req.AssigneeID,
		CreatorID:   req.CreatorID,
		DueDate:     req.DueDate,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	id, err := s.repo.Create(ctx, task)
	if err != nil {
		return nil, err
	}
	task.ID = id

	return task, nil
}

func (s *TaskService) GetTask(ctx context.Context, id int64) (*repo.Task, error) {
	task, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, ErrNotFound
	}
	return task, nil
}

type ListTasksFilter struct {
	PageSize   int
	PageNumber int
	ProjectID  int64
	AssigneeID int64
	Status     string
	Priority   string
}

func (s *TaskService) ListTasks(ctx context.Context, filter *ListTasksFilter) ([]*repo.Task, int, error) {
	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	pageNumber := filter.PageNumber
	if pageNumber <= 0 {
		pageNumber = 1
	}
	offset := (pageNumber - 1) * pageSize

	return s.repo.List(ctx, pageSize, offset, filter.ProjectID, filter.AssigneeID, filter.Status, filter.Priority)
}

type UpdateTaskRequest struct {
	Title       *string
	Description *string
	Status      *string
	Priority    *string
	AssigneeID  *int64
	DueDate     *time.Time
}

func (s *TaskService) UpdateTask(ctx context.Context, id, userID int64, req *UpdateTaskRequest) (*repo.Task, error) {
	task, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, ErrNotFound
	}

	// Track changes for history
	if req.Title != nil && *req.Title != task.Title {
		s.addHistory(ctx, id, userID, "title", task.Title, *req.Title)
	}
	if req.Status != nil && *req.Status != task.Status {
		s.addHistory(ctx, id, userID, "status", task.Status, *req.Status)
	}
	if req.Priority != nil && *req.Priority != task.Priority {
		s.addHistory(ctx, id, userID, "priority", task.Priority, *req.Priority)
	}

	if err := s.repo.Update(ctx, id, req.Title, req.Description, req.Status, req.Priority, req.AssigneeID, req.DueDate); err != nil {
		return nil, err
	}

	return s.repo.GetByID(ctx, id)
}

func (s *TaskService) addHistory(ctx context.Context, taskID, userID int64, field, oldValue, newValue string) {
	h := &repo.TaskHistory{
		TaskID:    taskID,
		UserID:    userID,
		Field:     field,
		OldValue:  oldValue,
		NewValue:  newValue,
		ChangedAt: time.Now(),
	}
	_ = s.repo.AddHistory(ctx, h)
}

func (s *TaskService) DeleteTask(ctx context.Context, id int64) error {
	task, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if task == nil {
		return ErrNotFound
	}
	return s.repo.Delete(ctx, id)
}

func (s *TaskService) GetUserTasks(ctx context.Context, userID int64, status string) ([]*repo.Task, error) {
	return s.repo.GetUserTasks(ctx, userID, status)
}

// Comments
type CreateCommentRequest struct {
	TaskID  int64
	UserID  int64
	Content string
}

func (s *TaskService) CreateComment(ctx context.Context, req *CreateCommentRequest) (*repo.TaskComment, error) {
	comment := &repo.TaskComment{
		TaskID:    req.TaskID,
		UserID:    req.UserID,
		Content:   req.Content,
		CreatedAt: time.Now(),
	}

	id, err := s.repo.CreateComment(ctx, comment)
	if err != nil {
		return nil, err
	}
	comment.ID = id

	return comment, nil
}

func (s *TaskService) GetComments(ctx context.Context, taskID int64) ([]*repo.TaskComment, error) {
	return s.repo.GetComments(ctx, taskID)
}

func (s *TaskService) DeleteComment(ctx context.Context, id int64) error {
	return s.repo.DeleteComment(ctx, id)
}

// History
func (s *TaskService) GetHistory(ctx context.Context, taskID int64) ([]*repo.TaskHistory, error) {
	return s.repo.GetHistory(ctx, taskID)
}
