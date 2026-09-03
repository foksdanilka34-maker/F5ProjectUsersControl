package core

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/business/dto"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/business/repo"
)

var ErrInvalidWebhookToken = errors.New("invalid gitlab webhook token")

const (
	eventPush         = "push"
	eventMergeRequest = "merge_request"
	eventPipeline     = "pipeline"
	eventNote         = "note"
)

type MergeRequestCommenter interface {
	CommentOnMergeRequest(ctx context.Context, projectID, mergeRequestIID int64, body string) error
}

type GitLabWebhookService struct {
	repo        GitLabRepository
	tasks       *TaskService
	txManager   TxManager
	broadcaster EventBroadcaster
	commenter   MergeRequestCommenter
}

func NewGitLabWebhookService(
	repo GitLabRepository,
	tasks *TaskService,
	txManager TxManager,
	broadcaster EventBroadcaster,
	commenter MergeRequestCommenter,
) *GitLabWebhookService {
	return &GitLabWebhookService{
		repo:        repo,
		tasks:       tasks,
		txManager:   txManager,
		broadcaster: broadcaster,
		commenter:   commenter,
	}
}

func (s *GitLabWebhookService) Accept(ctx context.Context, projectID int64, token, eventType, deliveryID string, payload []byte) error {
	integration, err := s.repo.GetIntegration(ctx, projectID)
	if err != nil {
		return err
	}
	if integration == nil || !integration.IsActive {
		return ErrIntegrationNotConfigured
	}
	if !secureEqual(integration.WebhookSecret, token) {
		return ErrInvalidWebhookToken
	}

	eventType = normalizeEventType(eventType)
	if eventType == "" {
		return nil
	}

	return s.txManager.WithinTx(ctx, func(r *repo.RepositoryRegistry) error {
		_, err := r.GitLab().InsertWebhookEvent(ctx, projectID, eventType, deliveryID, payload)
		return err
	})
}

func (s *GitLabWebhookService) RunWorker(ctx context.Context, interval time.Duration, batchSize int) {
	log.Printf("[GitLab Webhook Worker] Started with interval %s, batch size %d", interval, batchSize)

	if err := s.txManager.WithinTx(ctx, func(r *repo.RepositoryRegistry) error {
		return r.GitLab().RequeueStaleWebhookEvents(ctx, 5*time.Minute)
	}); err != nil {
		log.Printf("[GitLab Webhook Worker] Failed to requeue stale events: %v", err)
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("[GitLab Webhook Worker] Context cancelled, shutting down...")
			return
		case <-ticker.C:
			s.processBatch(ctx, batchSize)
		}
	}
}

func (s *GitLabWebhookService) processBatch(ctx context.Context, batchSize int) {
	var events []dto.GitLabWebhookEvent

	err := s.txManager.WithinTx(ctx, func(r *repo.RepositoryRegistry) error {
		var err error
		events, err = r.GitLab().FetchPendingWebhookEvents(ctx, batchSize)
		return err
	})
	if err != nil {
		if ctx.Err() == nil {
			log.Printf("[GitLab Webhook Worker] Error fetching batch: %v", err)
		}
		return
	}

	for _, event := range events {
		if ctx.Err() != nil {
			return
		}

		procErr := s.processEvent(ctx, event)
		_ = s.txManager.WithinTx(ctx, func(r *repo.RepositoryRegistry) error {
			if procErr != nil {
				log.Printf("[GitLab Webhook Worker] Failed to process %s (%s): %v", event.ID, event.EventType, procErr)
				return r.GitLab().MarkWebhookFailed(ctx, event.ID, procErr.Error())
			}
			return r.GitLab().MarkWebhookProcessed(ctx, event.ID)
		})
	}
}

func (s *GitLabWebhookService) processEvent(ctx context.Context, event dto.GitLabWebhookEvent) error {
	integration, err := s.repo.GetIntegration(ctx, event.ProjectID)
	if err != nil {
		return err
	}
	if integration == nil {
		return nil
	}

	switch event.EventType {
	case eventPush:
		return s.handlePush(ctx, integration, event.Payload)
	case eventMergeRequest:
		return s.handleMergeRequest(ctx, integration, event.Payload)
	case eventPipeline:
		return s.handlePipeline(ctx, integration, event.Payload)
	case eventNote:
		return s.handleNote(ctx, integration, event.Payload)
	default:
		return nil
	}
}

func (s *GitLabWebhookService) handlePush(ctx context.Context, integration *dto.GitLabIntegration, raw []byte) error {
	var payload struct {
		Ref      string `json:"ref"`
		UserName string `json:"user_name"`
		Project  struct {
			WebURL string `json:"web_url"`
		} `json:"project"`
		Commits []struct {
			ID      string `json:"id"`
			Message string `json:"message"`
			URL     string `json:"url"`
		} `json:"commits"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return err
	}

	branch := strings.TrimPrefix(payload.Ref, "refs/heads/")
	if branch == "" {
		return nil
	}

	sources := []string{branch}
	for _, commit := range payload.Commits {
		sources = append(sources, commit.Message)
	}

	task, err := s.resolveTask(ctx, integration, sources...)
	if err != nil || task == nil {
		return err
	}

	branchURL := payload.Project.WebURL + "/-/tree/" + branch
	links := []dto.TaskGitLinkDTO{{
		TaskID:     task.ID,
		Kind:       dto.GitLinkBranch,
		ExternalID: branch,
		Title:      &branch,
		WebURL:     &branchURL,
		AuthorName: optional(payload.UserName),
	}}

	for i, commit := range payload.Commits {
		if i >= 3 {
			break
		}
		title := firstLine(commit.Message)
		links = append(links, dto.TaskGitLinkDTO{
			TaskID:     task.ID,
			Kind:       dto.GitLinkCommit,
			ExternalID: shortSHA(commit.ID),
			Title:      &title,
			WebURL:     optional(commit.URL),
			AuthorName: optional(payload.UserName),
		})
	}

	if err := s.saveLinks(ctx, links); err != nil {
		return err
	}

	if integration.AutoMoveInProgress && task.Status == "TODO" {
		if _, err := s.tasks.SystemTransition(ctx, task.ID, "IN_PROGRESS",
			fmt.Sprintf("GitLab: работа начата в ветке %s", branch)); err != nil {
			return err
		}
	}

	return nil
}

func (s *GitLabWebhookService) handleMergeRequest(ctx context.Context, integration *dto.GitLabIntegration, raw []byte) error {
	var payload struct {
		User struct {
			Name string `json:"name"`
		} `json:"user"`
		ObjectAttributes struct {
			IID          int64  `json:"iid"`
			Title        string `json:"title"`
			Description  string `json:"description"`
			SourceBranch string `json:"source_branch"`
			State        string `json:"state"`
			Action       string `json:"action"`
			URL          string `json:"url"`
		} `json:"object_attributes"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return err
	}

	attrs := payload.ObjectAttributes
	task, err := s.resolveTask(ctx, integration, attrs.SourceBranch, attrs.Title, attrs.Description)
	if err != nil || task == nil {
		return err
	}

	link := dto.TaskGitLinkDTO{
		TaskID:     task.ID,
		Kind:       dto.GitLinkMergeRequest,
		ExternalID: strconv.FormatInt(attrs.IID, 10),
		Title:      optional(attrs.Title),
		State:      optional(attrs.State),
		WebURL:     optional(attrs.URL),
		AuthorName: optional(payload.User.Name),
	}
	if err := s.saveLinks(ctx, []dto.TaskGitLinkDTO{link}); err != nil {
		return err
	}

	switch attrs.Action {
	case "open", "reopen":
		if !integration.AutoMoveReview || task.Status == "REVIEW" || task.Status == "DONE" {
			return nil
		}
		updated, err := s.tasks.SubmitForReview(ctx, task.ID)
		if err != nil {
			return err
		}
		s.comment(ctx, integration.ProjectID, attrs.IID,
			fmt.Sprintf("F5: задача %s переведена на код-ревью, ревьюеры назначены автоматически.",
				BuildTaskKey(integration.TaskKeyPrefix, updated.ID)))

	case "merge":
		if !integration.AutoCloseOnMerge || task.Status == "DONE" {
			return nil
		}

		complete, err := s.tasks.IsReviewComplete(ctx, task.ID)
		if err != nil {
			return err
		}

		if !complete && task.Status == "REVIEW" {
			s.notifyAssignee(task, "review_pending",
				"Merge request влит, но задача ждёт подтверждения ревьюеров: "+task.Title)
			s.comment(ctx, integration.ProjectID, attrs.IID,
				"F5: ветка влита, но задача остаётся на ревью — не все ревьюеры подтвердили изменения.")
			return nil
		}

		if _, err := s.tasks.SystemTransition(ctx, task.ID, "DONE",
			fmt.Sprintf("GitLab: merge request !%d влит в %s", attrs.IID, integration.DefaultBranch)); err != nil {
			return err
		}
		s.notifyAssignee(task, "task_completed", "Задача закрыта после влития merge request: "+task.Title)

	case "close":
		if task.Status == "REVIEW" {
			if _, err := s.tasks.SystemTransition(ctx, task.ID, "IN_PROGRESS",
				fmt.Sprintf("GitLab: merge request !%d закрыт без влития", attrs.IID)); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *GitLabWebhookService) handlePipeline(ctx context.Context, integration *dto.GitLabIntegration, raw []byte) error {
	var payload struct {
		ObjectAttributes struct {
			ID         int64  `json:"id"`
			Ref        string `json:"ref"`
			SHA        string `json:"sha"`
			Status     string `json:"status"`
			Duration   *int   `json:"duration"`
			CreatedAt  string `json:"created_at"`
			FinishedAt string `json:"finished_at"`
			URL        string `json:"url"`
		} `json:"object_attributes"`
		MergeRequest struct {
			SourceBranch string `json:"source_branch"`
		} `json:"merge_request"`
		Commit struct {
			Message string `json:"message"`
		} `json:"commit"`
		Project struct {
			WebURL string `json:"web_url"`
		} `json:"project"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return err
	}

	attrs := payload.ObjectAttributes
	if attrs.ID == 0 {
		return nil
	}

	task, err := s.resolveTask(ctx, integration, attrs.Ref, payload.MergeRequest.SourceBranch, payload.Commit.Message)
	if err != nil {
		return err
	}

	record := dto.GitLabPipelineDTO{
		ProjectID:   integration.ProjectID,
		PipelineID:  attrs.ID,
		Ref:         attrs.Ref,
		SHA:         shortSHA(attrs.SHA),
		Status:      attrs.Status,
		DurationSec: attrs.Duration,
		StartedAt:   parseGitLabTime(attrs.CreatedAt),
		FinishedAt:  parseGitLabTime(attrs.FinishedAt),
		UpdatedAt:   time.Now(),
	}

	if attrs.URL != "" {
		record.WebURL = &attrs.URL
	} else if payload.Project.WebURL != "" {
		url := fmt.Sprintf("%s/-/pipelines/%d", payload.Project.WebURL, attrs.ID)
		record.WebURL = &url
	}

	if task != nil {
		record.TaskID = &task.ID
	}

	if err := s.txManager.WithinTx(ctx, func(r *repo.RepositoryRegistry) error {
		return r.GitLab().UpsertPipeline(ctx, &record)
	}); err != nil {
		return err
	}

	if s.broadcaster != nil {
		s.broadcaster.Broadcast("gitlab:pipeline", record)
	}

	if task != nil && attrs.Status == "failed" {
		s.notifyAssignee(task, "pipeline_failed", "Пайплайн упал по задаче: "+task.Title)
	}

	return nil
}

func (s *GitLabWebhookService) handleNote(ctx context.Context, integration *dto.GitLabIntegration, raw []byte) error {
	var payload struct {
		User struct {
			Name string `json:"name"`
		} `json:"user"`
		ObjectAttributes struct {
			Note         string `json:"note"`
			NoteableType string `json:"noteable_type"`
		} `json:"object_attributes"`
		MergeRequest struct {
			SourceBranch string `json:"source_branch"`
			Title        string `json:"title"`
		} `json:"merge_request"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return err
	}

	if payload.ObjectAttributes.NoteableType != "MergeRequest" || payload.ObjectAttributes.Note == "" {
		return nil
	}

	task, err := s.resolveTask(ctx, integration, payload.MergeRequest.SourceBranch, payload.MergeRequest.Title)
	if err != nil || task == nil {
		return err
	}

	author := payload.User.Name
	if author == "" {
		author = "GitLab"
	}

	comment := dto.TaskCommentDTO{
		TaskID:    task.ID,
		UserID:    0,
		Content:   fmt.Sprintf("[GitLab] %s: %s", author, strings.TrimSpace(payload.ObjectAttributes.Note)),
		CreatedAt: time.Now(),
	}

	return s.txManager.WithinTx(ctx, func(r *repo.RepositoryRegistry) error {
		_, err := r.Task().CreateComment(ctx, &comment)
		return err
	})
}

func (s *GitLabWebhookService) resolveTask(ctx context.Context, integration *dto.GitLabIntegration, sources ...string) (*dto.TaskDTO, error) {
	taskID := ParseTaskID(integration.TaskKeyPrefix, sources...)
	if taskID == 0 {
		return nil, nil
	}

	task, err := s.tasks.GetTask(ctx, taskID)
	if errors.Is(err, ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if task.ProjectID != integration.ProjectID {
		return nil, nil
	}

	return &task, nil
}

func (s *GitLabWebhookService) saveLinks(ctx context.Context, links []dto.TaskGitLinkDTO) error {
	err := s.txManager.WithinTx(ctx, func(r *repo.RepositoryRegistry) error {
		for i := range links {
			if err := r.GitLab().UpsertLink(ctx, &links[i]); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	if s.broadcaster != nil {
		for _, link := range links {
			s.broadcaster.Broadcast("gitlab:link", link)
		}
	}
	return nil
}

func (s *GitLabWebhookService) comment(ctx context.Context, projectID, mergeRequestIID int64, body string) {
	if s.commenter == nil || mergeRequestIID == 0 {
		return
	}
	if err := s.commenter.CommentOnMergeRequest(ctx, projectID, mergeRequestIID, body); err != nil {
		log.Printf("[GitLab Webhook Worker] Failed to comment on MR !%d: %v", mergeRequestIID, err)
	}
}

func (s *GitLabWebhookService) notifyAssignee(task *dto.TaskDTO, kind, message string) {
	if s.broadcaster == nil || task.AssigneeID == nil {
		return
	}
	s.broadcaster.Broadcast("notification", map[string]any{
		"target_user_id": *task.AssigneeID,
		"kind":           kind,
		"task_id":        task.ID,
		"task_title":     task.Title,
		"project_id":     task.ProjectID,
		"message":        message,
	})
}

func secureEqual(expected, actual string) bool {
	return subtle.ConstantTimeCompare([]byte(expected), []byte(actual)) == 1
}

func normalizeEventType(header string) string {
	switch strings.ToLower(strings.TrimSpace(header)) {
	case "push", "push hook", "tag push hook":
		return eventPush
	case "merge_request", "merge request hook":
		return eventMergeRequest
	case "pipeline", "pipeline hook":
		return eventPipeline
	case "note", "note hook":
		return eventNote
	default:
		return ""
	}
}

func parseGitLabTime(value string) *time.Time {
	if value == "" {
		return nil
	}
	layouts := []string{time.RFC3339, "2006-01-02 15:04:05 UTC", "2006-01-02 15:04:05 -0700"}
	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, value); err == nil {
			return &parsed
		}
	}
	return nil
}

func firstLine(value string) string {
	line := strings.TrimSpace(strings.SplitN(value, "\n", 2)[0])
	if len(line) > 200 {
		line = line[:200]
	}
	return line
}

func shortSHA(sha string) string {
	if len(sha) > 8 {
		return sha[:8]
	}
	return sha
}

func optional(value string) *string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return &value
}
