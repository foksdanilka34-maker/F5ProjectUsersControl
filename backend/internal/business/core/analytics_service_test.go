package core_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/business/core"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/business/dto"
)

type mockAnalyticsRepo struct {
	summaryCalls   int32
	empCalls       int32
	projCalls      int32
	simulateSlowDB bool
}

func (m *mockAnalyticsRepo) GetSummary(ctx context.Context) (*dto.DashboardStatsDTO, error) {
	atomic.AddInt32(&m.summaryCalls, 1)
	if m.simulateSlowDB {
		time.Sleep(50 * time.Millisecond)
	}
	return &dto.DashboardStatsDTO{
		TotalProjects:   10,
		ActiveProjects:  8,
		TotalTasks:      50,
		CompletedTasks:  40,
		CompletedOnTime: 35,
	}, nil
}

func (m *mockAnalyticsRepo) GetAllEmployeeAnalytics(ctx context.Context) ([]dto.EmployeeMetricsDTO, error) {
	atomic.AddInt32(&m.empCalls, 1)
	if m.simulateSlowDB {
		time.Sleep(50 * time.Millisecond)
	}
	return []dto.EmployeeMetricsDTO{
		{EmployeeID: 1, CompletionRate: 95.0, CompletedTasks: 20},
		{EmployeeID: 2, CompletionRate: 85.0, CompletedTasks: 15},
		{EmployeeID: 3, CompletionRate: 98.0, CompletedTasks: 25},
	}, nil
}

func (m *mockAnalyticsRepo) GetAllProjectAnalytics(ctx context.Context) ([]dto.ProjectMetricsDTO, error) {
	atomic.AddInt32(&m.projCalls, 1)
	if m.simulateSlowDB {
		time.Sleep(50 * time.Millisecond)
	}
	return []dto.ProjectMetricsDTO{
		{ProjectID: 101, OnTimeRate: 90.0, HealthStatus: "HEALTH_STATUS_HEALTHY"},
		{ProjectID: 102, OnTimeRate: 40.0, HealthStatus: "HEALTH_STATUS_CRITICAL"},
	}, nil
}

func (m *mockAnalyticsRepo) GetProjectAnalytics(ctx context.Context, projectID int64) (*dto.ProjectMetricsDTO, error) {
	return &dto.ProjectMetricsDTO{ProjectID: projectID}, nil
}

func (m *mockAnalyticsRepo) GetEmployeeAnalytics(ctx context.Context, userID int64) (*dto.EmployeeMetricsDTO, error) {
	return &dto.EmployeeMetricsDTO{EmployeeID: userID}, nil
}

func (m *mockAnalyticsRepo) GetProductivityTrends(ctx context.Context, days int) ([]dto.ProductivityTrendEntryDTO, error) {
	return []dto.ProductivityTrendEntryDTO{}, nil
}

func TestAnalyticsService_FanOutFanIn(t *testing.T) {
	mockRepo := &mockAnalyticsRepo{}
	svc := core.NewAnalyticsService(mockRepo)

	ctx := context.Background()
	dashboard, err := svc.GetDashboard(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if dashboard.TotalProjects != 10 {
		t.Errorf("expected 10 total projects, got %d", dashboard.TotalProjects)
	}

	// Verify Fan-In calculated top employees correctly sorted by completion rate
	if len(dashboard.TopEmployees) != 3 {
		t.Fatalf("expected 3 top employees, got %d", len(dashboard.TopEmployees))
	}
	if dashboard.TopEmployees[0].EmployeeID != 3 { // Highest rate 98.0%
		t.Errorf("expected top employee to be ID 3, got %d", dashboard.TopEmployees[0].EmployeeID)
	}

	// Verify problematic projects filtered
	if len(dashboard.ProblematicProjects) != 1 || dashboard.ProblematicProjects[0].ProjectID != 102 {
		t.Errorf("expected problematic project 102, got %+v", dashboard.ProblematicProjects)
	}
}

func TestAnalyticsService_SingleflightDeduplication(t *testing.T) {
	mockRepo := &mockAnalyticsRepo{simulateSlowDB: true}
	svc := core.NewAnalyticsService(mockRepo)

	concurrentRequests := 30
	var wg sync.WaitGroup
	wg.Add(concurrentRequests)

	startSignal := make(chan struct{})

	for i := 0; i < concurrentRequests; i++ {
		go func() {
			defer wg.Done()
			<-startSignal
			_, err := svc.GetDashboard(context.Background())
			if err != nil {
				t.Errorf("unexpected error in concurrent request: %v", err)
			}
		}()
	}

	// Unleash all 30 requests at the exact same instant
	close(startSignal)
	wg.Wait()

	// Thanks to singleflight.Group, DB calls should be deduplicated (far fewer than 30)
	summaryCalls := atomic.LoadInt32(&mockRepo.summaryCalls)
	if summaryCalls > 2 {
		t.Errorf("expected singleflight to deduplicate calls (<=2), but got %d DB queries", summaryCalls)
	}
}
