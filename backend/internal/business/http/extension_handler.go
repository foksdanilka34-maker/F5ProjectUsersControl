package http

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/business/core"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/business/dto"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/business/http/middleware"
)

const maxPropertyBodySize = 64 << 10

type ExtensionService interface {
	Register(ctx context.Context, req dto.SaveExtensionRequest) (*dto.ExtensionDTO, error)
	List(ctx context.Context) ([]dto.ExtensionDTO, error)
	Delete(ctx context.Context, key string) error

	SetInstalled(ctx context.Context, projectID int64, key string, enabled bool) error
	Uninstall(ctx context.Context, projectID int64, key string) error
	ListForProject(ctx context.Context, projectID int64) ([]dto.ProjectExtensionDTO, error)

	AuthenticateExtension(ctx context.Context, key, secret string) (*dto.ExtensionDTO, error)
	SetTaskProperty(ctx context.Context, extension *dto.ExtensionDTO, taskID int64, key string, value []byte) error
	GetTaskProperty(ctx context.Context, extension *dto.ExtensionDTO, taskID int64, key string) ([]byte, error)
	ListTaskProperties(ctx context.Context, extension *dto.ExtensionDTO, taskID int64) ([]dto.TaskEntityPropertyDTO, error)
}

type ExtensionHandler struct {
	service       ExtensionService
	authValidator *middleware.JWTValidator
}

func NewExtensionHandler(mux *http.ServeMux, service ExtensionService, authValidator *middleware.JWTValidator) *ExtensionHandler {
	h := &ExtensionHandler{
		service:       service,
		authValidator: authValidator,
	}
	h.registerRoutes(mux)
	return h
}

func (h *ExtensionHandler) registerRoutes(mux *http.ServeMux) {
	authMW := middleware.Auth(h.authValidator)
	adminRole := middleware.RequireRoles("admin", "director")
	managerRole := middleware.RequireRoles("admin", "manager", "developer", "director")

	mux.Handle("GET /api/v1/extensions", authMW(http.HandlerFunc(h.ListRegistry)))
	mux.Handle("POST /api/v1/extensions", authMW(adminRole(http.HandlerFunc(h.Register))))
	mux.Handle("DELETE /api/v1/extensions/{key}", authMW(adminRole(http.HandlerFunc(h.Delete))))

	mux.Handle("GET /api/v1/extensions/projects/{id}", authMW(http.HandlerFunc(h.ListForProject)))
	mux.Handle("POST /api/v1/extensions/projects/{id}/{key}/toggle", authMW(managerRole(http.HandlerFunc(h.Toggle))))
	mux.Handle("DELETE /api/v1/extensions/projects/{id}/{key}", authMW(managerRole(http.HandlerFunc(h.Uninstall))))

	mux.Handle("GET /api/v1/extensions/properties/tasks/{taskId}", http.HandlerFunc(h.ListProperties))
	mux.Handle("GET /api/v1/extensions/properties/tasks/{taskId}/{key}", http.HandlerFunc(h.GetProperty))
	mux.Handle("PUT /api/v1/extensions/properties/tasks/{taskId}/{key}", http.HandlerFunc(h.SetProperty))
}

func (h *ExtensionHandler) ListRegistry(w http.ResponseWriter, r *http.Request) {
	extensions, err := h.service.List(r.Context())
	if err != nil {
		writeExtensionError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"extensions": extensions})
}

func (h *ExtensionHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.SaveExtensionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}

	extension, err := h.service.Register(r.Context(), req)
	if err != nil {
		writeExtensionError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, extension)
}

func (h *ExtensionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	if err := h.service.Delete(r.Context(), key); err != nil {
		writeExtensionError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "extension removed"})
}

func (h *ExtensionHandler) ListForProject(w http.ResponseWriter, r *http.Request) {
	projectID, ok := pathInt64(w, r, "id", "invalid project id")
	if !ok {
		return
	}

	extensions, err := h.service.ListForProject(r.Context(), projectID)
	if err != nil {
		writeExtensionError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"extensions": extensions})
}

func (h *ExtensionHandler) Toggle(w http.ResponseWriter, r *http.Request) {
	projectID, ok := pathInt64(w, r, "id", "invalid project id")
	if !ok {
		return
	}
	key := r.PathValue("key")

	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}

	if err := h.service.SetInstalled(r.Context(), projectID, key, req.Enabled); err != nil {
		writeExtensionError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "extension updated"})
}

func (h *ExtensionHandler) Uninstall(w http.ResponseWriter, r *http.Request) {
	projectID, ok := pathInt64(w, r, "id", "invalid project id")
	if !ok {
		return
	}
	key := r.PathValue("key")

	if err := h.service.Uninstall(r.Context(), projectID, key); err != nil {
		writeExtensionError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "extension uninstalled"})
}

func (h *ExtensionHandler) authenticateRequest(w http.ResponseWriter, r *http.Request) *dto.ExtensionDTO {
	key := r.Header.Get("X-Extension-Key")
	secret := r.Header.Get("X-Extension-Secret")
	if key == "" || secret == "" {
		http.Error(w, `{"error":"missing extension credentials"}`, http.StatusUnauthorized)
		return nil
	}

	extension, err := h.service.AuthenticateExtension(r.Context(), key, secret)
	if err != nil {
		writeExtensionError(w, err)
		return nil
	}
	return extension
}

func (h *ExtensionHandler) SetProperty(w http.ResponseWriter, r *http.Request) {
	extension := h.authenticateRequest(w, r)
	if extension == nil {
		return
	}

	taskID, ok := pathInt64(w, r, "taskId", "invalid task id")
	if !ok {
		return
	}
	key := r.PathValue("key")

	value, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxPropertyBodySize))
	if err != nil {
		http.Error(w, `{"error":"failed to read body"}`, http.StatusBadRequest)
		return
	}
	if !json.Valid(value) {
		http.Error(w, `{"error":"value must be valid json"}`, http.StatusBadRequest)
		return
	}

	if err := h.service.SetTaskProperty(r.Context(), extension, taskID, key, value); err != nil {
		writeExtensionError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "property saved"})
}

func (h *ExtensionHandler) GetProperty(w http.ResponseWriter, r *http.Request) {
	extension := h.authenticateRequest(w, r)
	if extension == nil {
		return
	}

	taskID, ok := pathInt64(w, r, "taskId", "invalid task id")
	if !ok {
		return
	}
	key := r.PathValue("key")

	value, err := h.service.GetTaskProperty(r.Context(), extension, taskID, key)
	if err != nil {
		writeExtensionError(w, err)
		return
	}
	if value == nil {
		http.Error(w, `{"error":"property not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(value)
}

func (h *ExtensionHandler) ListProperties(w http.ResponseWriter, r *http.Request) {
	extension := h.authenticateRequest(w, r)
	if extension == nil {
		return
	}

	taskID, ok := pathInt64(w, r, "taskId", "invalid task id")
	if !ok {
		return
	}

	properties, err := h.service.ListTaskProperties(r.Context(), extension, taskID)
	if err != nil {
		writeExtensionError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"properties": properties})
}

func writeExtensionError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, core.ErrNotFound):
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
	case errors.Is(err, core.ErrExtensionExists):
		http.Error(w, `{"error":"Расширение с таким ключом уже зарегистрировано"}`, http.StatusConflict)
	case errors.Is(err, core.ErrExtensionNotInstalled):
		http.Error(w, `{"error":"Расширение не установлено для этого проекта"}`, http.StatusForbidden)
	case errors.Is(err, core.ErrInvalidExtensionAuth):
		http.Error(w, `{"error":"invalid extension credentials"}`, http.StatusUnauthorized)
	default:
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
	}
}
