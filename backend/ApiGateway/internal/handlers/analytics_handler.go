package handlers

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/foksdanilka34-maker/F5ProjectUsersControl/ApiGateway/internal/models"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/ApiGateway/internal/service"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/ApiGateway/pkg/response"
	analyticsv1 "github.com/foksdanilka34-maker/F5ProjectUsersControl/gen/go/analytics_service"
	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type AnalyticsHandler struct {
	analyticsService *service.AnalyticsServiceClient
	projectService   *service.ProjectServiceClient
}

func NewAnalyticsHandler(analyticsService *service.AnalyticsServiceClient, projectService *service.ProjectServiceClient) *AnalyticsHandler {
	return &AnalyticsHandler{
		analyticsService: analyticsService,
		projectService:   projectService,
	}
}

// GetDashboardStats returns aggregated dashboard metrics for the configured period.
// @Summary      Get dashboard statistics
// @Tags         Analytics
// @Security     ApiKeyAuth
// @Produce      json
// @Param        start_date query string false "Start of date range"
// @Param        end_date   query string false "End of date range"
// @Success      200 {object} response.Response
// @Failure      400 {object} response.Response
// @Failure      500 {object} response.Response
// @Router       /api/v1/analytics/dashboard [get]
func (h *AnalyticsHandler) GetDashboardStats(c *gin.Context) {
	var query models.DashboardStatsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		log.Printf("GetDashboardStats bind error: %v", err)
		response.BadRequest(c, "Invalid dashboard stats query: "+err.Error())
		return
	}

	req := &analyticsv1.GetDashboardStatsRequest{}
	start, end, err := parseDateRange(query.StartDate, query.EndDate)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	req.StartDate = start
	req.EndDate = end

	stats, err := h.analyticsService.GetDashboardStats(c.Request.Context(), req)
	if err != nil {
		log.Printf("GetDashboardStats service error: %v", err)
		response.InternalServerError(c, "Failed to fetch dashboard stats")
		return
	}

	response.Success(c, http.StatusOK, stats, "Dashboard stats retrieved successfully")
}

// StreamDashboardStats streams dashboard stats via Server-Sent Events for the configured period.
// @Summary      Stream dashboard statistics
// @Tags         Analytics
// @Security     ApiKeyAuth
// @Produce      text/event-stream
// @Param        start_date       query string false "Start of date range"
// @Param        end_date         query string false "End of date range"
// @Param        interval_seconds query int    false "Refresh interval seconds (default 5)"
// @Success      200 {string} string "SSE stream"
// @Failure      400 {object} response.Response
// @Router       /api/v1/analytics/dashboard/stream [get]
func (h *AnalyticsHandler) StreamDashboardStats(c *gin.Context) {
	var query models.DashboardStatsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		log.Printf("StreamDashboardStats bind error: %v", err)
		response.BadRequest(c, "Invalid dashboard stats query: "+err.Error())
		return
	}

	req := &analyticsv1.GetDashboardStatsRequest{}
	start, end, err := parseDateRange(query.StartDate, query.EndDate)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	req.StartDate = start
	req.EndDate = end

	interval := 5 * time.Second
	if raw := strings.TrimSpace(c.Query("interval_seconds")); raw != "" {
		if seconds, err := strconv.Atoi(raw); err == nil && seconds > 0 {
			interval = time.Duration(seconds) * time.Second
		}
	}

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")

	sendSnapshot := func(ctx context.Context) {
		stats, err := h.analyticsService.GetDashboardStats(ctx, req)
		if err != nil {
			log.Printf("StreamDashboardStats service error: %v", err)
			c.SSEvent("dashboard-stats", response.Response{
				Success: false,
				Error:   "Failed to fetch dashboard stats",
			})
			return
		}

		c.SSEvent("dashboard-stats", response.Response{
			Success: true,
			Message: "Dashboard stats update",
			Data:    stats,
		})
	}

	ctx := c.Request.Context()
	sendSnapshot(ctx)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	c.Stream(func(w io.Writer) bool {
		select {
		case <-ctx.Done():
			return false
		case <-ticker.C:
			sendSnapshot(ctx)
			return true
		}
	})
}

// GetEmployeeMetrics fetches metrics for a single employee.
// @Summary      Get employee metrics
// @Tags         Analytics
// @Security     ApiKeyAuth
// @Produce      json
// @Param        id path string true "Employee ID"
// @Success      200 {object} response.Response
// @Failure      400 {object} response.Response
// @Failure      500 {object} response.Response
// @Router       /api/v1/analytics/employees/{id}/metrics [get]
func (h *AnalyticsHandler) GetEmployeeMetrics(c *gin.Context) {
	employeeID := strings.TrimSpace(c.Param("id"))
	if employeeID == "" {
		response.BadRequest(c, "Employee ID is required")
		return
	}

	req := &analyticsv1.GetEmployeeMetricsRequest{EmployeeId: employeeID}
	metrics, err := h.analyticsService.GetEmployeeMetrics(c.Request.Context(), req)
	if err != nil {
		log.Printf("GetEmployeeMetrics service error: %v", err)
		response.InternalServerError(c, "Failed to fetch employee metrics")
		return
	}

	response.Success(c, http.StatusOK, metrics, "Employee metrics retrieved successfully")
}

// ListEmployeeMetrics lists employee metrics with pagination and optional filters.
// @Summary      List employee metrics
// @Tags         Analytics
// @Security     ApiKeyAuth
// @Produce      json
// @Param        page_size     query int false "Page size"
// @Param        page_number   query int false "Page number"
// @Param        start_date    query string false "Start date"
// @Param        end_date      query string false "End date"
// @Param        department_id query string false "Department filter"
// @Success      200 {object} response.Response
// @Failure      400 {object} response.Response
// @Failure      500 {object} response.Response
// @Router       /api/v1/analytics/employees/metrics [get]
func (h *AnalyticsHandler) ListEmployeeMetrics(c *gin.Context) {
	var query models.ListEmployeeMetricsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		log.Printf("ListEmployeeMetrics bind error: %v", err)
		response.BadRequest(c, "Invalid metrics query: "+err.Error())
		return
	}

	normalizePagination(&query.PaginationQuery)
	start, end, err := parseDateRange(query.StartDate, query.EndDate)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	req := &analyticsv1.ListEmployeeMetricsRequest{
		PageSize:   proto.Int32(query.PageSize),
		PageNumber: proto.Int32(query.PageNumber),
		StartDate:  start,
		EndDate:    end,
	}
	if query.DepartmentID != "" {
		req.DepartmentId = proto.String(query.DepartmentID)
	}

	resp, err := h.analyticsService.ListEmployeeMetrics(c.Request.Context(), req)
	if err != nil {
		log.Printf("ListEmployeeMetrics service error: %v", err)
		response.InternalServerError(c, "Failed to list employee metrics")
		return
	}

	meta := response.PaginationMeta{
		PageSize:   query.PageSize,
		PageNumber: query.PageNumber,
		TotalCount: int64(resp.GetTotalCount()),
	}
	response.Paginated(c, http.StatusOK, resp, meta, "Employee metrics retrieved successfully")
}

// GetTopPerformers returns best-performing employees based on productivity metrics.
// @Summary      Get top performers
// @Tags         Analytics
// @Security     ApiKeyAuth
// @Produce      json
// @Param        start_date    query string false "Start date"
// @Param        end_date      query string false "End date"
// @Param        limit         query int false "Max number of performers"
// @Param        department_id query string false "Department filter"
// @Success      200 {object} response.Response
// @Failure      400 {object} response.Response
// @Failure      500 {object} response.Response
// @Router       /api/v1/analytics/employees/top-performers [get]
func (h *AnalyticsHandler) GetTopPerformers(c *gin.Context) {
	var query models.TopPerformersQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		log.Printf("GetTopPerformers bind error: %v", err)
		response.BadRequest(c, "Invalid top performers query: "+err.Error())
		return
	}

	limit := query.Limit
	if limit == 0 {
		limit = 10
	}

	start, end, err := parseDateRange(query.StartDate, query.EndDate)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	req := &analyticsv1.GetTopPerformersRequest{
		Limit:     limit,
		StartDate: start,
		EndDate:   end,
	}
	if query.DepartmentID != "" {
		req.DepartmentId = proto.String(query.DepartmentID)
	}

	resp, err := h.analyticsService.GetTopPerformers(c.Request.Context(), req)
	if err != nil {
		log.Printf("GetTopPerformers service error: %v", err)
		response.InternalServerError(c, "Failed to fetch top performers")
		return
	}

	meta := response.PaginationMeta{PageSize: limit, PageNumber: 1, TotalCount: int64(resp.GetTotalCount())}
	response.Paginated(c, http.StatusOK, resp, meta, "Top performers retrieved successfully")
}

// GetProjectMetrics fetches aggregated metrics for the given project.
// @Summary      Get project metrics
// @Tags         Analytics
// @Security     ApiKeyAuth
// @Produce      json
// @Param        id         path string true "Project ID"
// @Param        start_date query string false "Start date"
// @Param        end_date   query string false "End date"
// @Success      200 {object} response.Response
// @Failure      400 {object} response.Response
// @Failure      500 {object} response.Response
// @Router       /api/v1/analytics/projects/{id}/metrics [get]
func (h *AnalyticsHandler) GetProjectMetrics(c *gin.Context) {
	projectID := strings.TrimSpace(c.Param("id"))
	if projectID == "" {
		response.BadRequest(c, "Project ID is required")
		return
	}

	var query models.ProjectMetricsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		log.Printf("GetProjectMetrics bind error: %v", err)
		response.BadRequest(c, "Invalid project metrics query: "+err.Error())
		return
	}

	start, end, err := parseDateRange(query.StartDate, query.EndDate)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	req := &analyticsv1.GetProjectMetricsRequest{
		ProjectId: projectID,
		StartDate: start,
		EndDate:   end,
	}

	metrics, err := h.analyticsService.GetProjectMetrics(c.Request.Context(), req)
	if err != nil {
		log.Printf("GetProjectMetrics service error: %v", err)
		response.InternalServerError(c, "Failed to fetch project metrics")
		return
	}

	response.Success(c, http.StatusOK, metrics, "Project metrics retrieved successfully")
}

// ListProjectMetrics lists project metrics with pagination and optional manager filter.
// @Summary      List project metrics
// @Tags         Analytics
// @Security     ApiKeyAuth
// @Produce      json
// @Param        page_size   query int false "Page size"
// @Param        page_number query int false "Page number"
// @Param        start_date  query string false "Start date"
// @Param        end_date    query string false "End date"
// @Param        manager_id  query string false "Manager filter"
// @Success      200 {object} response.Response
// @Failure      400 {object} response.Response
// @Failure      500 {object} response.Response
// @Router       /api/v1/analytics/projects/metrics [get]
func (h *AnalyticsHandler) ListProjectMetrics(c *gin.Context) {
	var query models.ListProjectMetricsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		log.Printf("ListProjectMetrics bind error: %v", err)
		response.BadRequest(c, "Invalid project metrics query: "+err.Error())
		return
	}

	normalizePagination(&query.PaginationQuery)
	start, end, err := parseDateRange(query.StartDate, query.EndDate)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	req := &analyticsv1.ListProjectMetricsRequest{
		PageSize:   proto.Int32(query.PageSize),
		PageNumber: proto.Int32(query.PageNumber),
		StartDate:  start,
		EndDate:    end,
	}
	if query.ManagerID != "" {
		req.ManagerId = proto.String(query.ManagerID)
	}

	resp, err := h.analyticsService.ListProjectMetrics(c.Request.Context(), req)
	if err != nil {
		log.Printf("ListProjectMetrics service error: %v", err)
		response.InternalServerError(c, "Failed to list project metrics")
		return
	}

	meta := response.PaginationMeta{
		PageSize:   query.PageSize,
		PageNumber: query.PageNumber,
		TotalCount: int64(resp.GetTotalCount()),
	}
	response.Paginated(c, http.StatusOK, resp, meta, "Project metrics retrieved successfully")
}

// GetProductivityTrends returns productivity trends for the given teams.
// @Summary      Get productivity trends
// @Tags         Analytics
// @Security     ApiKeyAuth
// @Produce      json
// @Param        period        query string false "Period (DAILY/WEEKLY/MONTHLY)"
// @Param        limit         query int false "Number of data points"
// @Param        department_id query string false "Department filter"
// @Param        employee_id   query string false "Employee filter"
// @Success      200 {object} response.Response
// @Failure      400 {object} response.Response
// @Failure      500 {object} response.Response
// @Router       /api/v1/analytics/trends/productivity [get]
func (h *AnalyticsHandler) GetProductivityTrends(c *gin.Context) {
	var query models.ProductivityTrendsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		log.Printf("GetProductivityTrends bind error: %v", err)
		response.BadRequest(c, "Invalid productivity trends query: "+err.Error())
		return
	}

	period, err := parsePeriod(query.Period)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	limit := query.Limit
	if limit == 0 {
		limit = 30
	}

	req := &analyticsv1.GetProductivityTrendsRequest{
		Period: period,
		Limit:  limit,
	}
	if query.DepartmentID != "" {
		req.DepartmentId = proto.String(query.DepartmentID)
	}
	if query.EmployeeID != "" {
		req.EmployeeId = proto.String(query.EmployeeID)
	}

	trends, err := h.analyticsService.GetProductivityTrends(c.Request.Context(), req)
	if err != nil {
		log.Printf("GetProductivityTrends service error: %v", err)
		response.InternalServerError(c, "Failed to fetch productivity trends")
		return
	}

	response.Success(c, http.StatusOK, trends, "Productivity trends retrieved successfully")
}

// GetCompletionRateTrends returns completion rate trends for projects or departments.
// @Summary      Get completion rate trends
// @Tags         Analytics
// @Security     ApiKeyAuth
// @Produce      json
// @Param        period        query string false "Period (DAILY/WEEKLY/MONTHLY)"
// @Param        limit         query int false "Number of data points"
// @Param        project_id    query string false "Project filter"
// @Param        department_id query string false "Department filter"
// @Success      200 {object} response.Response
// @Failure      400 {object} response.Response
// @Failure      500 {object} response.Response
// @Router       /api/v1/analytics/trends/completion-rate [get]
func (h *AnalyticsHandler) GetCompletionRateTrends(c *gin.Context) {
	var query models.CompletionRateTrendsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		log.Printf("GetCompletionRateTrends bind error: %v", err)
		response.BadRequest(c, "Invalid completion trends query: "+err.Error())
		return
	}

	period, err := parsePeriod(query.Period)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	limit := query.Limit
	if limit == 0 {
		limit = 30
	}

	req := &analyticsv1.GetCompletionRateTrendsRequest{
		Period: period,
		Limit:  limit,
	}
	if query.ProjectID != "" {
		req.ProjectId = proto.String(query.ProjectID)
	}
	if query.DepartmentID != "" {
		req.DepartmentId = proto.String(query.DepartmentID)
	}

	trends, err := h.analyticsService.GetCompletionRateTrends(c.Request.Context(), req)
	if err != nil {
		log.Printf("GetCompletionRateTrends service error: %v", err)
		response.InternalServerError(c, "Failed to fetch completion rate trends")
		return
	}

	response.Success(c, http.StatusOK, trends, "Completion rate trends retrieved successfully")
}

func normalizePagination(p *models.PaginationQuery) {
	if p == nil {
		return
	}
	if p.PageSize <= 0 {
		p.PageSize = 20
	} else if p.PageSize > 100 {
		p.PageSize = 100
	}
	if p.PageNumber <= 0 {
		p.PageNumber = 1
	}
}

func parseDateRange(startRaw, endRaw string) (*timestamppb.Timestamp, *timestamppb.Timestamp, error) {
	start, err := toTimestamp(startRaw)
	if err != nil {
		return nil, nil, err
	}
	end, err := toTimestamp(endRaw)
	if err != nil {
		return nil, nil, err
	}

	if start != nil && end != nil && start.AsTime().After(end.AsTime()) {
		return nil, nil, fmt.Errorf("start_date must be before end_date")
	}

	return start, end, nil
}

func toTimestamp(value string) (*timestamppb.Timestamp, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}

	layouts := []string{time.RFC3339, "2006-01-02"}
	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, value); err == nil {
			return timestamppb.New(parsed), nil
		}
	}

	return nil, fmt.Errorf("invalid date format: %s", value)
}

func parsePeriod(value string) (analyticsv1.Period, error) {
	if value == "" {
		return analyticsv1.Period_PERIOD_WEEKLY, nil
	}

	switch strings.ToUpper(value) {
	case "DAILY":
		return analyticsv1.Period_PERIOD_DAILY, nil
	case "WEEKLY":
		return analyticsv1.Period_PERIOD_WEEKLY, nil
	case "MONTHLY":
		return analyticsv1.Period_PERIOD_MONTHLY, nil
	default:
		return analyticsv1.Period_PERIOD_UNSPECIFIED, fmt.Errorf("invalid period value: %s", value)
	}
}
