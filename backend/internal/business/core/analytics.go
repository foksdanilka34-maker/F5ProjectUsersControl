package core

import (
	"context"
	"time"

	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/business/repo"
)

type AnalyticsRepository interface {
	GetSummary(ctx context.Context) (*repo.AnalyticsSummary, error)
	GetProjectAnalytics(ctx context.Context, projectID int64) (*repo.ProjectAnalytics, error)
	GetEmployeeAnalytics(ctx context.Context, userID int64) (*repo.EmployeeAnalytics, error)
	GetTasksTimeSeries(ctx context.Context, startDate, endDate time.Time, projectID int64) ([]repo.TimeSeriesPoint, error)
	GetCompletedTasksTimeSeries(ctx context.Context, startDate, endDate time.Time, projectID int64) ([]repo.TimeSeriesPoint, error)
	GetTaskDistribution(ctx context.Context, projectID int64) ([]repo.TaskDistribution, error)
	GetPriorityDistribution(ctx context.Context, projectID int64) ([]repo.PriorityDistribution, error)
}

type AnalyticsService struct {
	repo AnalyticsRepository
}

func NewAnalyticsService(repo AnalyticsRepository) *AnalyticsService {
	return &AnalyticsService{repo: repo}
}

func (s *AnalyticsService) GetSummary(ctx context.Context) (*repo.AnalyticsSummary, error) {
	return s.repo.GetSummary(ctx)
}

func (s *AnalyticsService) GetProjectAnalytics(ctx context.Context, projectID int64) (*repo.ProjectAnalytics, error) {
	return s.repo.GetProjectAnalytics(ctx, projectID)
}

func (s *AnalyticsService) GetEmployeeAnalytics(ctx context.Context, userID int64) (*repo.EmployeeAnalytics, error) {
	return s.repo.GetEmployeeAnalytics(ctx, userID)
}

type TimeSeriesRequest struct {
	StartDate time.Time
	EndDate   time.Time
	ProjectID int64
}

func (s *AnalyticsService) GetTasksCreatedTimeSeries(ctx context.Context, req *TimeSeriesRequest) ([]repo.TimeSeriesPoint, error) {
	start := req.StartDate
	end := req.EndDate
	if start.IsZero() {
		start = time.Now().AddDate(0, -1, 0) // Last month
	}
	if end.IsZero() {
		end = time.Now()
	}
	return s.repo.GetTasksTimeSeries(ctx, start, end, req.ProjectID)
}

func (s *AnalyticsService) GetTasksCompletedTimeSeries(ctx context.Context, req *TimeSeriesRequest) ([]repo.TimeSeriesPoint, error) {
	start := req.StartDate
	end := req.EndDate
	if start.IsZero() {
		start = time.Now().AddDate(0, -1, 0)
	}
	if end.IsZero() {
		end = time.Now()
	}
	return s.repo.GetCompletedTasksTimeSeries(ctx, start, end, req.ProjectID)
}

func (s *AnalyticsService) GetTaskDistribution(ctx context.Context, projectID int64) ([]repo.TaskDistribution, error) {
	return s.repo.GetTaskDistribution(ctx, projectID)
}

func (s *AnalyticsService) GetPriorityDistribution(ctx context.Context, projectID int64) ([]repo.PriorityDistribution, error) {
	return s.repo.GetPriorityDistribution(ctx, projectID)
}

type Dashboard struct {
	Summary              *repo.AnalyticsSummary
	TaskStatusDist       []repo.TaskDistribution
	TaskPriorityDist     []repo.PriorityDistribution
	TasksCreatedSeries   []repo.TimeSeriesPoint
	TasksCompletedSeries []repo.TimeSeriesPoint
}

func (s *AnalyticsService) GetDashboard(ctx context.Context, projectID int64) (*Dashboard, error) {
	summary, err := s.repo.GetSummary(ctx)
	if err != nil {
		return nil, err
	}

	statusDist, err := s.repo.GetTaskDistribution(ctx, projectID)
	if err != nil {
		return nil, err
	}

	priorityDist, err := s.repo.GetPriorityDistribution(ctx, projectID)
	if err != nil {
		return nil, err
	}

	start := time.Now().AddDate(0, -1, 0)
	end := time.Now()

	createdSeries, err := s.repo.GetTasksTimeSeries(ctx, start, end, projectID)
	if err != nil {
		return nil, err
	}

	completedSeries, err := s.repo.GetCompletedTasksTimeSeries(ctx, start, end, projectID)
	if err != nil {
		return nil, err
	}

	return &Dashboard{
		Summary:              summary,
		TaskStatusDist:       statusDist,
		TaskPriorityDist:     priorityDist,
		TasksCreatedSeries:   createdSeries,
		TasksCompletedSeries: completedSeries,
	}, nil
}


