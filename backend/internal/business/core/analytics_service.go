package core

import (
	"context"
	"sort"
	"time"

	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/business/dto"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/singleflight"
)

type AnalyticsRepository interface {
	GetSummary(ctx context.Context) (*dto.DashboardStatsDTO, error)
	GetProjectAnalytics(ctx context.Context, projectID int64) (*dto.ProjectMetricsDTO, error)
	GetAllProjectAnalytics(ctx context.Context) ([]dto.ProjectMetricsDTO, error)
	GetEmployeeAnalytics(ctx context.Context, userID int64) (*dto.EmployeeMetricsDTO, error)
	GetAllEmployeeAnalytics(ctx context.Context) ([]dto.EmployeeMetricsDTO, error)
	GetProductivityTrends(ctx context.Context, days int) ([]dto.ProductivityTrendEntryDTO, error)
}

type AnalyticsService struct {
	repo         AnalyticsRepository
	requestGroup singleflight.Group
}

func NewAnalyticsService(repo AnalyticsRepository) *AnalyticsService {
	return &AnalyticsService{
		repo: repo,
	}
}

// GetDashboard calculates complex dashboard metrics using Fan-Out / Fan-In with errgroup
// and protects against Cache Stampede using singleflight.Group.
func (s *AnalyticsService) GetDashboard(ctx context.Context) (dto.DashboardStatsDTO, error) {
	result, err, _ := s.requestGroup.Do("dashboard_stats", func() (interface{}, error) {
		return s.computeDashboardStats(ctx)
	})
	if err != nil {
		return dto.DashboardStatsDTO{}, err
	}
	return result.(dto.DashboardStatsDTO), nil
}

func (s *AnalyticsService) computeDashboardStats(ctx context.Context) (dto.DashboardStatsDTO, error) {
	var (
		summary             *dto.DashboardStatsDTO
		employeeMetricsList []dto.EmployeeMetricsDTO
		projectMetricsList  []dto.ProjectMetricsDTO
	)

	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		var err error
		summary, err = s.repo.GetSummary(gCtx)
		return err
	})

	g.Go(func() error {
		var err error
		employeeMetricsList, err = s.repo.GetAllEmployeeAnalytics(gCtx)
		return err
	})

	g.Go(func() error {
		var err error
		projectMetricsList, err = s.repo.GetAllProjectAnalytics(gCtx)
		return err
	})

	if err := g.Wait(); err != nil {
		return dto.DashboardStatsDTO{}, err
	}

	res := *summary
	res.CalculatedAt = time.Now()

	sort.Slice(employeeMetricsList, func(i, j int) bool {
		if employeeMetricsList[i].CompletionRate == employeeMetricsList[j].CompletionRate {
			return employeeMetricsList[i].CompletedTasks > employeeMetricsList[j].CompletedTasks
		}
		return employeeMetricsList[i].CompletionRate > employeeMetricsList[j].CompletionRate
	})

	topCount := 5
	if len(employeeMetricsList) < topCount {
		topCount = len(employeeMetricsList)
	}

	res.TopEmployees = make([]dto.TopEmployeeDTO, 0, topCount)
	for i := 0; i < topCount; i++ {
		e := employeeMetricsList[i]
		res.TopEmployees = append(res.TopEmployees, dto.TopEmployeeDTO{
			EmployeeID:     e.EmployeeID,
			CompletionRate: e.CompletionRate,
			TasksCompleted: e.CompletedTasks,
		})
	}

	res.ProblematicProjects = make([]dto.ProblematicProjectDTO, 0)
	for _, p := range projectMetricsList {
		if p.HealthStatus == "HEALTH_STATUS_AT_RISK" || p.HealthStatus == "HEALTH_STATUS_CRITICAL" {
			res.ProblematicProjects = append(res.ProblematicProjects, dto.ProblematicProjectDTO{
				ProjectID:    p.ProjectID,
				OnTimeRate:   p.OnTimeRate,
				HealthStatus: p.HealthStatus,
			})
		}
	}

	return res, nil
}

func (s *AnalyticsService) GetProjectMetrics(ctx context.Context, projectID int64) (dto.ProjectMetricsDTO, error) {
	m, err := s.repo.GetProjectAnalytics(ctx, projectID)
	if err != nil {
		return dto.ProjectMetricsDTO{}, err
	}
	if m == nil {
		return dto.ProjectMetricsDTO{}, ErrNotFound
	}
	return *m, nil
}

func (s *AnalyticsService) ListAllProjectMetrics(ctx context.Context) ([]dto.ProjectMetricsDTO, error) {
	return s.repo.GetAllProjectAnalytics(ctx)
}

func (s *AnalyticsService) GetEmployeeMetrics(ctx context.Context, userID int64) (dto.EmployeeMetricsDTO, error) {
	m, err := s.repo.GetEmployeeAnalytics(ctx, userID)
	if err != nil {
		return dto.EmployeeMetricsDTO{}, err
	}
	if m == nil {
		return dto.EmployeeMetricsDTO{}, ErrNotFound
	}
	return *m, nil
}

func (s *AnalyticsService) ListAllEmployeeMetrics(ctx context.Context) ([]dto.EmployeeMetricsDTO, error) {
	return s.repo.GetAllEmployeeAnalytics(ctx)
}

func (s *AnalyticsService) GetProductivityTrends(ctx context.Context, days int) (dto.ProductivityTrendsDTO, error) {
	if days <= 0 {
		days = 30
	}
	entries, err := s.repo.GetProductivityTrends(ctx, days)
	if err != nil {
		return dto.ProductivityTrendsDTO{}, err
	}
	return dto.ProductivityTrendsDTO{
		Entries: entries,
		Period:  "PERIOD_DAILY",
	}, nil
}
