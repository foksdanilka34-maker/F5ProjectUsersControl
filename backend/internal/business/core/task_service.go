package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/business/dto"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/business/repo"
	"github.com/google/uuid"
)

const (
	ReviewAssignPrefix  = "REVIEW:ASSIGN:"
	ReviewApprovePrefix = "REVIEW:APPROVE:"
)

type TaskRepository interface {
	Create(ctx context.Context, t *dto.TaskDTO) (int64, error)
	GetByID(ctx context.Context, id int64) (*dto.TaskDTO, error)
	List(ctx context.Context, filter dto.ListTasksFilter) ([]dto.TaskDTO, error)
	Update(ctx context.Context, id int64, req dto.UpdateTaskRequest) error
	Delete(ctx context.Context, id int64) error

	CreateComment(ctx context.Context, c *dto.TaskCommentDTO) (int64, error)
	GetComments(ctx context.Context, taskID int64) ([]dto.TaskCommentDTO, error)
	DeleteComment(ctx context.Context, id int64) error

	AddHistory(ctx context.Context, h *dto.TaskHistoryDTO) error
	GetHistory(ctx context.Context, taskID int64) ([]dto.TaskHistoryDTO, error)
}

type EventBroadcaster interface {
	Broadcast(eventType string, payload any)
}

type TaskService struct {
	repo        TaskRepository
	projectRepo ProjectRepository
	txManager   TxManager
	broadcaster EventBroadcaster
}

func NewTaskService(
	repo TaskRepository,
	projectRepo ProjectRepository,
	txManager TxManager,
	broadcaster EventBroadcaster,
) *TaskService {
	return &TaskService{
		repo:        repo,
		projectRepo: projectRepo,
		txManager:   txManager,
		broadcaster: broadcaster,
	}
}

func (s *TaskService) CreateTask(ctx context.Context, creatorID int64, req dto.CreateTaskRequest) (dto.TaskDTO, error) {
	if req.ProjectID == 0 || req.Title == "" {
		return dto.TaskDTO{}, errors.New("project_id and title are required")
	}

	priority := "medium"
	if req.PriorityStr != nil && *req.PriorityStr != "" {
		priority = strings.ToLower(*req.PriorityStr)
	} else if req.Priority != nil {
		switch *req.Priority {
		case 1:
			priority = "low"
		case 2:
			priority = "medium"
		case 3:
			priority = "high"
		case 4:
			priority = "critical"
		}
	}

	var dueDate *time.Time
	if req.DueDate != nil && *req.DueDate != "" {
		if t, err := time.Parse("2006-01-02", *req.DueDate); err == nil {
			dueDate = &t
		}
	}

	task := &dto.TaskDTO{
		ProjectID:   req.ProjectID,
		Title:       req.Title,
		Description: req.Description,
		Status:      "TODO",
		Priority:    priority,
		AssigneeID:  req.AssigneeID,
		CreatorID:   creatorID,
		DueDate:     dueDate,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	var id int64
	err := s.txManager.WithinTx(ctx, func(r *repo.RepositoryRegistry) error {
		var err error
		id, err = r.Task().Create(ctx, task)
		if err != nil {
			return err
		}
		task.ID = id

		// Outbox Event
		eventPayload := dto.TaskEventPayload{
			EventID:    uuid.New().String(),
			TaskID:     id,
			ProjectID:  task.ProjectID,
			Title:      task.Title,
			Status:     task.Status,
			AssigneeID: task.AssigneeID,
			Timestamp:  time.Now(),
		}
		payloadBytes, _ := json.Marshal(eventPayload)
		_, err = r.Outbox().Insert(ctx, "task.event.created", payloadBytes)
		return err
	})

	if err != nil {
		return dto.TaskDTO{}, err
	}

	if s.broadcaster != nil {
		s.broadcaster.Broadcast("task:created", task)
		if task.AssigneeID != nil {
			s.broadcaster.Broadcast("notification", map[string]any{
				"target_user_id": *task.AssigneeID,
				"kind":           "task_assigned",
				"task_id":        task.ID,
				"task_title":     task.Title,
				"project_id":     task.ProjectID,
				"message":        "Вам назначена задача: " + task.Title,
			})
		}
	}

	return *task, nil
}

func (s *TaskService) GetTask(ctx context.Context, id int64) (dto.TaskDTO, error) {
	t, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return dto.TaskDTO{}, err
	}
	if t == nil {
		return dto.TaskDTO{}, ErrNotFound
	}
	return *t, nil
}

func (s *TaskService) ListTasks(ctx context.Context, filter dto.ListTasksFilter) ([]dto.TaskDTO, error) {
	return s.repo.List(ctx, filter)
}

func (s *TaskService) UpdateTask(ctx context.Context, id, userID int64, req dto.UpdateTaskRequest) (dto.TaskDTO, error) {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil || existing == nil {
		return dto.TaskDTO{}, ErrNotFound
	}

	if req.Priority != nil {
		var p string
		switch *req.Priority {
		case 1:
			p = "low"
		case 2:
			p = "medium"
		case 3:
			p = "high"
		case 4:
			p = "critical"
		default:
			p = "medium"
		}
		req.PriorityStr = &p
	}

	err = s.txManager.WithinTx(ctx, func(r *repo.RepositoryRegistry) error {
		if req.Title != nil && *req.Title != existing.Title {
			_ = r.Task().AddHistory(ctx, &dto.TaskHistoryDTO{
				TaskID: id, UserID: userID, Field: "title", OldValue: existing.Title, NewValue: *req.Title, ChangedAt: time.Now(),
			})
		}
		if req.Status != nil && *req.Status != existing.Status {
			_ = r.Task().AddHistory(ctx, &dto.TaskHistoryDTO{
				TaskID: id, UserID: userID, Field: "status", OldValue: existing.Status, NewValue: *req.Status, ChangedAt: time.Now(),
			})
		}

		if err := r.Task().Update(ctx, id, req); err != nil {
			return err
		}

		eventPayload := dto.TaskEventPayload{
			EventID:    uuid.New().String(),
			TaskID:     id,
			ProjectID:  existing.ProjectID,
			Title:      existing.Title,
			Status:     existing.Status,
			AssigneeID: existing.AssigneeID,
			Timestamp:  time.Now(),
		}
		if req.Title != nil {
			eventPayload.Title = *req.Title
		}
		if req.Status != nil {
			eventPayload.Status = *req.Status
		}
		if req.AssigneeID != nil {
			eventPayload.AssigneeID = req.AssigneeID
		}

		payloadBytes, _ := json.Marshal(eventPayload)
		_, err := r.Outbox().Insert(ctx, "task.event.updated", payloadBytes)
		return err
	})

	if err != nil {
		return dto.TaskDTO{}, err
	}

	updated, err := s.GetTask(ctx, id)
	if err == nil && s.broadcaster != nil {
		s.broadcaster.Broadcast("task:updated", updated)
	}
	return updated, err
}

func (s *TaskService) DeleteTask(ctx context.Context, id int64) error {
	err := s.txManager.WithinTx(ctx, func(r *repo.RepositoryRegistry) error {
		if err := r.Task().Delete(ctx, id); err != nil {
			return err
		}
		eventPayload := dto.TaskEventPayload{
			EventID:   uuid.New().String(),
			TaskID:    id,
			Timestamp: time.Now(),
		}
		payloadBytes, _ := json.Marshal(eventPayload)
		_, err := r.Outbox().Insert(ctx, "task.event.deleted", payloadBytes)
		return err
	})

	if err != nil {
		return err
	}

	if s.broadcaster != nil {
		s.broadcaster.Broadcast("task:deleted", map[string]int64{"id": id})
	}
	return nil
}

func (s *TaskService) MoveTask(ctx context.Context, id int64, req dto.MoveTaskRequest) (dto.TaskDTO, error) {
	if req.NewStatus == "REVIEW" {
		return s.SubmitForReview(ctx, id)
	}

	existing, err := s.GetTask(ctx, id)
	if err != nil {
		return dto.TaskDTO{}, err
	}

	if existing.Status == "REVIEW" {
		return dto.TaskDTO{}, errors.New("task is currently in review and cannot be moved manually")
	}

	updateReq := dto.UpdateTaskRequest{
		Status: &req.NewStatus,
	}

	err = s.txManager.WithinTx(ctx, func(r *repo.RepositoryRegistry) error {
		if err := r.Task().Update(ctx, id, updateReq); err != nil {
			return err
		}

		eventPayload := dto.TaskEventPayload{
			EventID:    uuid.New().String(),
			TaskID:     id,
			ProjectID:  existing.ProjectID,
			Title:      existing.Title,
			Status:     req.NewStatus,
			AssigneeID: existing.AssigneeID,
			Timestamp:  time.Now(),
		}
		payloadBytes, _ := json.Marshal(eventPayload)
		_, err := r.Outbox().Insert(ctx, "task.event.moved", payloadBytes)
		return err
	})

	if err != nil {
		return dto.TaskDTO{}, err
	}

	updated, err := s.GetTask(ctx, id)
	if err == nil && s.broadcaster != nil {
		s.broadcaster.Broadcast("task:moved", updated)
	}
	return updated, err
}

func (s *TaskService) AssignTask(ctx context.Context, id, assigneeID int64) (dto.TaskDTO, error) {
	updateReq := dto.UpdateTaskRequest{
		AssigneeID: &assigneeID,
	}

	var updated dto.TaskDTO
	err := s.txManager.WithinTx(ctx, func(r *repo.RepositoryRegistry) error {
		if err := r.Task().Update(ctx, id, updateReq); err != nil {
			return err
		}

		existing, err := r.Task().GetByID(ctx, id)
		if err != nil || existing == nil {
			return ErrNotFound
		}
		updated = *existing

		eventPayload := dto.TaskEventPayload{
			EventID:    uuid.New().String(),
			TaskID:     id,
			ProjectID:  existing.ProjectID,
			Title:      existing.Title,
			Status:     existing.Status,
			AssigneeID: &assigneeID,
			Timestamp:  time.Now(),
		}
		payloadBytes, _ := json.Marshal(eventPayload)
		_, err = r.Outbox().Insert(ctx, "task.event.assigned", payloadBytes)
		return err
	})

	if err != nil {
		return dto.TaskDTO{}, err
	}

	if s.broadcaster != nil {
		s.broadcaster.Broadcast("task:assigned", updated)
		s.broadcaster.Broadcast("notification", map[string]any{
			"target_user_id": assigneeID,
			"kind":           "task_assigned",
			"task_id":        updated.ID,
			"task_title":     updated.Title,
			"project_id":     updated.ProjectID,
			"message":        "Вам назначена задача: " + updated.Title,
		})
	}
	return updated, nil
}

func (s *TaskService) SubmitForReview(ctx context.Context, id int64) (dto.TaskDTO, error) {
	task, err := s.GetTask(ctx, id)
	if err != nil {
		return dto.TaskDTO{}, err
	}

	members, err := s.projectRepo.GetMembers(ctx, task.ProjectID)
	if err != nil {
		return dto.TaskDTO{}, fmt.Errorf("failed to get project members: %w", err)
	}

	var candidates []int64
	for _, m := range members {
		if task.AssigneeID != nil && m.UserID == *task.AssigneeID {
			continue
		}
		candidates = append(candidates, m.UserID)
	}

	reviewers := candidates
	if len(candidates) > 2 {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		r.Shuffle(len(candidates), func(i, j int) {
			candidates[i], candidates[j] = candidates[j], candidates[i]
		})
		reviewers = candidates[:2]
	}

	err = s.txManager.WithinTx(ctx, func(r *repo.RepositoryRegistry) error {
		comments, _ := r.Task().GetComments(ctx, id)
		for _, c := range comments {
			if strings.HasPrefix(c.Content, "REVIEW:") {
				_ = r.Task().DeleteComment(ctx, c.ID)
			}
		}

		for _, revID := range reviewers {
			_, err := r.Task().CreateComment(ctx, &dto.TaskCommentDTO{
				TaskID:    id,
				UserID:    revID,
				Content:   fmt.Sprintf("%s%d", ReviewAssignPrefix, revID),
				CreatedAt: time.Now(),
			})
			if err != nil {
				return err
			}
		}

		reviewStatus := "REVIEW"
		if err := r.Task().Update(ctx, id, dto.UpdateTaskRequest{Status: &reviewStatus}); err != nil {
			return err
		}

		// Outbox Event
		eventPayload := dto.TaskEventPayload{
			EventID:    uuid.New().String(),
			TaskID:     id,
			ProjectID:  task.ProjectID,
			Title:      task.Title,
			Status:     "REVIEW",
			AssigneeID: task.AssigneeID,
			Timestamp:  time.Now(),
		}
		payloadBytes, _ := json.Marshal(eventPayload)
		_, err := r.Outbox().Insert(ctx, "task.event.review_requested", payloadBytes)
		return err
	})

	if err != nil {
		return dto.TaskDTO{}, err
	}

	updated, err := s.GetTask(ctx, id)
	if err == nil && s.broadcaster != nil {
		s.broadcaster.Broadcast("task:moved", updated)
		for _, revID := range reviewers {
			s.broadcaster.Broadcast("notification", map[string]any{
				"target_user_id": revID,
				"kind":           "review_assigned",
				"task_id":        updated.ID,
				"task_title":     updated.Title,
				"project_id":     updated.ProjectID,
				"message":        "Вам назначена задача на ревью: " + updated.Title,
			})
		}
	}
	return updated, err
}

func (s *TaskService) ApproveTask(ctx context.Context, id, userID int64) (dto.TaskDTO, error) {
	task, err := s.GetTask(ctx, id)
	if err != nil {
		return dto.TaskDTO{}, err
	}
	if task.Status != "REVIEW" {
		return dto.TaskDTO{}, errors.New("task is not in review status")
	}

	comments, err := s.repo.GetComments(ctx, id)
	if err != nil {
		return dto.TaskDTO{}, err
	}

	isAssigned := false
	alreadyApproved := false
	for _, c := range comments {
		if c.Content == fmt.Sprintf("%s%d", ReviewAssignPrefix, userID) {
			isAssigned = true
		}
		if c.Content == fmt.Sprintf("%s%d", ReviewApprovePrefix, userID) {
			alreadyApproved = true
		}
	}

	if !isAssigned {
		return dto.TaskDTO{}, errors.New("user is not an assigned reviewer for this task")
	}
	if alreadyApproved {
		return dto.TaskDTO{}, errors.New("user has already approved this task")
	}

	var allDone bool
	err = s.txManager.WithinTx(ctx, func(r *repo.RepositoryRegistry) error {
		_, err := r.Task().CreateComment(ctx, &dto.TaskCommentDTO{
			TaskID:    id,
			UserID:    userID,
			Content:   fmt.Sprintf("%s%d", ReviewApprovePrefix, userID),
			CreatedAt: time.Now(),
		})
		if err != nil {
			return err
		}

		allComments, err := r.Task().GetComments(ctx, id)
		if err != nil {
			return err
		}

		assigned := make(map[int64]bool)
		approved := make(map[int64]bool)
		for _, c := range allComments {
			if strings.HasPrefix(c.Content, ReviewAssignPrefix) {
				uid, _ := strconv.ParseInt(strings.TrimPrefix(c.Content, ReviewAssignPrefix), 10, 64)
				if uid > 0 {
					assigned[uid] = true
				}
			}
			if strings.HasPrefix(c.Content, ReviewApprovePrefix) {
				uid, _ := strconv.ParseInt(strings.TrimPrefix(c.Content, ReviewApprovePrefix), 10, 64)
				if uid > 0 {
					approved[uid] = true
				}
			}
		}

		allApproved := len(assigned) > 0
		for uid := range assigned {
			if !approved[uid] {
				allApproved = false
				break
			}
		}

		eventType := "task.event.reviewed"
		if allApproved {
			doneStatus := "DONE"
			allDone = true
			if err := r.Task().Update(ctx, id, dto.UpdateTaskRequest{Status: &doneStatus}); err != nil {
				return err
			}
			eventType = "task.event.approved"
		}

		eventPayload := dto.TaskEventPayload{
			EventID:    uuid.New().String(),
			TaskID:     id,
			ProjectID:  task.ProjectID,
			Title:      task.Title,
			Status:     task.Status,
			AssigneeID: task.AssigneeID,
			Timestamp:  time.Now(),
		}
		payloadBytes, _ := json.Marshal(eventPayload)
		_, err = r.Outbox().Insert(ctx, eventType, payloadBytes)
		return err
	})

	if err != nil {
		return dto.TaskDTO{}, err
	}

	updated, err := s.GetTask(ctx, id)
	if err == nil && s.broadcaster != nil {
		s.broadcaster.Broadcast("task:updated", updated)
		if allDone && updated.AssigneeID != nil {
			s.broadcaster.Broadcast("notification", map[string]any{
				"target_user_id": *updated.AssigneeID,
				"kind":           "review_approved",
				"task_id":        updated.ID,
				"task_title":     updated.Title,
				"project_id":     updated.ProjectID,
				"message":        "Ваша задача одобрена всеми ревьюерами: " + updated.Title,
			})
		}
	}
	return updated, err
}

func (s *TaskService) GetReviewStatus(ctx context.Context, id int64) (dto.ReviewStatusResponse, error) {
	task, err := s.GetTask(ctx, id)
	if err != nil {
		return dto.ReviewStatusResponse{}, err
	}

	comments, err := s.repo.GetComments(ctx, id)
	if err != nil {
		return dto.ReviewStatusResponse{}, err
	}

	assigned := make(map[int64]bool)
	approved := make(map[int64]bool)
	for _, c := range comments {
		if strings.HasPrefix(c.Content, ReviewAssignPrefix) {
			uid, _ := strconv.ParseInt(strings.TrimPrefix(c.Content, ReviewAssignPrefix), 10, 64)
			if uid > 0 {
				assigned[uid] = true
			}
		}
		if strings.HasPrefix(c.Content, ReviewApprovePrefix) {
			uid, _ := strconv.ParseInt(strings.TrimPrefix(c.Content, ReviewApprovePrefix), 10, 64)
			if uid > 0 {
				approved[uid] = true
			}
		}
	}

	var reviewers []dto.ReviewerInfo
	for uid := range assigned {
		reviewers = append(reviewers, dto.ReviewerInfo{
			UserID:   uid,
			Approved: approved[uid],
		})
	}

	return dto.ReviewStatusResponse{
		Reviewers: reviewers,
		IsActive:  task.Status == "REVIEW",
	}, nil
}

func (s *TaskService) SystemTransition(ctx context.Context, id int64, newStatus, reason string) (dto.TaskDTO, error) {
	task, err := s.GetTask(ctx, id)
	if err != nil {
		return dto.TaskDTO{}, err
	}
	if task.Status == newStatus {
		return task, nil
	}

	err = s.txManager.WithinTx(ctx, func(r *repo.RepositoryRegistry) error {
		_ = r.Task().AddHistory(ctx, &dto.TaskHistoryDTO{
			TaskID: id, UserID: 0, Field: "status", OldValue: task.Status, NewValue: newStatus, ChangedAt: time.Now(),
		})

		if err := r.Task().Update(ctx, id, dto.UpdateTaskRequest{Status: &newStatus}); err != nil {
			return err
		}

		if reason != "" {
			_, _ = r.Task().CreateComment(ctx, &dto.TaskCommentDTO{
				TaskID:    id,
				UserID:    0,
				Content:   reason,
				CreatedAt: time.Now(),
			})
		}

		eventPayload := dto.TaskEventPayload{
			EventID:    uuid.New().String(),
			TaskID:     id,
			ProjectID:  task.ProjectID,
			Title:      task.Title,
			Status:     newStatus,
			AssigneeID: task.AssigneeID,
			Timestamp:  time.Now(),
		}
		payloadBytes, _ := json.Marshal(eventPayload)
		_, err := r.Outbox().Insert(ctx, "task.event.moved", payloadBytes)
		return err
	})

	if err != nil {
		return dto.TaskDTO{}, err
	}

	updated, err := s.GetTask(ctx, id)
	if err == nil && s.broadcaster != nil {
		s.broadcaster.Broadcast("task:moved", updated)
	}
	return updated, err
}

func (s *TaskService) IsReviewComplete(ctx context.Context, id int64) (bool, error) {
	status, err := s.GetReviewStatus(ctx, id)
	if err != nil {
		return false, err
	}
	if len(status.Reviewers) == 0 {
		return false, nil
	}
	for _, reviewer := range status.Reviewers {
		if !reviewer.Approved {
			return false, nil
		}
	}
	return true, nil
}
