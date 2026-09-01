package http

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/employee/dto"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/employee/http/middleware"
)

type OrgService interface {
	CreateDepartment(ctx context.Context, name string) (dto.DepartmentDTO, error)
	GetDepartment(ctx context.Context, id int64) (dto.DepartmentDTO, error)
	ListDepartments(ctx context.Context) ([]dto.DepartmentDTO, error)
	UpdateDepartment(ctx context.Context, id int64, name string) (dto.DepartmentDTO, error)
	DeleteDepartment(ctx context.Context, id int64) error

	CreatePosition(ctx context.Context, name string) (dto.PositionDTO, error)
	GetPosition(ctx context.Context, id int64) (dto.PositionDTO, error)
	ListPositions(ctx context.Context) ([]dto.PositionDTO, error)
	UpdatePosition(ctx context.Context, id int64, name string) (dto.PositionDTO, error)
	DeletePosition(ctx context.Context, id int64) error

	CreateSkill(ctx context.Context, name string) (dto.SkillDTO, error)
	ListSkills(ctx context.Context) ([]dto.SkillDTO, error)
	DeleteSkill(ctx context.Context, id int64) error
}

type OrgHandler struct {
	service       OrgService
	authValidator TokenValidator
}

func NewOrgHandler(mux *http.ServeMux, service OrgService, authValidator TokenValidator) *OrgHandler {
	h := &OrgHandler{
		service:       service,
		authValidator: authValidator,
	}
	h.registerRoutes(mux)
	return h
}

func (h *OrgHandler) registerRoutes(mux *http.ServeMux) {
	authMW := middleware.Auth(h.authValidator)
	adminRole := middleware.RequireRoles("admin", "developer")

	// Departments
	mux.Handle("GET /api/v1/departments", authMW(middleware.Chaos(http.HandlerFunc(h.ListDepartments))))
	mux.Handle("GET /api/v1/departments/{id}", authMW(middleware.Chaos(http.HandlerFunc(h.GetDepartment))))
	mux.Handle("POST /api/v1/departments", authMW(adminRole(middleware.Chaos(http.HandlerFunc(h.CreateDepartment)))))
	mux.Handle("PUT /api/v1/departments/{id}", authMW(adminRole(middleware.Chaos(http.HandlerFunc(h.UpdateDepartment)))))
	mux.Handle("DELETE /api/v1/departments/{id}", authMW(adminRole(middleware.Chaos(http.HandlerFunc(h.DeleteDepartment)))))

	// Positions
	mux.Handle("GET /api/v1/positions", authMW(middleware.Chaos(http.HandlerFunc(h.ListPositions))))
	mux.Handle("GET /api/v1/positions/{id}", authMW(middleware.Chaos(http.HandlerFunc(h.GetPosition))))
	mux.Handle("POST /api/v1/positions", authMW(adminRole(middleware.Chaos(http.HandlerFunc(h.CreatePosition)))))
	mux.Handle("PUT /api/v1/positions/{id}", authMW(adminRole(middleware.Chaos(http.HandlerFunc(h.UpdatePosition)))))
	mux.Handle("DELETE /api/v1/positions/{id}", authMW(adminRole(middleware.Chaos(http.HandlerFunc(h.DeletePosition)))))

	// Skills
	mux.Handle("GET /api/v1/skills", authMW(middleware.Chaos(http.HandlerFunc(h.ListSkills))))
	mux.Handle("POST /api/v1/skills", authMW(adminRole(middleware.Chaos(http.HandlerFunc(h.CreateSkill)))))
	mux.Handle("DELETE /api/v1/skills/{id}", authMW(adminRole(middleware.Chaos(http.HandlerFunc(h.DeleteSkill)))))
}

// Departments
func (h *OrgHandler) ListDepartments(w http.ResponseWriter, r *http.Request) {
	list, err := h.service.ListDepartments(r.Context())
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"departments": list})
}

func (h *OrgHandler) GetDepartment(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}
	dept, err := h.service.GetDepartment(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(dept)
}

func (h *OrgHandler) CreateDepartment(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		http.Error(w, `{"error":"name is required"}`, http.StatusBadRequest)
		return
	}
	dept, err := h.service.CreateDepartment(r.Context(), req.Name)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(dept)
}

func (h *OrgHandler) UpdateDepartment(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		http.Error(w, `{"error":"name is required"}`, http.StatusBadRequest)
		return
	}
	dept, err := h.service.UpdateDepartment(r.Context(), id, req.Name)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(dept)
}

func (h *OrgHandler) DeleteDepartment(ctx http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(ctx, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}
	if err := h.service.DeleteDepartment(r.Context(), id); err != nil {
		http.Error(ctx, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	ctx.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(ctx).Encode(map[string]string{"message": "department deleted"})
}

// Positions
func (h *OrgHandler) ListPositions(w http.ResponseWriter, r *http.Request) {
	list, err := h.service.ListPositions(r.Context())
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"positions": list})
}

func (h *OrgHandler) GetPosition(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}
	pos, err := h.service.GetPosition(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(pos)
}

func (h *OrgHandler) CreatePosition(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		http.Error(w, `{"error":"name is required"}`, http.StatusBadRequest)
		return
	}
	pos, err := h.service.CreatePosition(r.Context(), req.Name)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(pos)
}

func (h *OrgHandler) UpdatePosition(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		http.Error(w, `{"error":"name is required"}`, http.StatusBadRequest)
		return
	}
	pos, err := h.service.UpdatePosition(r.Context(), id, req.Name)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(pos)
}

func (h *OrgHandler) DeletePosition(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}
	if err := h.service.DeletePosition(r.Context(), id); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "position deleted"})
}

// Skills
func (h *OrgHandler) ListSkills(w http.ResponseWriter, r *http.Request) {
	list, err := h.service.ListSkills(r.Context())
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"skills": list})
}

func (h *OrgHandler) CreateSkill(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		http.Error(w, `{"error":"name is required"}`, http.StatusBadRequest)
		return
	}
	skill, err := h.service.CreateSkill(r.Context(), req.Name)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(skill)
}

func (h *OrgHandler) DeleteSkill(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}
	if err := h.service.DeleteSkill(r.Context(), id); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}
