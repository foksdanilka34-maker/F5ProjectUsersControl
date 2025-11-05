package handlers

import (
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/foksdanilka34-maker/F5ProjectUsersControl/ApiGateway/internal/models"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/ApiGateway/internal/service"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/ApiGateway/pkg/response"
	analyticsv1 "github.com/foksdanilka34-maker/F5ProjectUsersControl/gen/go/analytics_service"
	projectpb "github.com/foksdanilka34-maker/F5ProjectUsersControl/gen/go/project_service"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var supportedDateLayouts = []string{
	time.RFC3339,
	"2006-01-02",
}

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

func (h *AnalyticsHandler) GetEmployeeMetrics(c *gin.Context) {
	employeeID := c.Param("id")
	if employeeID == "" {
		response.BadRequest(c, "Employee ID is required")
		return
	}

	var query models.EmployeeMetricsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c, "Invalid query parameters: "+err.Error())
		return
	}

	startTS, endTS, err := parseDateRange(query.StartDate, query.EndDate)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	req := &analyticsv1.GetEmployeeMetricsRequest{
		EmployeeId: employeeID,
	}
	if startTS != nil {
		req.StartDate = startTS
	}
	if endTS != nil {
		req.EndDate = endTS
	}

	metrics, err := h.analyticsService.GetEmployeeMetrics(c.Request.Context(), req)
	if err != nil {
		log.Printf("GetEmployeeMetrics service error: %v", err)
		response.InternalServerError(c, "Failed to retrieve employee metrics")
		return
	}

	response.Success(c, http.StatusOK, metrics, "Employee metrics retrieved successfully")
}

func (h *AnalyticsHandler) ListEmployeeMetrics(c *gin.Context) {
	var query models.ListEmployeeMetricsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c, "Invalid query parameters: "+err.Error())
		return
	}

	if query.PageSize == 0 {
		query.PageSize = 10
	}
	if query.PageNumber == 0 {
		query.PageNumber = 1
	}

	startTS, endTS, err := parseDateRange(query.StartDate, query.EndDate)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	req := &analyticsv1.ListEmployeeMetricsRequest{
		PageSize:   query.PageSize,
		PageNumber: query.PageNumber,
	}
	if query.DepartmentID != "" {
		req.DepartmentId = &query.DepartmentID
	}
	if startTS != nil {
		req.StartDate = startTS
	}
	if endTS != nil {
		req.EndDate = endTS
	}

	metrics, err := h.analyticsService.ListEmployeeMetrics(c.Request.Context(), req)
	if err != nil {
		log.Printf("ListEmployeeMetrics service error: %v", err)
		response.InternalServerError(c, "Failed to list employee metrics")
		return
	}

	response.Success(c, http.StatusOK, metrics, "Employee metrics retrieved successfully")
}

func (h *AnalyticsHandler) GetTopPerformers(c *gin.Context) {
	var query models.TopPerformersQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c, "Invalid query parameters: "+err.Error())
		return
	}

	if query.Limit == 0 {
		query.Limit = 5
	}

	startTS, endTS, err := parseDateRange(query.StartDate, query.EndDate)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	req := &analyticsv1.GetTopPerformersRequest{
		Limit: query.Limit,
	}
	if query.DepartmentID != "" {
		req.DepartmentId = &query.DepartmentID
	}
	if startTS != nil {
		req.StartDate = startTS
	}
	if endTS != nil {
		req.EndDate = endTS
	}

	metrics, err := h.analyticsService.GetTopPerformers(c.Request.Context(), req)
	if err != nil {
		log.Printf("GetTopPerformers service error: %v", err)
		response.InternalServerError(c, "Failed to retrieve top performers")
		return
	}

	response.Success(c, http.StatusOK, metrics, "Top performers retrieved successfully")
}

func (h *AnalyticsHandler) GetProjectMetrics(c *gin.Context) {
	projectID := c.Param("id")
	if projectID == "" {
		response.BadRequest(c, "Project ID is required")
		return
	}

	project, err := h.ensureProjectExists(c, projectID)
	if err != nil {
		return
	}

	var query models.ProjectMetricsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c, "Invalid query parameters: "+err.Error())
		return
	}

	startTS, endTS, err := parseDateRange(query.StartDate, query.EndDate)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	req := &analyticsv1.GetProjectMetricsRequest{
		ProjectId: projectID,
	}
	if startTS != nil {
		req.StartDate = startTS
	}
	if endTS != nil {
		req.EndDate = endTS
	}

	metrics, err := h.analyticsService.GetProjectMetrics(c.Request.Context(), req)
	if err != nil {
		log.Printf("GetProjectMetrics service error: %v", err)
		response.InternalServerError(c, "Failed to retrieve project metrics")
		return
	}

	payload := gin.H{
		"metrics":       metrics.Metrics,
		"calculated_at": metrics.CalculatedAt,
	}
	if project != nil {
		payload["project"] = project
	}

	response.Success(c, http.StatusOK, payload, "Project metrics retrieved successfully")
}

func (h *AnalyticsHandler) ListProjectMetrics(c *gin.Context) {
	var query models.ListProjectMetricsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c, "Invalid query parameters: "+err.Error())
		return
	}

	if query.PageSize == 0 {
		query.PageSize = 10
	}
	if query.PageNumber == 0 {
		query.PageNumber = 1
	}

	startTS, endTS, err := parseDateRange(query.StartDate, query.EndDate)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	req := &analyticsv1.ListProjectMetricsRequest{
		PageSize:   query.PageSize,
		PageNumber: query.PageNumber,
	}
	if query.ManagerID != "" {
		req.ManagerId = &query.ManagerID
	}
	if startTS != nil {
		req.StartDate = startTS
	}
	if endTS != nil {
		req.EndDate = endTS
	}

	metrics, err := h.analyticsService.ListProjectMetrics(c.Request.Context(), req)
	if err != nil {
		log.Printf("ListProjectMetrics service error: %v", err)
		response.InternalServerError(c, "Failed to list project metrics")
		return
	}

	projects := h.collectProjects(c, metrics.Metrics)

	payload := gin.H{
		"metrics":     metrics.Metrics,
		"total_count": metrics.TotalCount,
	}
	if len(projects) > 0 {
		payload["projects"] = projects
	}

	response.Success(c, http.StatusOK, payload, "Project metrics retrieved successfully")
}

func (h *AnalyticsHandler) GetDepartmentMetrics(c *gin.Context) {
	departmentID := c.Param("id")
	if departmentID == "" {
		response.BadRequest(c, "Department ID is required")
		return
	}

	var query models.DepartmentMetricsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c, "Invalid query parameters: "+err.Error())
		return
	}

	startTS, endTS, err := parseDateRange(query.StartDate, query.EndDate)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	req := &analyticsv1.GetDepartmentMetricsRequest{
		DepartmentId: departmentID,
	}
	if startTS != nil {
		req.StartDate = startTS
	}
	if endTS != nil {
		req.EndDate = endTS
	}

	metrics, err := h.analyticsService.GetDepartmentMetrics(c.Request.Context(), req)
	if err != nil {
		log.Printf("GetDepartmentMetrics service error: %v", err)
		response.InternalServerError(c, "Failed to retrieve department metrics")
		return
	}

	response.Success(c, http.StatusOK, metrics, "Department metrics retrieved successfully")
}

func (h *AnalyticsHandler) ListDepartmentMetrics(c *gin.Context) {
	var query models.ListDepartmentMetricsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c, "Invalid query parameters: "+err.Error())
		return
	}

	if query.PageSize == 0 {
		query.PageSize = 10
	}
	if query.PageNumber == 0 {
		query.PageNumber = 1
	}

	startTS, endTS, err := parseDateRange(query.StartDate, query.EndDate)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	req := &analyticsv1.ListDepartmentMetricsRequest{
		PageSize:   query.PageSize,
		PageNumber: query.PageNumber,
	}
	if startTS != nil {
		req.StartDate = startTS
	}
	if endTS != nil {
		req.EndDate = endTS
	}

	metrics, err := h.analyticsService.ListDepartmentMetrics(c.Request.Context(), req)
	if err != nil {
		log.Printf("ListDepartmentMetrics service error: %v", err)
		response.InternalServerError(c, "Failed to list department metrics")
		return
	}

	response.Success(c, http.StatusOK, metrics, "Department metrics retrieved successfully")
}

func (h *AnalyticsHandler) GetProductivityTrends(c *gin.Context) {
	var query models.ProductivityTrendsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c, "Invalid query parameters: "+err.Error())
		return
	}

	if query.Limit == 0 {
		query.Limit = 30
	}
	if query.Period == "" {
		query.Period = "DAILY"
	}

	req := &analyticsv1.GetProductivityTrendsRequest{
		Period: query.Period,
		Limit:  query.Limit,
	}
	if query.DepartmentID != "" {
		req.DepartmentId = &query.DepartmentID
	}
	if query.EmployeeID != "" {
		req.EmployeeId = &query.EmployeeID
	}

	trends, err := h.analyticsService.GetProductivityTrends(c.Request.Context(), req)
	if err != nil {
		log.Printf("GetProductivityTrends service error: %v", err)
		response.InternalServerError(c, "Failed to retrieve productivity trends")
		return
	}

	response.Success(c, http.StatusOK, trends, "Productivity trends retrieved successfully")
}

func (h *AnalyticsHandler) GetCompletionRateTrends(c *gin.Context) {
	var query models.CompletionRateTrendsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c, "Invalid query parameters: "+err.Error())
		return
	}

	if query.Limit == 0 {
		query.Limit = 30
	}
	if query.Period == "" {
		query.Period = "DAILY"
	}

	if query.ProjectID != "" {
		if _, err := h.ensureProjectExists(c, query.ProjectID); err != nil {
			return
		}
	}

	req := &analyticsv1.GetCompletionRateTrendsRequest{
		Period: query.Period,
		Limit:  query.Limit,
	}
	if query.ProjectID != "" {
		req.ProjectId = &query.ProjectID
	}
	if query.DepartmentID != "" {
		req.DepartmentId = &query.DepartmentID
	}

	trends, err := h.analyticsService.GetCompletionRateTrends(c.Request.Context(), req)
	if err != nil {
		log.Printf("GetCompletionRateTrends service error: %v", err)
		response.InternalServerError(c, "Failed to retrieve completion rate trends")
		return
	}

	response.Success(c, http.StatusOK, trends, "Completion rate trends retrieved successfully")
}

func (h *AnalyticsHandler) GetDashboardStats(c *gin.Context) {
	var query models.DashboardStatsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c, "Invalid query parameters: "+err.Error())
		return
	}

	startTS, endTS, err := parseDateRange(query.StartDate, query.EndDate)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	req := &analyticsv1.GetDashboardStatsRequest{}
	if startTS != nil {
		req.StartDate = startTS
	}
	if endTS != nil {
		req.EndDate = endTS
	}

	stats, err := h.analyticsService.GetDashboardStats(c.Request.Context(), req)
	if err != nil {
		log.Printf("GetDashboardStats service error: %v", err)
		response.InternalServerError(c, "Failed to retrieve dashboard stats")
		return
	}

	response.Success(c, http.StatusOK, stats, "Dashboard stats retrieved successfully")
}

func (h *AnalyticsHandler) ensureProjectExists(c *gin.Context, projectID string) (*projectpb.Project, error) {
	project, err := h.projectService.GetProject(c.Request.Context(), projectID)
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			response.NotFound(c, "Project not found")
			return nil, err
		}
		log.Printf("ensureProjectExists error: %v", err)
		response.InternalServerError(c, "Failed to verify project")
		return nil, err
	}
	return project, nil
}

func (h *AnalyticsHandler) collectProjects(c *gin.Context, metrics []*analyticsv1.ProjectMetrics) map[string]*projectpb.Project {
	projects := make(map[string]*projectpb.Project)
	for _, metric := range metrics {
		if metric == nil || metric.ProjectId == "" {
			continue
		}
		if _, exists := projects[metric.ProjectId]; exists {
			continue
		}

		project, err := h.projectService.GetProject(c.Request.Context(), metric.ProjectId)
		if err != nil {
			st, ok := status.FromError(err)
			if ok && st.Code() == codes.NotFound {
				log.Printf("collectProjects warning: project %s not found", metric.ProjectId)
			} else {
				log.Printf("collectProjects error: %v", err)
			}
			continue
		}
		projects[metric.ProjectId] = project
	}
	return projects
}

func parseDateRange(start, end string) (*timestamppb.Timestamp, *timestamppb.Timestamp, error) {
	startTime, err := parseDate(start)
	if err != nil {
		return nil, nil, err
	}
	endTime, err := parseDate(end)
	if err != nil {
		return nil, nil, err
	}
	if startTime != nil && endTime != nil && startTime.After(*endTime) {
		return nil, nil, errors.New("start_date must be before end_date")
	}

	var startTS, endTS *timestamppb.Timestamp
	if startTime != nil {
		startTS = timestamppb.New(*startTime)
	}
	if endTime != nil {
		endTS = timestamppb.New(*endTime)
	}
	return startTS, endTS, nil
}

func parseDate(value string) (*time.Time, error) {
	if value == "" {
		return nil, nil
	}
	for _, layout := range supportedDateLayouts {
		if t, err := time.Parse(layout, value); err == nil {
			parsed := t
			return &parsed, nil
		}
	}
	return nil, errors.New("invalid date format, use RFC3339 or YYYY-MM-DD")
}
