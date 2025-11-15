package service

import (
	"context"
	"time"

	analyticsv1 "github.com/foksdanilka34-maker/F5ProjectUsersControl/gen/go/analytics_service"
	"google.golang.org/grpc"
)

type AnalyticsServiceClient struct {
	conn   *grpc.ClientConn
	client analyticsv1.AnalyticsServiceClient
}

func NewAnalyticsServiceClient(host, port string) *AnalyticsServiceClient {
	conn := dialGRPC("analytics service", host, port, 10*time.Second)
	return &AnalyticsServiceClient{
		conn:   conn,
		client: analyticsv1.NewAnalyticsServiceClient(conn),
	}
}

func (c *AnalyticsServiceClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *AnalyticsServiceClient) GetEmployeeMetrics(ctx context.Context, req *analyticsv1.GetEmployeeMetricsRequest) (*analyticsv1.EmployeeMetricsResponse, error) {
	return c.client.GetEmployeeMetrics(ctx, req)
}

func (c *AnalyticsServiceClient) ListEmployeeMetrics(ctx context.Context, req *analyticsv1.ListEmployeeMetricsRequest) (*analyticsv1.ListEmployeeMetricsResponse, error) {
	return c.client.ListEmployeeMetrics(ctx, req)
}

func (c *AnalyticsServiceClient) GetTopPerformers(ctx context.Context, req *analyticsv1.GetTopPerformersRequest) (*analyticsv1.ListEmployeeMetricsResponse, error) {
	return c.client.GetTopPerformers(ctx, req)
}

func (c *AnalyticsServiceClient) GetProjectMetrics(ctx context.Context, req *analyticsv1.GetProjectMetricsRequest) (*analyticsv1.ProjectMetricsResponse, error) {
	return c.client.GetProjectMetrics(ctx, req)
}

func (c *AnalyticsServiceClient) ListProjectMetrics(ctx context.Context, req *analyticsv1.ListProjectMetricsRequest) (*analyticsv1.ListProjectMetricsResponse, error) {
	return c.client.ListProjectMetrics(ctx, req)
}

func (c *AnalyticsServiceClient) GetProductivityTrends(ctx context.Context, req *analyticsv1.GetProductivityTrendsRequest) (*analyticsv1.ProductivityTrendsResponse, error) {
	return c.client.GetProductivityTrends(ctx, req)
}

func (c *AnalyticsServiceClient) GetCompletionRateTrends(ctx context.Context, req *analyticsv1.GetCompletionRateTrendsRequest) (*analyticsv1.CompletionRateTrendsResponse, error) {
	return c.client.GetCompletionRateTrends(ctx, req)
}

func (c *AnalyticsServiceClient) GetDashboardStats(ctx context.Context, req *analyticsv1.GetDashboardStatsRequest) (*analyticsv1.DashboardStatsResponse, error) {
	return c.client.GetDashboardStats(ctx, req)
}
