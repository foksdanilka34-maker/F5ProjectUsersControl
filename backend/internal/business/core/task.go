package core

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/business/repo"
)

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

type TaskService struct {
	repo        TaskRepository
	projectRepo ProjectRepository
}

func NewTaskService(repo TaskRepository) *TaskService {
	return &TaskService{repo: repo}
}

func (s *TaskService) SetProjectRepo(projectRepo ProjectRepository) {
	s.projectRepo = projectRepo
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
		Status:      "TODO",
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

func (s *TaskService) GetHistory(ctx context.Context, taskID int64) ([]*repo.TaskHistory, error) {
	return s.repo.GetHistory(ctx, taskID)
}

// Review constants for structured comment prefixes
const (
	ReviewAssignPrefix  = "REVIEW:ASSIGN:"
	ReviewApprovePrefix = "REVIEW:APPROVE:"
)

// ReviewStatus represents the current review state of a task
type ReviewStatus struct {
	Reviewers []ReviewerInfo `json:"reviewers"`
	IsActive  bool           `json:"is_active"`
}

type ReviewerInfo struct {
	UserID   int64 `json:"user_id"`
	Approved bool  `json:"approved"`
}

// SubmitForReview assigns 2 random reviewers from project team (excluding assignee) and sets status to IN_REVIEW
func (s *TaskService) SubmitForReview(ctx context.Context, taskID int64) (*repo.Task, error) {
	task, err := s.repo.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, ErrNotFound
	}

	if s.projectRepo == nil {
		return nil, fmt.Errorf("project repository not configured")
	}

	// Get project members
	members, err := s.projectRepo.GetMembers(ctx, task.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project members: %w", err)
	}

	// Filter out the assignee
	var candidates []int64
	for _, m := range members {
		if task.AssigneeID != nil && m.UserID == *task.AssigneeID {
			continue
		}
		candidates = append(candidates, m.UserID)
	}

	if len(candidates) < 2 {
		return nil, fmt.Errorf("not enough team members for review (need at least 2 reviewers, have %d)", len(candidates))
	}

	// Shuffle and pick 2
	rand.Shuffle(len(candidates), func(i, j int) {
		candidates[i], candidates[j] = candidates[j], candidates[i]
	})
	reviewers := candidates[:2]

	// Clear any previous review comments for this task
	existingComments, _ := s.repo.GetComments(ctx, taskID)
	for _, c := range existingComments {
		if strings.HasPrefix(c.Content, "REVIEW:") {
			_ = s.repo.DeleteComment(ctx, c.ID)
		}
	}

	// Create REVIEW:ASSIGN comments
	for _, reviewerID := range reviewers {
		comment := &repo.TaskComment{
			TaskID:    taskID,
			UserID:    reviewerID,
			Content:   fmt.Sprintf("%s%d", ReviewAssignPrefix, reviewerID),
			CreatedAt: time.Now(),
		}
		if _, err := s.repo.CreateComment(ctx, comment); err != nil {
			return nil, fmt.Errorf("failed to assign reviewer: %w", err)
		}
	}

	// Set status to IN_REVIEW
	inReview := "IN_REVIEW"
	if err := s.repo.Update(ctx, taskID, nil, nil, &inReview, nil, nil, nil); err != nil {
		return nil, err
	}

	s.addHistory(ctx, taskID, 0, "status", task.Status, "IN_REVIEW")

	return s.repo.GetByID(ctx, taskID)
}

// ApproveTask approves a task review by the given user
func (s *TaskService) ApproveTask(ctx context.Context, taskID, userID int64) (*repo.Task, error) {
	task, err := s.repo.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, ErrNotFound
	}

	if task.Status != "IN_REVIEW" {
		return nil, fmt.Errorf("task is not in review status")
	}

	// Check that user is an assigned reviewer
	comments, err := s.repo.GetComments(ctx, taskID)
	if err != nil {
		return nil, err
	}

	isReviewer := false
	for _, c := range comments {
		if c.Content == fmt.Sprintf("%s%d", ReviewAssignPrefix, userID) {
			isReviewer = true
			break
		}
	}
	if !isReviewer {
		return nil, fmt.Errorf("user is not assigned as reviewer for this task")
	}

	// Check if already approved
	for _, c := range comments {
		if c.Content == fmt.Sprintf("%s%d", ReviewApprovePrefix, userID) {
			return nil, fmt.Errorf("user has already approved this task")
		}
	}

	// Create approval comment
	approveComment := &repo.TaskComment{
		TaskID:    taskID,
		UserID:    userID,
		Content:   fmt.Sprintf("%s%d", ReviewApprovePrefix, userID),
		CreatedAt: time.Now(),
	}
	if _, err := s.repo.CreateComment(ctx, approveComment); err != nil {
		return nil, fmt.Errorf("failed to create approval: %w", err)
	}

	// Check if all reviewers have approved
	reviewStatus := s.parseReviewStatus(comments)
	// Add current approval
	for i, r := range reviewStatus.Reviewers {
		if r.UserID == userID {
			reviewStatus.Reviewers[i].Approved = true
		}
	}

	allApproved := true
	for _, r := range reviewStatus.Reviewers {
		if !r.Approved {
			allApproved = false
			break
		}
	}

	if allApproved {
		// Move to DONE
		done := "DONE"
		if err := s.repo.Update(ctx, taskID, nil, nil, &done, nil, nil, nil); err != nil {
			return nil, err
		}
		s.addHistory(ctx, taskID, userID, "status", "IN_REVIEW", "DONE")
	}

	return s.repo.GetByID(ctx, taskID)
}

// GetReviewStatus returns the review status of a task
func (s *TaskService) GetReviewStatus(ctx context.Context, taskID int64) (*ReviewStatus, error) {
	task, err := s.repo.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, ErrNotFound
	}

	comments, err := s.repo.GetComments(ctx, taskID)
	if err != nil {
		return nil, err
	}

	status := s.parseReviewStatus(comments)
	status.IsActive = task.Status == "IN_REVIEW"
	return status, nil
}

func (s *TaskService) parseReviewStatus(comments []*repo.TaskComment) *ReviewStatus {
	status := &ReviewStatus{
		Reviewers: []ReviewerInfo{},
	}

	// Find assigned reviewers
	reviewerMap := make(map[int64]*ReviewerInfo)
	for _, c := range comments {
		if strings.HasPrefix(c.Content, ReviewAssignPrefix) {
			idStr := strings.TrimPrefix(c.Content, ReviewAssignPrefix)
			var uid int64
			fmt.Sscanf(idStr, "%d", &uid)
			if uid > 0 {
				reviewerMap[uid] = &ReviewerInfo{UserID: uid, Approved: false}
			}
		}
	}

	// Check approvals
	for _, c := range comments {
		if strings.HasPrefix(c.Content, ReviewApprovePrefix) {
			idStr := strings.TrimPrefix(c.Content, ReviewApprovePrefix)
			var uid int64
			fmt.Sscanf(idStr, "%d", &uid)
			if r, ok := reviewerMap[uid]; ok {
				r.Approved = true
			}
		}
	}

	for _, r := range reviewerMap {
		status.Reviewers = append(status.Reviewers, *r)
	}

	return status
}
