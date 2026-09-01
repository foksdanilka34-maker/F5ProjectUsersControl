package http

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/business/dto"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/business/http/middleware"
)

type ProjectService interface {
	CreateProject(ctx context.Context, req dto.CreateProjectRequest) (dto.ProjectDTO, error)
	GetProject(ctx context.Context, id int64) (dto.ProjectDTO, error)
	ListProjects(ctx context.Context, filter dto.ListProjectsFilter) (dto.ListProjectsResponse, error)
	UpdateProject(ctx context.Context, id int64, req dto.UpdateProjectRequest) (dto.ProjectDTO, error)
	DeleteProject(ctx context.Context, id int64) error

	AddMember(ctx context.Context, projectID, userID int64, role string) error
	RemoveMember(ctx context.Context, projectID, userID int64) error
	ListMembers(ctx context.Context, projectID int64) ([]dto.ProjectMemberDTO, error)
}

type ProjectHandler struct {
	service       ProjectService
	authValidator *middleware.JWTValidator
}

func NewProjectHandler(mux *http.ServeMux, service ProjectService, authValidator *middleware.JWTValidator) *ProjectHandler {
	h := &ProjectHandler{
		service:       service,
		authValidator: authValidator,
	}
	h.registerRoutes(mux)
	return h
}

func (h *ProjectHandler) registerRoutes(mux *http.ServeMux) {
	authMW := middleware.Auth(h.authValidator)
	managerRole := middleware.RequireRoles("admin", "manager", "developer", "director")

	mux.Handle("GET /api/v1/projects", authMW(middleware.Chaos(http.HandlerFunc(h.ListProjects))))
	mux.Handle("GET /api/v1/projects/{id}", authMW(middleware.Chaos(http.HandlerFunc(h.GetProject))))
	mux.Handle("POST /api/v1/projects", authMW(managerRole(middleware.Chaos(http.HandlerFunc(h.CreateProject)))))
	mux.Handle("PUT /api/v1/projects/{id}", authMW(managerRole(middleware.Chaos(http.HandlerFunc(h.UpdateProject)))))
	mux.Handle("DELETE /api/v1/projects/{id}", authMW(managerRole(middleware.Chaos(http.HandlerFunc(h.DeleteProject)))))

	// Members
	mux.Handle("GET /api/v1/projects/{id}/members", authMW(middleware.Chaos(http.HandlerFunc(h.ListMembers))))
	mux.Handle("POST /api/v1/projects/{id}/members", authMW(managerRole(middleware.Chaos(http.HandlerFunc(h.AddMember)))))
	mux.Handle("DELETE /api/v1/projects/{id}/members/{userId}", authMW(managerRole(middleware.Chaos(http.HandlerFunc(h.RemoveMember)))))
}

func (h *ProjectHandler) ListProjects(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	pageSize, _ := strconv.Atoi(query.Get("page_size"))
	pageNumber, _ := strconv.Atoi(query.Get("page_number"))
	managerID, _ := strconv.ParseInt(query.Get("manager_id"), 10, 64)
	memberID, _ := strconv.ParseInt(query.Get("member_id"), 10, 64)
	status := query.Get("status")

	filter := dto.ListProjectsFilter{
		PageSize:   pageSize,
		PageNumber: pageNumber,
		ManagerID:  managerID,
		MemberID:   memberID,
		Status:     status,
	}

	res, err := h.service.ListProjects(r.Context(), filter)
	if err != nil {
		log.Println("list projects error:", err)
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(res)
}

func (h *ProjectHandler) GetProject(w http.ResponseWriter, r *http.Request) {
	param := r.PathValue("id")
	id, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		http.Error(w, `{"error":"invalid project id"}`, http.StatusBadRequest)
		return
	}

	res, err := h.service.GetProject(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"project not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(res)
}

func (h *ProjectHandler) CreateProject(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}

	if req.ManagerID == 0 {
		userID, _ := r.Context().Value(middleware.UserIDKey).(int64)
		req.ManagerID = userID
	}

	res, err := h.service.CreateProject(r.Context(), req)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "already exists") {
			http.Error(w, `{"error":"Проект с таким именем уже существует"}`, http.StatusConflict)
			return
		}
		http.Error(w, `{"error":"`+errMsg+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(res)
}

func (h *ProjectHandler) UpdateProject(w http.ResponseWriter, r *http.Request) {
	param := r.PathValue("id")
	id, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		http.Error(w, `{"error":"invalid project id"}`, http.StatusBadRequest)
		return
	}

	var req dto.UpdateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}

	res, err := h.service.UpdateProject(r.Context(), id, req)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(res)
}

func (h *ProjectHandler) DeleteProject(w http.ResponseWriter, r *http.Request) {
	param := r.PathValue("id")
	id, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		http.Error(w, `{"error":"invalid project id"}`, http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteProject(r.Context(), id); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "project deleted"})
}

func (h *ProjectHandler) ListMembers(w http.ResponseWriter, r *http.Request) {
	param := r.PathValue("id")
	id, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		http.Error(w, `{"error":"invalid project id"}`, http.StatusBadRequest)
		return
	}

	members, err := h.service.ListMembers(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"members": members})
}

func (h *ProjectHandler) AddMember(w http.ResponseWriter, r *http.Request) {
	param := r.PathValue("id")
	projectID, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		http.Error(w, `{"error":"invalid project id"}`, http.StatusBadRequest)
		return
	}

	var req struct {
		UserID int64  `json:"user_id"`
		Role   string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.UserID == 0 {
		http.Error(w, `{"error":"user_id is required"}`, http.StatusBadRequest)
		return
	}

	if err := h.service.AddMember(r.Context(), projectID, req.UserID, req.Role); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "member added"})
}

func (h *ProjectHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	param := r.PathValue("id")
	projectID, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		http.Error(w, `{"error":"invalid project id"}`, http.StatusBadRequest)
		return
	}

	userParam := r.PathValue("userId")
	userID, err := strconv.ParseInt(userParam, 10, 64)
	if err != nil {
		http.Error(w, `{"error":"invalid user id"}`, http.StatusBadRequest)
		return
	}

	if err := h.service.RemoveMember(r.Context(), projectID, userID); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "member removed"})
}
