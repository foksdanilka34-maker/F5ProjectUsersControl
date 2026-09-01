package http

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/business/dto"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/business/http/middleware"
)

type AnalyticsService interface {
	GetDashboard(ctx context.Context) (dto.DashboardStatsDTO, error)
	GetProjectMetrics(ctx context.Context, projectID int64) (dto.ProjectMetricsDTO, error)
	ListAllProjectMetrics(ctx context.Context) ([]dto.ProjectMetricsDTO, error)
	GetEmployeeMetrics(ctx context.Context, userID int64) (dto.EmployeeMetricsDTO, error)
	ListAllEmployeeMetrics(ctx context.Context) ([]dto.EmployeeMetricsDTO, error)
	GetProductivityTrends(ctx context.Context, days int) (dto.ProductivityTrendsDTO, error)
}

type AnalyticsHandler struct {
	service       AnalyticsService
	authValidator *middleware.JWTValidator
}

func NewAnalyticsHandler(mux *http.ServeMux, service AnalyticsService, authValidator *middleware.JWTValidator) *AnalyticsHandler {
	h := &AnalyticsHandler{
		service:       service,
		authValidator: authValidator,
	}
	h.registerRoutes(mux)
	return h
}

func (h *AnalyticsHandler) registerRoutes(mux *http.ServeMux) {
	authMW := middleware.Auth(h.authValidator)

	mux.Handle("GET /api/v1/analytics/dashboard", authMW(middleware.Chaos(http.HandlerFunc(h.GetDashboard))))
	mux.Handle("GET /api/v1/analytics/summary", authMW(middleware.Chaos(http.HandlerFunc(h.GetDashboard))))
	mux.Handle("GET /api/v1/analytics/project/{id}", authMW(middleware.Chaos(http.HandlerFunc(h.GetProjectMetrics))))
	mux.Handle("GET /api/v1/analytics/projects", authMW(middleware.Chaos(http.HandlerFunc(h.ListAllProjectMetrics))))
	mux.Handle("GET /api/v1/analytics/employee/{id}", authMW(middleware.Chaos(http.HandlerFunc(h.GetEmployeeMetrics))))
	mux.Handle("GET /api/v1/analytics/employees", authMW(middleware.Chaos(http.HandlerFunc(h.ListAllEmployeeMetrics))))
	mux.Handle("GET /api/v1/analytics/trends", authMW(middleware.Chaos(http.HandlerFunc(h.GetTrends))))
}

func (h *AnalyticsHandler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	stats, err := h.service.GetDashboard(r.Context())
	if err != nil {
		log.Println("get dashboard error:", err)
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(stats)
}

func (h *AnalyticsHandler) GetProjectMetrics(w http.ResponseWriter, r *http.Request) {
	param := r.PathValue("id")
	id, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		http.Error(w, `{"error":"invalid project id"}`, http.StatusBadRequest)
		return
	}

	metrics, err := h.service.GetProjectMetrics(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(metrics)
}

func (h *AnalyticsHandler) ListAllProjectMetrics(w http.ResponseWriter, r *http.Request) {
	list, err := h.service.ListAllProjectMetrics(r.Context())
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"metrics":     list,
		"total_count": len(list),
	})
}

func (h *AnalyticsHandler) GetEmployeeMetrics(w http.ResponseWriter, r *http.Request) {
	param := r.PathValue("id")
	id, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		http.Error(w, `{"error":"invalid employee id"}`, http.StatusBadRequest)
		return
	}

	metrics, err := h.service.GetEmployeeMetrics(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(metrics)
}

func (h *AnalyticsHandler) ListAllEmployeeMetrics(w http.ResponseWriter, r *http.Request) {
	list, err := h.service.ListAllEmployeeMetrics(r.Context())
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"metrics":     list,
		"total_count": len(list),
	})
}

func (h *AnalyticsHandler) GetTrends(w http.ResponseWriter, r *http.Request) {
	days, _ := strconv.Atoi(r.URL.Query().Get("days"))
	if days <= 0 {
		days = 30
	}

	trends, err := h.service.GetProductivityTrends(r.Context(), days)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(trends)
}
