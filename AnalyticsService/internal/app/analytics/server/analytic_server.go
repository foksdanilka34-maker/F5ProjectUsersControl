package server

import (
	"context"
	"log"
	"time"

	"github.com/foksdanilka34-maker/F5ProjectUsersControl/AnalyticsService/internal/app/analytics"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/AnalyticsService/internal/app/analytics/core"
	analyticsv1 "github.com/foksdanilka34-maker/F5ProjectUsersControl/gen/go/analytics_service"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type AnalyticsServer struct {
	analyticsv1.UnimplementedAnalyticsServiceServer
	core core.CoreLogic
}

func NewAnalyticsServer(core core.CoreLogic) *AnalyticsServer {
	return &AnalyticsServer{
		core: core,
	}
}

func (s *AnalyticsServer) Register(grpcServer *grpc.Server) {
	analyticsv1.RegisterAnalyticsServiceServer(grpcServer, s)
}

func (s *AnalyticsServer) GetEmployeeMetrics(ctx context.Context, req *analyticsv1.GetEmployeeMetricsRequest) (*analyticsv1.EmployeeMetricsResponse, error) {
	metrics, err := s.core.GetEmployeeMetrics(ctx, req.EmployeeId)
	if err != nil {
		log.Printf("failed to get employee metrics: %v", err)
		return nil, err
	}

	metricsResponse := s.employeeMetricsToProto(metrics)

	return &analyticsv1.EmployeeMetricsResponse{
		Metrics: metricsResponse,
	}, nil
}

func (s *AnalyticsServer) employeeMetricsToProto(m *analytics.EmployeeMetrics) *analyticsv1.EmployeeMetrics {
	return &analyticsv1.EmployeeMetrics{
		EmployeeId:               m.EmployeeID,
		MetricDate:               timestamppb.New(m.MetricDate),
		AssignedTasks:            m.AssignedTasks,
		CompletedTasks:           m.CompletedTasks,
		InProgressTasks:          m.InProgressTasks,
		OverdueTasks:             m.OverdueTasks,
		OnTimeCompletedTasks:     m.OnTimeCompletionTask,
		TotalTaskDurationSeconds: m.TotalTaskDurationSeconds,

		TaskCompletionRate:   &m.TaskCompletionRate,
		OnTimeCompletionRate: &m.OnTimeCompletionRate,
		EfficiencyScore:      &m.EfficiencyScore,
	}
}

func (s *AnalyticsServer) GetProjectMetrics(ctx context.Context, req *analyticsv1.GetProjectMetricsRequest) (*analyticsv1.ProjectMetricsResponse, error) {
	metrics, err := s.core.GetProjectMetrics(ctx, req.ProjectId)
	if err != nil {
		log.Printf("failed to get project metrics: %v", err)
		return nil, err
	}

	metricsResponse := s.projectMetricsToProto(metrics)

	return &analyticsv1.ProjectMetricsResponse{
		Metrics:      metricsResponse,
		CalculatedAt: timestamppb.Now(),
	}, nil
}

func (s *AnalyticsServer) projectMetricsToProto(m *analytics.ProjectMetrics) *analyticsv1.ProjectMetrics {
	var projectedEndDate *timestamppb.Timestamp
	if m.ProjectedEndDate != nil {
		projectedEndDate = timestamppb.New(*m.ProjectedEndDate)
	}

	var daysUntilDue *int32
	if m.DaysUntilDue != nil {
		daysUntilDue = m.DaysUntilDue
	}

	return &analyticsv1.ProjectMetrics{
		ProjectId:                         m.ProjectID,
		ManagerId:                         m.ManagerID,
		MetricDate:                        timestamppb.New(m.MetricDate),
		TotalTasks:                        m.TotalTasks,
		CompletedTasks:                    m.CompletedTasks,
		InProgressTasks:                   m.InProgressTasks,
		OverdueTasks:                      m.OverdueTasks,
		OnTimeCompletedTasks:              m.OnTimeCompletedTasks,
		TeamSize:                          m.TeamSize,
		TotalTaskDurationSecondsCompleted: m.TotalTaskDurationSecondsCompleted,
		TotalPriorityWeightCompleted:      m.TotalPriorityWeightCompleted,
		DeliveryPerformance:               m.DeliveryPerformance,
		SchedulePerformance:               m.SchedulePerformance,
		QualityPerformance:                m.QualityPerformance,
		TeamPerformance:                   m.TeamPerformance,
		HealthIndex:                       m.HealthIndex,
		RiskScore:                         m.RiskScore,
		HealthStatus:                      analyticsv1.HealthStatus(m.HealthStatus),
		Velocity:                          m.Velocity,
		ProjectedEndDate:                  projectedEndDate,
		TeamCapacityUtilization:           m.TeamCapacityUtilization,
		AvgTeamEfficiency:                 m.AvgTeamEfficiency,
		IsAtRisk:                          m.IsAtRisk,
		DaysUntilDue:                      daysUntilDue,
	}
}

func (s *AnalyticsServer) ListEmployeeMetrics(ctx context.Context, req *analyticsv1.ListEmployeeMetricsRequest) (*analyticsv1.ListEmployeeMetricsResponse, error) {
	var pageSize int32
	if req.PageSize != nil {
		pageSize = *req.PageSize
	}

	var pageNumber int32
	if req.PageNumber != nil {
		pageNumber = *req.PageNumber
	}

	var departmentID string
	if req.DepartmentId != nil {
		departmentID = *req.DepartmentId
	}

	var startDate, endDate *time.Time
	if req.StartDate != nil {
		t := req.StartDate.AsTime()
		startDate = &t
	}
	if req.EndDate != nil {
		t := req.EndDate.AsTime()
		endDate = &t
	}

	listReq := &analytics.ListEmployeeMetrics{
		PageSize:     pageSize,
		PageNumber:   pageNumber,
		DepartmentID: departmentID,
		StartDate:    startDate,
		EndDate:      endDate,
	}

	metrics, totalCount, err := s.core.ListEmployeeMetrics(ctx, listReq)
	if err != nil {
		log.Printf("failed to list employee metrics: %v", err)
		return nil, err
	}

	protoMetrics := make([]*analyticsv1.EmployeeMetrics, len(metrics))
	for i, m := range metrics {
		protoMetrics[i] = s.employeeMetricsToProto(m)
	}

	return &analyticsv1.ListEmployeeMetricsResponse{
		Metrics:    protoMetrics,
		TotalCount: totalCount,
	}, nil
}

func (s *AnalyticsServer) GetTopPerformers(ctx context.Context, req *analyticsv1.GetTopPerformersRequest) (*analyticsv1.ListEmployeeMetricsResponse, error) {
	var startDate, endDate *time.Time
	if req.StartDate != nil {
		t := req.StartDate.AsTime()
		startDate = &t
	}
	if req.EndDate != nil {
		t := req.EndDate.AsTime()
		endDate = &t
	}

	metrics, err := s.core.GetTopPerformers(ctx, req.Limit, startDate, endDate)
	if err != nil {
		log.Printf("failed to get top performers: %v", err)
		return nil, err
	}

	protoMetrics := make([]*analyticsv1.EmployeeMetrics, len(metrics))
	for i, m := range metrics {
		protoMetrics[i] = s.employeeMetricsToProto(m)
	}

	return &analyticsv1.ListEmployeeMetricsResponse{
		Metrics:    protoMetrics,
		TotalCount: int32(len(metrics)),
	}, nil
}

func (s *AnalyticsServer) ListProjectMetrics(ctx context.Context, req *analyticsv1.ListProjectMetricsRequest) (*analyticsv1.ListProjectMetricsResponse, error) {
	var pageSize, pageNumber *int32
	if req.PageSize != nil {
		pageSize = req.PageSize
	}
	if req.PageNumber != nil {
		pageNumber = req.PageNumber
	}

	var managerID string
	if req.ManagerId != nil {
		managerID = *req.ManagerId
	}

	var startDate, endDate *time.Time
	if req.StartDate != nil {
		t := req.StartDate.AsTime()
		startDate = &t
	}
	if req.EndDate != nil {
		t := req.EndDate.AsTime()
		endDate = &t
	}

	listReq := &analytics.ListProjectMetrics{
		PageSize:   pageSize,
		PageNumber: pageNumber,
		ManagerID:  managerID,
		StartDate:  startDate,
		EndDate:    endDate,
	}

	metrics, totalCount, err := s.core.ListProjectMetrics(ctx, listReq)
	if err != nil {
		log.Printf("failed to list project metrics: %v", err)
		return nil, err
	}

	protoMetrics := make([]*analyticsv1.ProjectMetrics, len(metrics))
	for i, m := range metrics {
		protoMetrics[i] = s.projectMetricsToProto(m)
	}

	return &analyticsv1.ListProjectMetricsResponse{
		Metrics:    protoMetrics,
		TotalCount: totalCount,
	}, nil
}

func (s *AnalyticsServer) GetProductivityTrends(ctx context.Context, req *analyticsv1.GetProductivityTrendsRequest) (*analyticsv1.ProductivityTrendsResponse, error) {
	var departmentID string
	if req.DepartmentId != nil {
		departmentID = *req.DepartmentId
	}

	var employeeID *string
	if req.EmployeeId != nil {
		employeeID = req.EmployeeId
	}

	trendsReq := &analytics.ProductivityTrends{
		Period:       analytics.Period(req.Period),
		Limit:        req.Limit,
		DepartmentID: departmentID,
		EmployeeID:   employeeID,
	}

	trends, err := s.core.GetProductivityTrends(ctx, trendsReq)
	if err != nil {
		log.Printf("failed to get productivity trends: %v", err)
		return nil, err
	}

	entries := make([]*analyticsv1.ProductivityTrendEntry, len(trends.Prod))
	for i, t := range trends.Prod {
		entries[i] = &analyticsv1.ProductivityTrendEntry{
			Date:                 timestamppb.New(t.Date),
			AvgEfficiency:        float64(t.AvgEfficiency),
			TotalTasksCompleted:  t.TotalTasksCompleted,
			TotalEmployeesActive: t.TotalEmployeesActive,
		}
	}

	return &analyticsv1.ProductivityTrendsResponse{
		Entries: entries,
		Period:  analyticsv1.Period(trends.Period),
	}, nil
}

func (s *AnalyticsServer) GetCompletionRateTrends(ctx context.Context, req *analyticsv1.GetCompletionRateTrendsRequest) (*analyticsv1.CompletionRateTrendsResponse, error) {
	var projectID string
	if req.ProjectId != nil {
		projectID = *req.ProjectId
	}

	trendsReq := &analytics.ComletionRateTrends{
		Period:    analytics.Period(req.Period),
		Limit:     req.Limit,
		ProjectID: projectID,
	}

	trends, err := s.core.GetCompletionRateTrends(ctx, trendsReq)
	if err != nil {
		log.Printf("failed to get completion rate trends: %v", err)
		return nil, err
	}

	entries := make([]*analyticsv1.CompletionRateTrendEntry, len(trends.CompTrend))
	for i, t := range trends.CompTrend {
		entries[i] = &analyticsv1.CompletionRateTrendEntry{
			Date:           timestamppb.New(t.Date),
			OnTimeRate:     float64(t.OnTimeRate),
			OverallRate:    float64(t.OverallRate),
			CompletedCount: t.CompletedCount,
			OverdueCount:   t.OverDueCount,
		}
	}

	return &analyticsv1.CompletionRateTrendsResponse{
		Entries: entries,
		Period:  analyticsv1.Period(trends.Period),
	}, nil
}

func (s *AnalyticsServer) GetDashboardStats(ctx context.Context, req *analyticsv1.GetDashboardStatsRequest) (*analyticsv1.DashboardStatsResponse, error) {
	var startDate, endDate *time.Time
	if req.StartDate != nil {
		t := req.StartDate.AsTime()
		startDate = &t
	}
	if req.EndDate != nil {
		t := req.EndDate.AsTime()
		endDate = &t
	}

	stats, err := s.core.GetDashboardStats(ctx, startDate, endDate)
	if err != nil {
		log.Printf("failed to get dashboard stats: %v", err)
		return nil, err
	}

	topEmployees := make([]*analyticsv1.TopEmployee, len(stats.TopEmployees))
	for i, e := range stats.TopEmployees {
		topEmployees[i] = &analyticsv1.TopEmployee{
			EmployeeId:      e.ID,
			EfficiencyScore: float64(e.EfficiencyScore),
			TasksCompleted:  e.TaskCompleted,
		}
	}

	problematicProjects := make([]*analyticsv1.BottomProject, len(stats.ProblematicProjects))
	for i, p := range stats.ProblematicProjects {
		problematicProjects[i] = &analyticsv1.BottomProject{
			ProjectId:   p.ProjectID,
			HealthScore: float64(p.HealthScore),
			OnTimeRate:  float64(p.OnTimeRate),
		}
	}

	return &analyticsv1.DashboardStatsResponse{
		TotalEmployees:       stats.TotalEmployees,
		ActiveEmployees:      stats.ActiveEmployees,
		TotalProjects:        stats.TotalProjects,
		ActiveProjects:       stats.ActiveProjects,
		TotalTasks:           stats.TotalTasks,
		CompletedTasks:       stats.CompletedTasks,
		OverdueTasks:         stats.OverDueTasks,
		AvgCompanyEfficiency: float64(stats.AvgCompanyEfficiency),
		AvgOnTimeRate:        float64(stats.AvgOnTimeRate),
		TopEmployees:         topEmployees,
		ProblematicProjects:  problematicProjects,
		CalculatedAt:         timestamppb.New(stats.CalculatedAt),
	}, nil
}

func (s *AnalyticsServer) SubscribeToMetricsUpdates(req *analyticsv1.SubscribeToMetricsUpdatesRequest, stream analyticsv1.AnalyticsService_SubscribeToMetricsUpdatesServer) error {
	ctx := stream.Context()
	ticker := time.NewTicker(time.Duration(req.UpdateIntervalSeconds) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if req.EmployeeId != nil {
				metrics, err := s.core.GetEmployeeMetrics(ctx, *req.EmployeeId)
				if err != nil {
					log.Printf("failed to get employee metrics for streaming: %v", err)
					continue
				}

				update := &analyticsv1.MetricsUpdate{
					Update: &analyticsv1.MetricsUpdate_EmployeeMetrics{
						EmployeeMetrics: s.employeeMetricsToProto(metrics),
					},
					UpdatedAt: timestamppb.Now(),
				}

				if err := stream.Send(update); err != nil {
					log.Printf("failed to send employee metrics update: %v", err)
					return err
				}
			}

			if req.ProjectId != nil {
				metrics, err := s.core.GetProjectMetrics(ctx, *req.ProjectId)
				if err != nil {
					log.Printf("failed to get project metrics for streaming: %v", err)
					continue
				}

				update := &analyticsv1.MetricsUpdate{
					Update: &analyticsv1.MetricsUpdate_ProjectMetrics{
						ProjectMetrics: s.projectMetricsToProto(metrics),
					},
					UpdatedAt: timestamppb.Now(),
				}

				if err := stream.Send(update); err != nil {
					log.Printf("failed to send project metrics update: %v", err)
					return err
				}
			}
		}
	}
}
