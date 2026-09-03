package core

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/business/dto"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/business/repo"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/pkg/gitlab"
)

var (
	ErrIntegrationNotConfigured = errors.New("gitlab integration is not configured for this project")
	ErrIntegrationNoToken       = errors.New("gitlab access token is not set")
)

type GitLabRepository interface {
	GetIntegration(ctx context.Context, projectID int64) (*dto.GitLabIntegration, error)
	ListLinks(ctx context.Context, taskID int64) ([]dto.TaskGitLinkDTO, error)
	ListPipelines(ctx context.Context, taskID int64) ([]dto.GitLabPipelineDTO, error)
	ProjectSummary(ctx context.Context, projectID int64) ([]dto.ProjectGitSummaryItem, error)
}

type TokenSealer interface {
	Seal(plaintext string) ([]byte, error)
	Open(ciphertext []byte) (string, error)
}

type GitLabService struct {
	repo        GitLabRepository
	taskRepo    TaskRepository
	txManager   TxManager
	sealer      TokenSealer
	broadcaster EventBroadcaster
	publicURL   string
}

func NewGitLabService(
	repo GitLabRepository,
	taskRepo TaskRepository,
	txManager TxManager,
	sealer TokenSealer,
	broadcaster EventBroadcaster,
	publicURL string,
) *GitLabService {
	return &GitLabService{
		repo:        repo,
		taskRepo:    taskRepo,
		txManager:   txManager,
		sealer:      sealer,
		broadcaster: broadcaster,
		publicURL:   strings.TrimRight(publicURL, "/"),
	}
}

func (s *GitLabService) GetIntegration(ctx context.Context, projectID int64) (*dto.GitLabIntegrationResponse, error) {
	integration, err := s.repo.GetIntegration(ctx, projectID)
	if err != nil || integration == nil {
		return nil, err
	}
	return s.toResponse(integration), nil
}

func (s *GitLabService) SaveIntegration(ctx context.Context, projectID int64, req dto.SaveGitLabIntegrationRequest) (*dto.GitLabIntegrationResponse, error) {
	if req.GitLabProjectID <= 0 {
		return nil, errors.New("gitlab_project_id is required")
	}

	existing, err := s.repo.GetIntegration(ctx, projectID)
	if err != nil {
		return nil, err
	}

	integration := &dto.GitLabIntegration{
		ProjectID:          projectID,
		BaseURL:            strings.TrimRight(strings.TrimSpace(req.BaseURL), "/"),
		GitLabProjectID:    req.GitLabProjectID,
		DefaultBranch:      strings.TrimSpace(req.DefaultBranch),
		TaskKeyPrefix:      NormalizeTaskKeyPrefix(req.TaskKeyPrefix),
		AutoMoveInProgress: true,
		AutoMoveReview:     true,
		AutoCloseOnMerge:   true,
		IsActive:           true,
	}

	if integration.BaseURL == "" {
		integration.BaseURL = "https://gitlab.com"
	}
	if integration.DefaultBranch == "" {
		integration.DefaultBranch = "main"
	}

	if existing != nil {
		integration.WebhookSecret = existing.WebhookSecret
		integration.AutoMoveInProgress = existing.AutoMoveInProgress
		integration.AutoMoveReview = existing.AutoMoveReview
		integration.AutoCloseOnMerge = existing.AutoCloseOnMerge
		integration.IsActive = existing.IsActive
		integration.AccessTokenEnc = existing.AccessTokenEnc
		integration.CreatedAt = existing.CreatedAt
	} else {
		secret, err := generateSecret()
		if err != nil {
			return nil, err
		}
		integration.WebhookSecret = secret
	}

	applyBool(&integration.AutoMoveInProgress, req.AutoMoveInProgress)
	applyBool(&integration.AutoMoveReview, req.AutoMoveReview)
	applyBool(&integration.AutoCloseOnMerge, req.AutoCloseOnMerge)
	applyBool(&integration.IsActive, req.IsActive)

	if token := strings.TrimSpace(req.AccessToken); token != "" {
		if s.sealer == nil {
			return nil, errors.New("token encryption is not available")
		}
		sealed, err := s.sealer.Seal(token)
		if err != nil {
			return nil, err
		}
		integration.AccessTokenEnc = sealed
	}

	err = s.txManager.WithinTx(ctx, func(r *repo.RepositoryRegistry) error {
		return r.GitLab().UpsertIntegration(ctx, integration)
	})
	if err != nil {
		return nil, err
	}

	saved, err := s.repo.GetIntegration(ctx, projectID)
	if err != nil || saved == nil {
		return nil, err
	}
	return s.toResponse(saved), nil
}

func (s *GitLabService) DeleteIntegration(ctx context.Context, projectID int64) error {
	return s.txManager.WithinTx(ctx, func(r *repo.RepositoryRegistry) error {
		return r.GitLab().DeleteIntegration(ctx, projectID)
	})
}

func (s *GitLabService) TestConnection(ctx context.Context, projectID int64) (*gitlab.Project, error) {
	integration, err := s.requireIntegration(ctx, projectID)
	if err != nil {
		return nil, err
	}

	client, err := s.newClient(integration)
	if err != nil {
		return nil, err
	}

	return client.GetProject(ctx, integration.GitLabProjectID)
}

func (s *GitLabService) GetTaskGit(ctx context.Context, taskID int64) (dto.TaskGitOverview, error) {
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil || task == nil {
		return dto.TaskGitOverview{}, ErrNotFound
	}

	integration, err := s.repo.GetIntegration(ctx, task.ProjectID)
	if err != nil {
		return dto.TaskGitOverview{}, err
	}

	links, err := s.repo.ListLinks(ctx, taskID)
	if err != nil {
		return dto.TaskGitOverview{}, err
	}

	pipelines, err := s.repo.ListPipelines(ctx, taskID)
	if err != nil {
		return dto.TaskGitOverview{}, err
	}

	overview := dto.TaskGitOverview{
		Links:     links,
		Pipelines: pipelines,
	}

	if integration != nil && integration.IsActive {
		overview.Connected = true
		overview.TaskKey = BuildTaskKey(integration.TaskKeyPrefix, taskID)
		overview.SuggestedBranch = BuildBranchName(integration.TaskKeyPrefix, taskID, task.Title)
		overview.RepositoryURL = fmt.Sprintf("%s/projects/%d", integration.BaseURL, integration.GitLabProjectID)
	}

	return overview, nil
}

func (s *GitLabService) ProjectSummary(ctx context.Context, projectID int64) ([]dto.ProjectGitSummaryItem, error) {
	return s.repo.ProjectSummary(ctx, projectID)
}

func (s *GitLabService) CreateBranch(ctx context.Context, taskID int64) (dto.TaskGitLinkDTO, error) {
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil || task == nil {
		return dto.TaskGitLinkDTO{}, ErrNotFound
	}

	integration, err := s.requireIntegration(ctx, task.ProjectID)
	if err != nil {
		return dto.TaskGitLinkDTO{}, err
	}

	client, err := s.newClient(integration)
	if err != nil {
		return dto.TaskGitLinkDTO{}, err
	}

	branchName := BuildBranchName(integration.TaskKeyPrefix, taskID, task.Title)
	branch, err := client.CreateBranch(ctx, integration.GitLabProjectID, branchName, integration.DefaultBranch)
	if err != nil {
		return dto.TaskGitLinkDTO{}, err
	}

	state := "created"
	link := dto.TaskGitLinkDTO{
		TaskID:     taskID,
		Kind:       dto.GitLinkBranch,
		ExternalID: branch.Name,
		Title:      &branch.Name,
		State:      &state,
		UpdatedAt:  time.Now(),
	}
	if branch.WebURL != "" {
		link.WebURL = &branch.WebURL
	}

	if err := s.txManager.WithinTx(ctx, func(r *repo.RepositoryRegistry) error {
		return r.GitLab().UpsertLink(ctx, &link)
	}); err != nil {
		return dto.TaskGitLinkDTO{}, err
	}

	if s.broadcaster != nil {
		s.broadcaster.Broadcast("gitlab:link", link)
	}

	return link, nil
}

func (s *GitLabService) RetryPipeline(ctx context.Context, taskID, pipelineID int64) (dto.GitLabPipelineDTO, error) {
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil || task == nil {
		return dto.GitLabPipelineDTO{}, ErrNotFound
	}

	integration, err := s.requireIntegration(ctx, task.ProjectID)
	if err != nil {
		return dto.GitLabPipelineDTO{}, err
	}

	client, err := s.newClient(integration)
	if err != nil {
		return dto.GitLabPipelineDTO{}, err
	}

	pipeline, err := client.RetryPipeline(ctx, integration.GitLabProjectID, pipelineID)
	if err != nil {
		return dto.GitLabPipelineDTO{}, err
	}

	record := dto.GitLabPipelineDTO{
		ProjectID:  task.ProjectID,
		TaskID:     &taskID,
		PipelineID: pipeline.ID,
		Ref:        pipeline.Ref,
		SHA:        pipeline.SHA,
		Status:     pipeline.Status,
		UpdatedAt:  time.Now(),
	}
	if pipeline.WebURL != "" {
		record.WebURL = &pipeline.WebURL
	}

	if err := s.txManager.WithinTx(ctx, func(r *repo.RepositoryRegistry) error {
		return r.GitLab().UpsertPipeline(ctx, &record)
	}); err != nil {
		return dto.GitLabPipelineDTO{}, err
	}

	if s.broadcaster != nil {
		s.broadcaster.Broadcast("gitlab:pipeline", record)
	}

	return record, nil
}

func (s *GitLabService) requireIntegration(ctx context.Context, projectID int64) (*dto.GitLabIntegration, error) {
	integration, err := s.repo.GetIntegration(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if integration == nil || !integration.IsActive {
		return nil, ErrIntegrationNotConfigured
	}
	return integration, nil
}

func (s *GitLabService) newClient(integration *dto.GitLabIntegration) (*gitlab.Client, error) {
	if len(integration.AccessTokenEnc) == 0 || s.sealer == nil {
		return nil, ErrIntegrationNoToken
	}

	token, err := s.sealer.Open(integration.AccessTokenEnc)
	if err != nil {
		return nil, err
	}

	return gitlab.NewClient(integration.BaseURL, token), nil
}

func (s *GitLabService) toResponse(integration *dto.GitLabIntegration) *dto.GitLabIntegrationResponse {
	return &dto.GitLabIntegrationResponse{
		GitLabIntegration: *integration,
		TokenSet:          len(integration.AccessTokenEnc) > 0,
		WebhookSecret:     integration.WebhookSecret,
		WebhookURL:        s.publicURL + "/api/v1/gitlab/webhook/" + strconv.FormatInt(integration.ProjectID, 10),
	}
}

func applyBool(target *bool, value *bool) {
	if value != nil {
		*target = *value
	}
}

func generateSecret() (string, error) {
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func (s *GitLabService) CommentOnMergeRequest(ctx context.Context, projectID, mergeRequestIID int64, body string) error {
	integration, err := s.requireIntegration(ctx, projectID)
	if err != nil {
		return err
	}

	client, err := s.newClient(integration)
	if err != nil {
		return err
	}

	return client.CreateMergeRequestNote(ctx, integration.GitLabProjectID, mergeRequestIID, body)
}
