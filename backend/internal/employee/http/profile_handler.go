package http

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/employee/dto"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/employee/http/middleware"
)

type ProfileService interface {
	CreateProfile(ctx context.Context, req dto.CreateProfileRequest) (dto.ProfileDTO, error)
	GetProfile(ctx context.Context, userID int64) (dto.ProfileDTO, error)
	ListProfiles(ctx context.Context, filter dto.ListProfilesFilter) (dto.ListProfilesResponse, error)
	UpdateProfile(ctx context.Context, userID int64, req dto.UpdateProfileRequest) (dto.ProfileDTO, error)
	ChangeUserStatus(ctx context.Context, userID int64, isActive bool) error
	AddSkill(ctx context.Context, employeeID, skillID int64) error
	RemoveSkill(ctx context.Context, employeeID, skillID int64) error
}

type TokenValidator interface {
	ValidateToken(tokenStr string) (int64, string, error)
}

type ProfileHandler struct {
	service       ProfileService
	authValidator TokenValidator
}

func NewProfileHandler(mux *http.ServeMux, service ProfileService, authValidator TokenValidator) *ProfileHandler {
	h := &ProfileHandler{
		service:       service,
		authValidator: authValidator,
	}
	h.registerRoutes(mux)
	return h
}

func (h *ProfileHandler) registerRoutes(mux *http.ServeMux) {
	authMW := middleware.Auth(h.authValidator)

	mux.Handle("GET /api/v1/profiles", authMW(middleware.Chaos(http.HandlerFunc(h.ListProfiles))))
	mux.Handle("GET /api/v1/profiles/{id}", authMW(middleware.Chaos(http.HandlerFunc(h.GetProfile))))
	mux.Handle("POST /api/v1/profiles", authMW(middleware.RequireRoles("admin", "hr", "developer")(middleware.Chaos(http.HandlerFunc(h.CreateProfile)))))
	mux.Handle("PUT /api/v1/profiles/{id}", authMW(middleware.Chaos(http.HandlerFunc(h.UpdateProfile))))
	mux.Handle("DELETE /api/v1/profiles/{id}", authMW(middleware.RequireRoles("admin", "developer")(middleware.Chaos(http.HandlerFunc(h.DeleteProfile)))))

	mux.Handle("POST /api/v1/profiles/{id}/skills", authMW(middleware.Chaos(http.HandlerFunc(h.AddSkill))))
	mux.Handle("DELETE /api/v1/profiles/{id}/skills/{skillId}", authMW(middleware.Chaos(http.HandlerFunc(h.RemoveSkill))))
}

func (h *ProfileHandler) ListProfiles(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	pageSize, _ := strconv.Atoi(query.Get("page_size"))
	pageNumber, _ := strconv.Atoi(query.Get("page_number"))
	deptID, _ := strconv.ParseInt(query.Get("department_id"), 10, 64)
	posID, _ := strconv.ParseInt(query.Get("position_id"), 10, 64)

	filter := dto.ListProfilesFilter{
		PageSize:     pageSize,
		PageNumber:   pageNumber,
		DepartmentID: deptID,
		PositionID:   posID,
	}

	res, err := h.service.ListProfiles(r.Context(), filter)
	if err != nil {
		log.Println("list profiles error:", err)
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(res)
}

func (h *ProfileHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	param := r.PathValue("id")
	id, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		http.Error(w, `{"error":"invalid profile id"}`, http.StatusBadRequest)
		return
	}

	res, err := h.service.GetProfile(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"profile not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(res)
}

func (h *ProfileHandler) CreateProfile(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println("create profile decode error:", err)
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}

	res, err := h.service.CreateProfile(r.Context(), req)
	if err != nil {
		log.Println("create profile error:", err)
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(res)
}

func (h *ProfileHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	param := r.PathValue("id")
	id, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		http.Error(w, `{"error":"invalid profile id"}`, http.StatusBadRequest)
		return
	}

	var req dto.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}

	res, err := h.service.UpdateProfile(r.Context(), id, req)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(res)
}

func (h *ProfileHandler) DeleteProfile(w http.ResponseWriter, r *http.Request) {
	param := r.PathValue("id")
	id, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		http.Error(w, `{"error":"invalid profile id"}`, http.StatusBadRequest)
		return
	}

	if err := h.service.ChangeUserStatus(r.Context(), id, false); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "profile deactivated"})
}

func (h *ProfileHandler) AddSkill(w http.ResponseWriter, r *http.Request) {
	param := r.PathValue("id")
	id, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		http.Error(w, `{"error":"invalid profile id"}`, http.StatusBadRequest)
		return
	}

	var req struct {
		SkillID int64 `json:"skill_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.SkillID == 0 {
		http.Error(w, `{"error":"skill_id is required"}`, http.StatusBadRequest)
		return
	}

	if err := h.service.AddSkill(r.Context(), id, req.SkillID); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "skill added"})
}

func (h *ProfileHandler) RemoveSkill(w http.ResponseWriter, r *http.Request) {
	param := r.PathValue("id")
	id, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		http.Error(w, `{"error":"invalid profile id"}`, http.StatusBadRequest)
		return
	}

	skillParam := r.PathValue("skillId")
	skillID, err := strconv.ParseInt(skillParam, 10, 64)
	if err != nil {
		http.Error(w, `{"error":"invalid skill id"}`, http.StatusBadRequest)
		return
	}

	if err := h.service.RemoveSkill(r.Context(), id, skillID); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "skill removed"})
}
