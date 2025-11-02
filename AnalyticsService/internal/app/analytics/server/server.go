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
	core *core.Core
}

func NewAnalyticsServer(core *core.Core) *AnalyticsServer {
	return &AnalyticsServer{
		core: core,
	}
}

func (s *AnalyticsServer) Register(grpcServer *grpc.Server) {
	analyticsv1.RegisterAnalyticsServiceServer(grpcServer, s)
}

func (s *AnalyticsServer) GetEmployeeMetrics(ctx context.Context, req *analyticsv1.GetEmployeeMetricsRequest) (*analyticsv1.EmployeeMetricsResponse, error) {
	startDate := time.Now().AddDate(0, 0, -7)
	endDate := time.Now()

	if req.StartDate != nil {
		startDate = req.StartDate.AsTime()
	}
	if req.EndDate != nil {
		endDate = req.EndDate.AsTime()
	}

	metrics, err := s.core.GetEmployeeMetrics(ctx, req.EmployeeId, startDate, endDate)
	if err != nil {
		log.Printf("failed to get employee metrics: %v", err)
		return nil, err
	}

	protoMetrics := make([]*analyticsv1.EmployeeMetrics, len(metrics))
	for i, m := range metrics {
		protoMetrics[i] = s.employeeMetricsToProto(m)
	}

	return &analyticsv1.EmployeeMetricsResponse{
		Metrics:      protoMetrics,
		CalculatedAt: timestamppb.Now(),
	}, nil
}

func (s *AnalyticsServer) ListEmployeeMetrics(ctx context.Context, req *analyticsv1.ListEmployeeMetricsRequest) (*analyticsv1.ListEmployeeMetricsResponse, error) {
	startDate := time.Now().AddDate(0, -1, 0)
	endDate := time.Now()

	if req.StartDate != nil {
		startDate = req.StartDate.AsTime()
	}
	if req.EndDate != nil {
		endDate = req.EndDate.AsTime()
	}

	deptID := ""
	if req.DepartmentId != nil {
		deptID = *req.DepartmentId
	}

	metrics, totalCount, err := s.core.ListEmployeeMetrics(ctx, req.PageSize, req.PageNumber, deptID, startDate, endDate)
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
	startDate := time.Now().AddDate(0, 0, -7)
	endDate := time.Now()

	if req.StartDate != nil {
		startDate = req.StartDate.AsTime()
	}
	if req.EndDate != nil {
		endDate = req.EndDate.AsTime()
	}

	deptID := ""
	if req.DepartmentId != nil {
		deptID = *req.DepartmentId
	}

	metrics, err := s.core.GetTopPerformers(ctx, req.Limit, deptID, startDate, endDate)
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

func (s *AnalyticsServer) GetProjectMetrics(ctx context.Context, req *analyticsv1.GetProjectMetricsRequest) (*analyticsv1.ProjectMetricsResponse, error) {
	startDate := time.Now().AddDate(0, 0, -7)
	endDate := time.Now()

	if req.StartDate != nil {
		startDate = req.StartDate.AsTime()
	}
	if req.EndDate != nil {
		endDate = req.EndDate.AsTime()
	}

	metrics, err := s.core.GetProjectMetrics(ctx, req.ProjectId, startDate, endDate)
	if err != nil {
		log.Printf("failed to get project metrics: %v", err)
		return nil, err
	}

	protoMetrics := make([]*analyticsv1.ProjectMetrics, len(metrics))
	for i, m := range metrics {
		protoMetrics[i] = s.projectMetricsToProto(m)
	}

	return &analyticsv1.ProjectMetricsResponse{
		Metrics:      protoMetrics,
		CalculatedAt: timestamppb.Now(),
	}, nil
}

func (s *AnalyticsServer) ListProjectMetrics(ctx context.Context, req *analyticsv1.ListProjectMetricsRequest) (*analyticsv1.ListProjectMetricsResponse, error) {
	startDate := time.Now().AddDate(0, -1, 0)
	endDate := time.Now()

	if req.StartDate != nil {
		startDate = req.StartDate.AsTime()
	}
	if req.EndDate != nil {
		endDate = req.EndDate.AsTime()
	}

	managerID := ""
	if req.ManagerId != nil {
		managerID = *req.ManagerId
	}

	metrics, totalCount, err := s.core.ListProjectMetrics(ctx, req.PageSize, req.PageNumber, managerID, startDate, endDate)
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

func (s *AnalyticsServer) GetDepartmentMetrics(ctx context.Context, req *analyticsv1.GetDepartmentMetricsRequest) (*analyticsv1.DepartmentMetricsResponse, error) {
	startDate := time.Now().AddDate(0, 0, -7)
	endDate := time.Now()

	if req.StartDate != nil {
		startDate = req.StartDate.AsTime()
	}
	if req.EndDate != nil {
		endDate = req.EndDate.AsTime()
	}

	metrics, err := s.core.GetDepartmentMetrics(ctx, req.DepartmentId, startDate, endDate)
	if err != nil {
		log.Printf("failed to get department metrics: %v", err)
		return nil, err
	}

	protoMetrics := make([]*analyticsv1.DepartmentMetrics, len(metrics))
	for i, m := range metrics {
		protoMetrics[i] = s.departmentMetricsToProto(m)
	}

	return &analyticsv1.DepartmentMetricsResponse{
		Metrics:      protoMetrics,
		CalculatedAt: timestamppb.Now(),
	}, nil
}

func (s *AnalyticsServer) ListDepartmentMetrics(ctx context.Context, req *analyticsv1.ListDepartmentMetricsRequest) (*analyticsv1.ListDepartmentMetricsResponse, error) {
	startDate := time.Now().AddDate(0, -1, 0)
	endDate := time.Now()

	if req.StartDate != nil {
		startDate = req.StartDate.AsTime()
	}
	if req.EndDate != nil {
		endDate = req.EndDate.AsTime()
	}

	metrics, totalCount, err := s.core.ListDepartmentMetrics(ctx, req.PageSize, req.PageNumber, startDate, endDate)
	if err != nil {
		log.Printf("failed to list department metrics: %v", err)
		return nil, err
	}

	protoMetrics := make([]*analyticsv1.DepartmentMetrics, len(metrics))
	for i, m := range metrics {
		protoMetrics[i] = s.departmentMetricsToProto(m)
	}

	return &analyticsv1.ListDepartmentMetricsResponse{
		Metrics:    protoMetrics,
		TotalCount: totalCount,
	}, nil
}

func (s *AnalyticsServer) SubscribeToMetricsUpdates(req *analyticsv1.SubscribeToMetricsUpdatesRequest, stream analyticsv1.AnalyticsService_SubscribeToMetricsUpdatesServer) error {
	ticker := time.NewTicker(time.Duration(req.UpdateIntervalSeconds) * time.Second)
	defer ticker.Stop()

	ctx := stream.Context()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			var update *analyticsv1.MetricsUpdate

			if req.EmployeeId != nil && *req.EmployeeId != "" {
				metrics, err := s.core.GetEmployeeMetrics(ctx, *req.EmployeeId, time.Now().AddDate(0, 0, -1), time.Now())
				if err != nil {
					log.Printf("failed to get employee metrics for streaming: %v", err)
					continue
				}

				if len(metrics) > 0 {
					update = &analyticsv1.MetricsUpdate{
						Update:    &analyticsv1.MetricsUpdate_EmployeeMetrics{EmployeeMetrics: s.employeeMetricsToProto(metrics[0])},
						UpdatedAt: timestamppb.Now(),
					}
				}
			} else if req.ProjectId != nil && *req.ProjectId != "" {
				metrics, err := s.core.GetProjectMetrics(ctx, *req.ProjectId, time.Now().AddDate(0, 0, -1), time.Now())
				if err != nil {
					log.Printf("failed to get project metrics for streaming: %v", err)
					continue
				}

				if len(metrics) > 0 {
					update = &analyticsv1.MetricsUpdate{
						Update:    &analyticsv1.MetricsUpdate_ProjectMetrics{ProjectMetrics: s.projectMetricsToProto(metrics[0])},
						UpdatedAt: timestamppb.Now(),
					}
				}
			} else if req.DepartmentId != nil && *req.DepartmentId != "" {
				metrics, err := s.core.GetDepartmentMetrics(ctx, *req.DepartmentId, time.Now().AddDate(0, 0, -1), time.Now())
				if err != nil {
					log.Printf("failed to get department metrics for streaming: %v", err)
					continue
				}

				if len(metrics) > 0 {
					update = &analyticsv1.MetricsUpdate{
						Update:    &analyticsv1.MetricsUpdate_DepartmentMetrics{DepartmentMetrics: s.departmentMetricsToProto(metrics[0])},
						UpdatedAt: timestamppb.Now(),
					}
				}
			}

			if update != nil {
				if err := stream.Send(update); err != nil {
					log.Printf("failed to send metrics update: %v", err)
					return err
				}
			}
		}
	}
}

func (s *AnalyticsServer) GetProductivityTrends(ctx context.Context, req *analyticsv1.GetProductivityTrendsRequest) (*analyticsv1.ProductivityTrendsResponse, error) {
	deptID := ""
	if req.DepartmentId != nil {
		deptID = *req.DepartmentId
	}

	empID := ""
	if req.EmployeeId != nil {
		empID = *req.EmployeeId
	}

	trends, err := s.core.CalculateProductivityTrends(ctx, req.Period, req.Limit, deptID, empID)
	if err != nil {
		log.Printf("failed to calculate productivity trends: %v", err)
		return nil, err
	}

	protoTrends := make([]*analyticsv1.ProductivityTrendEntry, len(trends))
	for i, trend := range trends {
		protoTrends[i] = &analyticsv1.ProductivityTrendEntry{
			Date:                 timestamppb.New(trend["date"].(time.Time)),
			AvgEfficiency:        trend["avg_efficiency"].(float64),
			TotalTasksCompleted:  trend["total_tasks_completed"].(int32),
			TotalEmployeesActive: trend["total_employees_active"].(int32),
		}
	}

	return &analyticsv1.ProductivityTrendsResponse{
		Entries: protoTrends,
		Period:  req.Period,
	}, nil
}

func (s *AnalyticsServer) GetCompletionRateTrends(ctx context.Context, req *analyticsv1.GetCompletionRateTrendsRequest) (*analyticsv1.CompletionRateTrendsResponse, error) {
	projID := ""
	if req.ProjectId != nil {
		projID = *req.ProjectId
	}

	deptID := ""
	if req.DepartmentId != nil {
		deptID = *req.DepartmentId
	}

	trends, err := s.core.CalculateCompletionRateTrends(ctx, req.Period, req.Limit, projID, deptID)
	if err != nil {
		log.Printf("failed to calculate completion rate trends: %v", err)
		return nil, err
	}

	protoTrends := make([]*analyticsv1.CompletionRateTrendEntry, len(trends))
	for i, trend := range trends {
		protoTrends[i] = &analyticsv1.CompletionRateTrendEntry{
			Date:           timestamppb.New(trend["date"].(time.Time)),
			OnTimeRate:     trend["on_time_rate"].(float64),
			OverallRate:    trend["overall_rate"].(float64),
			CompletedCount: trend["completed_count"].(int32),
			OverdueCount:   trend["overdue_count"].(int32),
		}
	}

	return &analyticsv1.CompletionRateTrendsResponse{
		Entries: protoTrends,
		Period:  req.Period,
	}, nil
}

func (s *AnalyticsServer) GetDashboardStats(ctx context.Context, req *analyticsv1.GetDashboardStatsRequest) (*analyticsv1.DashboardStatsResponse, error) {
	startDate := time.Now().AddDate(0, -1, 0)
	endDate := time.Now()

	if req.StartDate != nil {
		startDate = req.StartDate.AsTime()
	}
	if req.EndDate != nil {
		endDate = req.EndDate.AsTime()
	}

	stats, err := s.core.GetDashboardStats(ctx, startDate, endDate)
	if err != nil {
		log.Printf("failed to get dashboard stats: %v", err)
		return nil, err
	}

	if stats == nil {
		return &analyticsv1.DashboardStatsResponse{
			CalculatedAt: timestamppb.Now(),
		}, nil
	}

	topEmployees := make([]*analyticsv1.TopEmployee, 0)
	if empList, ok := stats["top_employees"].([]*analytics.EmployeeMetrics); ok {
		for _, emp := range empList {
			topEmployees = append(topEmployees, &analyticsv1.TopEmployee{
				EmployeeId:      emp.EmployeeID,
				Name:            emp.EmployeeName,
				EfficiencyScore: emp.EfficiencyScore,
				TasksCompleted:  emp.TasksCompleted,
			})
		}
	}

	return &analyticsv1.DashboardStatsResponse{
		TotalEmployees:       stats["total_employees"].(int32),
		ActiveEmployees:      stats["active_employees"].(int32),
		TotalProjects:        stats["total_projects"].(int32),
		ActiveProjects:       stats["active_projects"].(int32),
		TotalTasks:           stats["total_tasks"].(int32),
		CompletedTasks:       stats["completed_tasks"].(int32),
		OverdueTasks:         stats["overdue_tasks"].(int32),
		AvgCompanyEfficiency: stats["avg_efficiency"].(float64),
		AvgOnTimeRate:        stats["avg_on_time_rate"].(float64),
		TopEmployees:         topEmployees,
		CalculatedAt:         timestamppb.Now(),
	}, nil
}

func (s *AnalyticsServer) employeeMetricsToProto(m *analytics.EmployeeMetrics) *analyticsv1.EmployeeMetrics {
	return &analyticsv1.EmployeeMetrics{
		EmployeeId:             m.EmployeeID,
		EmployeeName:           m.EmployeeName,
		Department:             m.DepartmentID,
		Position:               m.PositionID,
		MetricDate:             timestamppb.New(m.MetricDate),
		TasksCompleted:         m.TasksCompleted,
		TasksAssigned:          m.TasksAssigned,
		AvgCompletionTimeHours: m.AvgCompletionTimeHours,
		OnTimeCompletionRate:   m.OnTimeCompletionRate,
		AvgTaskPriority:        m.AvgTaskPriority,
		SkillsUsed:             m.SkillsUsed,
		ProjectsInvolved:       m.ProjectsInvolved,
		EfficiencyScore:        m.EfficiencyScore,
	}
}

func (s *AnalyticsServer) projectMetricsToProto(m *analytics.ProjectMetrics) *analyticsv1.ProjectMetrics {
	return &analyticsv1.ProjectMetrics{
		ProjectId:            m.ProjectID,
		ProjectName:          m.ProjectName,
		ManagerId:            m.ManagerID,
		ManagerName:          m.ManagerName,
		MetricDate:           timestamppb.New(m.MetricDate),
		TotalTasks:           m.TotalTasks,
		CompletedTasks:       m.CompletedTasks,
		InProgressTasks:      m.InProgressTasks,
		OverdueTasks:         m.OverdueTasks,
		CompletionRate:       m.CompletionRate,
		OnTimeCompletionRate: m.OnTimeCompletionRate,
		TeamSize:             m.TeamSize,
		AvgTaskDurationHours: m.AvgTaskDurationHours,
		ProjectHealthScore:   m.ProjectHealthScore,
	}
}

func (s *AnalyticsServer) departmentMetricsToProto(m *analytics.DepartmentMetrics) *analyticsv1.DepartmentMetrics {
	return &analyticsv1.DepartmentMetrics{
		DepartmentId:             m.DepartmentID,
		DepartmentName:           m.DepartmentName,
		MetricDate:               timestamppb.New(m.MetricDate),
		TotalEmployees:           m.TotalEmployees,
		ActiveProjects:           m.ActiveProjects,
		TotalTasks:               m.TotalTasks,
		CompletedTasks:           m.CompletedTasks,
		AvgEmployeeEfficiency:    m.AvgEmployeeEfficiency,
		DepartmentCompletionRate: m.DepartmentCompletionRate,
		DepartmentOnTimeRate:     m.DepartmentOnTimeRate,
		DepartmentHealthScore:    m.DepartmentHealthScore,
	}
}
