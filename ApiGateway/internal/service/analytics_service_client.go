package service

import (
	"context"
	"fmt"
	"log"
	"time"

	analyticsv1 "github.com/foksdanilka34-maker/F5ProjectUsersControl/gen/go/analytics_service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type AnalyticsServiceClient struct {
	conn   *grpc.ClientConn
	client analyticsv1.AnalyticsServiceClient
}

func NewAnalyticsServiceClient(host, port string) *AnalyticsServiceClient {
	address := fmt.Sprintf("%s:%s", host, port)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(
		ctx,
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		log.Fatalf("Failed to connect to analytics service at %s: %v", address, err)
	}

	log.Printf("Successfully connected to analytics service at %s", address)

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
