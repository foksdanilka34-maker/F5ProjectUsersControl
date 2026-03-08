package handlers

import (
	"context"
	"log"
	"net/http"

	pb "github.com/foksdanilka34-maker/F5ProjectUsersControl/gen/go/business"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AnalyticsHTTPHandler struct {
	client pb.BusinessServiceClient
}

func NewAnalyticsHTTPHandler(client pb.BusinessServiceClient) *AnalyticsHTTPHandler {
	return &AnalyticsHTTPHandler{client: client}
}

func (h *AnalyticsHTTPHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/analytics/summary", h.GetSummary)
	mux.HandleFunc("GET /api/v1/analytics/project/{id}", h.GetProjectAnalytics)
	mux.HandleFunc("GET /api/v1/analytics/employee/{id}", h.GetEmployeeAnalytics)
	mux.HandleFunc("GET /api/v1/analytics/dashboard", h.GetDashboard)
}

func (h *AnalyticsHTTPHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	resp, err := h.client.GetDashboardStats(r.Context(), &pb.GetDashboardStatsRequest{})
	if err != nil {
		log.Println("analytics summary error:", err)
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"total_employees":     resp.TotalEmployees,
		"active_employees":    resp.ActiveEmployees,
		"total_projects":      resp.TotalProjects,
		"active_projects":     resp.ActiveProjects,
		"total_tasks":         resp.TotalTasks,
		"completed_tasks":     resp.CompletedTasks,
		"overdue_tasks":       resp.OverdueTasks,
		"completed_on_time":   resp.CompletedOnTime,
		"completed_late":      resp.CompletedLate,
		"avg_completion_rate": resp.AvgCompletionRate,
		"avg_on_time_rate":    resp.AvgOnTimeRate,
	})
}

func (h *AnalyticsHTTPHandler) GetProjectAnalytics(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := parseID(idStr)
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}

	resp, err := h.client.GetProjectMetrics(r.Context(), &pb.GetProjectMetricsRequest{ProjectId: id})
	if err != nil {
		log.Println("project analytics error:", err)
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	m := resp.Metrics
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"project_id":        m.ProjectId,
		"manager_id":        m.ManagerId,
		"total_tasks":       m.TotalTasks,
		"completed_tasks":   m.CompletedTasks,
		"completed_on_time": m.CompletedOnTime,
		"completed_late":    m.CompletedLate,
		"in_progress_tasks": m.InProgressTasks,
		"overdue_tasks":     m.OverdueTasks,
		"team_size":         m.TeamSize,
		"progress_percent":  m.ProgressPercent,
		"on_time_rate":      m.OnTimeRate,
		"health_status":     m.HealthStatus.String(),
	})
}

func (h *AnalyticsHTTPHandler) GetEmployeeAnalytics(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := parseID(idStr)
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}

	resp, err := h.client.GetEmployeeMetrics(r.Context(), &pb.GetEmployeeMetricsRequest{EmployeeId: id})
	if err != nil {
		log.Println("employee analytics error:", err)
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	m := resp.Metrics
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"employee_id":       m.EmployeeId,
		"assigned_tasks":    m.AssignedTasks,
		"completed_tasks":   m.CompletedTasks,
		"completed_on_time": m.CompletedOnTime,
		"completed_late":    m.CompletedLate,
		"in_progress_tasks": m.InProgressTasks,
		"overdue_tasks":     m.OverdueTasks,
		"completion_rate":   m.CompletionRate,
		"on_time_rate":      m.OnTimeRate,
	})
}

func (h *AnalyticsHTTPHandler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	resp, err := h.client.GetDashboardStats(r.Context(), &pb.GetDashboardStatsRequest{})
	if err != nil {

		if r.Context().Err() != context.Canceled && status.Code(err) != codes.Canceled {
			log.Println("dashboard error:", err)
		}
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	topEmployees := make([]map[string]interface{}, len(resp.TopEmployees))
	for i, e := range resp.TopEmployees {
		topEmployees[i] = map[string]interface{}{
			"employee_id":     e.EmployeeId,
			"completion_rate": e.CompletionRate,
			"tasks_completed": e.TasksCompleted,
		}
	}

	problematicProjects := make([]map[string]interface{}, len(resp.ProblematicProjects))
	for i, p := range resp.ProblematicProjects {
		problematicProjects[i] = map[string]interface{}{
			"project_id":    p.ProjectId,
			"on_time_rate":  p.OnTimeRate,
			"health_status": p.HealthStatus.String(),
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"total_employees":      resp.TotalEmployees,
		"active_employees":     resp.ActiveEmployees,
		"total_projects":       resp.TotalProjects,
		"active_projects":      resp.ActiveProjects,
		"total_tasks":          resp.TotalTasks,
		"completed_tasks":      resp.CompletedTasks,
		"overdue_tasks":        resp.OverdueTasks,
		"completed_on_time":    resp.CompletedOnTime,
		"completed_late":       resp.CompletedLate,
		"avg_completion_rate":  resp.AvgCompletionRate,
		"avg_on_time_rate":     resp.AvgOnTimeRate,
		"top_employees":        topEmployees,
		"problematic_projects": problematicProjects,
	})
}


