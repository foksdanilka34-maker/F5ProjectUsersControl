package http

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/business/core"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/business/dto"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/business/http/middleware"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/pkg/gitlab"
)

const maxWebhookBodySize = 1 << 20

type GitLabService interface {
	GetIntegration(ctx context.Context, projectID int64) (*dto.GitLabIntegrationResponse, error)
	SaveIntegration(ctx context.Context, projectID int64, req dto.SaveGitLabIntegrationRequest) (*dto.GitLabIntegrationResponse, error)
	DeleteIntegration(ctx context.Context, projectID int64) error
	TestConnection(ctx context.Context, projectID int64) (*gitlab.Project, error)
	GetTaskGit(ctx context.Context, taskID int64) (dto.TaskGitOverview, error)
	ProjectSummary(ctx context.Context, projectID int64) ([]dto.ProjectGitSummaryItem, error)
	CreateBranch(ctx context.Context, taskID int64) (dto.TaskGitLinkDTO, error)
	RetryPipeline(ctx context.Context, taskID, pipelineID int64) (dto.GitLabPipelineDTO, error)
}

type GitLabWebhookService interface {
	Accept(ctx context.Context, projectID int64, token, eventType, deliveryID string, payload []byte) error
}

type GitLabHandler struct {
	service       GitLabService
	webhooks      GitLabWebhookService
	authValidator *middleware.JWTValidator
}

func NewGitLabHandler(
	mux *http.ServeMux,
	service GitLabService,
	webhooks GitLabWebhookService,
	authValidator *middleware.JWTValidator,
) *GitLabHandler {
	h := &GitLabHandler{
		service:       service,
		webhooks:      webhooks,
		authValidator: authValidator,
	}
	h.registerRoutes(mux)
	return h
}

func (h *GitLabHandler) registerRoutes(mux *http.ServeMux) {
	authMW := middleware.Auth(h.authValidator)
	managerRole := middleware.RequireRoles("admin", "manager", "director")

	mux.Handle("POST /api/v1/gitlab/webhook/{projectId}", http.HandlerFunc(h.ReceiveWebhook))

	mux.Handle("GET /api/v1/gitlab/projects/{id}", authMW(http.HandlerFunc(h.GetIntegration)))
	mux.Handle("PUT /api/v1/gitlab/projects/{id}", authMW(managerRole(http.HandlerFunc(h.SaveIntegration))))
	mux.Handle("DELETE /api/v1/gitlab/projects/{id}", authMW(managerRole(http.HandlerFunc(h.DeleteIntegration))))
	mux.Handle("POST /api/v1/gitlab/projects/{id}/test", authMW(managerRole(http.HandlerFunc(h.TestConnection))))
	mux.Handle("GET /api/v1/gitlab/projects/{id}/summary", authMW(http.HandlerFunc(h.ProjectSummary)))

	mux.Handle("GET /api/v1/gitlab/tasks/{id}", authMW(http.HandlerFunc(h.GetTaskGit)))
	mux.Handle("POST /api/v1/gitlab/tasks/{id}/branch", authMW(http.HandlerFunc(h.CreateBranch)))
	mux.Handle("POST /api/v1/gitlab/tasks/{id}/pipelines/{pipelineId}/retry", authMW(http.HandlerFunc(h.RetryPipeline)))
}

func (h *GitLabHandler) ReceiveWebhook(w http.ResponseWriter, r *http.Request) {
	projectID, err := strconv.ParseInt(r.PathValue("projectId"), 10, 64)
	if err != nil {
		http.Error(w, `{"error":"invalid project id"}`, http.StatusBadRequest)
		return
	}

	payload, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxWebhookBodySize))
	if err != nil {
		http.Error(w, `{"error":"failed to read payload"}`, http.StatusBadRequest)
		return
	}

	eventType := r.Header.Get("X-Gitlab-Event")
	var probe struct {
		ObjectKind string `json:"object_kind"`
	}
	if err := json.Unmarshal(payload, &probe); err == nil && probe.ObjectKind != "" {
		eventType = probe.ObjectKind
	}

	err = h.webhooks.Accept(r.Context(), projectID,
		r.Header.Get("X-Gitlab-Token"), eventType, r.Header.Get("X-Gitlab-Event-UUID"), payload)

	switch {
	case errors.Is(err, core.ErrIntegrationNotConfigured):
		http.Error(w, `{"error":"integration is not configured"}`, http.StatusNotFound)
		return
	case errors.Is(err, core.ErrInvalidWebhookToken):
		http.Error(w, `{"error":"invalid webhook token"}`, http.StatusUnauthorized)
		return
	case err != nil:
		log.Println("gitlab webhook error:", err)
		http.Error(w, `{"error":"failed to accept webhook"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "accepted"})
}

func (h *GitLabHandler) GetIntegration(w http.ResponseWriter, r *http.Request) {
	projectID, ok := pathInt64(w, r, "id", "invalid project id")
	if !ok {
		return
	}

	integration, err := h.service.GetIntegration(r.Context(), projectID)
	if err != nil {
		writeGitLabError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"connected":   integration != nil,
		"integration": integration,
	})
}

func (h *GitLabHandler) SaveIntegration(w http.ResponseWriter, r *http.Request) {
	projectID, ok := pathInt64(w, r, "id", "invalid project id")
	if !ok {
		return
	}

	var req dto.SaveGitLabIntegrationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}

	integration, err := h.service.SaveIntegration(r.Context(), projectID, req)
	if err != nil {
		writeGitLabError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"connected":   true,
		"integration": integration,
	})
}

func (h *GitLabHandler) DeleteIntegration(w http.ResponseWriter, r *http.Request) {
	projectID, ok := pathInt64(w, r, "id", "invalid project id")
	if !ok {
		return
	}

	if err := h.service.DeleteIntegration(r.Context(), projectID); err != nil {
		writeGitLabError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "integration removed"})
}

func (h *GitLabHandler) TestConnection(w http.ResponseWriter, r *http.Request) {
	projectID, ok := pathInt64(w, r, "id", "invalid project id")
	if !ok {
		return
	}

	project, err := h.service.TestConnection(r.Context(), projectID)
	if err != nil {
		writeGitLabError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"project": project})
}

func (h *GitLabHandler) ProjectSummary(w http.ResponseWriter, r *http.Request) {
	projectID, ok := pathInt64(w, r, "id", "invalid project id")
	if !ok {
		return
	}

	items, err := h.service.ProjectSummary(r.Context(), projectID)
	if err != nil {
		writeGitLabError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *GitLabHandler) GetTaskGit(w http.ResponseWriter, r *http.Request) {
	taskID, ok := pathInt64(w, r, "id", "invalid task id")
	if !ok {
		return
	}

	overview, err := h.service.GetTaskGit(r.Context(), taskID)
	if err != nil {
		writeGitLabError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, overview)
}

func (h *GitLabHandler) CreateBranch(w http.ResponseWriter, r *http.Request) {
	taskID, ok := pathInt64(w, r, "id", "invalid task id")
	if !ok {
		return
	}

	link, err := h.service.CreateBranch(r.Context(), taskID)
	if err != nil {
		writeGitLabError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, link)
}

func (h *GitLabHandler) RetryPipeline(w http.ResponseWriter, r *http.Request) {
	taskID, ok := pathInt64(w, r, "id", "invalid task id")
	if !ok {
		return
	}

	pipelineID, ok := pathInt64(w, r, "pipelineId", "invalid pipeline id")
	if !ok {
		return
	}

	pipeline, err := h.service.RetryPipeline(r.Context(), taskID, pipelineID)
	if err != nil {
		writeGitLabError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, pipeline)
}

func pathInt64(w http.ResponseWriter, r *http.Request, param, message string) (int64, bool) {
	value, err := strconv.ParseInt(r.PathValue(param), 10, 64)
	if err != nil {
		http.Error(w, `{"error":"`+message+`"}`, http.StatusBadRequest)
		return 0, false
	}
	return value, true
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeGitLabError(w http.ResponseWriter, err error) {
	var apiErr *gitlab.APIError

	switch {
	case errors.Is(err, core.ErrNotFound):
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
	case errors.Is(err, core.ErrIntegrationNotConfigured):
		http.Error(w, `{"error":"Интеграция с GitLab не настроена"}`, http.StatusConflict)
	case errors.Is(err, core.ErrIntegrationNoToken):
		http.Error(w, `{"error":"Не задан access token GitLab"}`, http.StatusConflict)
	case errors.As(err, &apiErr):
		log.Println("gitlab api error:", err)
		writeJSON(w, http.StatusBadGateway, map[string]any{
			"error":       "GitLab отклонил запрос",
			"status_code": apiErr.StatusCode,
		})
	default:
		log.Println("gitlab handler error:", err)
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
	}
}
