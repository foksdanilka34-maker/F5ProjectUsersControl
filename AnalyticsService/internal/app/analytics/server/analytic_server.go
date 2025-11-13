package server

import (
	"context"
	"log"

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
		Metrics:      metricsResponse,
	}, nil
}

func (s *AnalyticsServer) employeeMetricsToProto(m *analytics.EmployeeMetrics) *analyticsv1.EmployeeMetrics {
	return &analyticsv1.EmployeeMetrics{
		EmployeeId:           m.EmployeeID,
		MetricDate:           timestamppb.New(m.MetricDate),
		AssignedTasks:        m.AssignedTasks,
		CompletedTasks:       m.CompletedTasks,
		InProgressTasks:      m.InProgressTasks,
		OverdueTasks:         m.OverdueTasks,
		OnTimeCompletedTasks: m.OnTimeCompletionTask,
		TotalTaskDurationSeconds: float64(m.TotalTaskDurationSeconds),

		TaskCompletionRate: &m.TaskCompletionRate,
		OnTimeCompletionRate: &m.OnTimeCompletionRate,
		EfficiencyScore: &m.EfficiencyScore,
	}
}

