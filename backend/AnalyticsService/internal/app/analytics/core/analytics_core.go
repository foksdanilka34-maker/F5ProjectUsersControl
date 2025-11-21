package core

import (
	"context"
	"fmt"
	"log"
	"math"
	"sync"
	"time"

	"github.com/foksdanilka34-maker/F5ProjectUsersControl/AnalyticsService/internal/app/analytics"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/AnalyticsService/internal/app/analytics/cache"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/AnalyticsService/internal/app/analytics/repo"
)

type Core struct {
	storage *repo.Storage
	cache   cache.Cache
	mx      *sync.RWMutex
}

const (
	employeeKeyFmt = "metrics:employee:%s"
	projectKeyFmt  = "metrics:project:%s"
	cacheTTL       = 15 * time.Minute
)

func NewCore(storage *repo.Storage, cacheLayer cache.Cache) *Core {
	return &Core{
		storage: storage,
		cache:   cacheLayer,
		mx:      &sync.RWMutex{},
	}
}

type CoreLogic interface {
	GetEmployeeMetrics(ctx context.Context, emplID string) (*analytics.EmployeeMetrics, error)
	SaveEmployeeMetrics(ctx context.Context, metrics *analytics.EmployeeMetrics) error
	UpdateEmployeeMetrics(ctx context.Context, emplID string, updateFunc func(*analytics.EmployeeMetrics)) error
	GetProjectMetrics(ctx context.Context, projectID string) (*analytics.ProjectMetrics, error)
	UpdateProjectMetrics(ctx context.Context, projectID string, updateFunc func(*analytics.ProjectMetrics)) error
	ListEmployeeMetrics(ctx context.Context, req *analytics.ListEmployeeMetrics) ([]*analytics.EmployeeMetrics, int32, error)
	GetTopPerformers(ctx context.Context, limit int32, startDate, endDate *time.Time) ([]*analytics.EmployeeMetrics, error)
	ListProjectMetrics(ctx context.Context, req *analytics.ListProjectMetrics) ([]*analytics.ProjectMetrics, int32, error)
	GetProductivityTrends(ctx context.Context, req *analytics.ProductivityTrends) (*analytics.ProductivityTrendsResp, error)
	GetCompletionRateTrends(ctx context.Context, req *analytics.ComletionRateTrends) (*analytics.CompletionRateTrendResp, error)
	GetDashboardStats(ctx context.Context, startDate, endDate *time.Time) (*analytics.DashboardStats, error)
}

func calculateScores(metrics *analytics.EmployeeMetrics) {
	if metrics.AssignedTasks > 0 {
		metrics.TaskCompletionRate = (float64(metrics.CompletedTasks) / float64(metrics.AssignedTasks)) * 100
	}
	if metrics.CompletedTasks > 0 {
		metrics.OnTimeCompletionRate = (float64(metrics.OnTimeCompletionTask) / float64(metrics.CompletedTasks)) * 100
	}

	speedBonus := 1.0
	if metrics.CompletedTasks > 0 {
		avgDurationSeconds := float64(metrics.TotalTaskDurationSeconds) / float64(metrics.CompletedTasks)
		const normalTaskDuration = 28800.0
		speedBonus = normalTaskDuration / (avgDurationSeconds + 1)
		if speedBonus > 1.5 {
			speedBonus = 1.5
		}
	}

	overduePenalty := float64(metrics.OverdueTasks) * 5.0

	baseScore := (metrics.OnTimeCompletionRate * 0.7) + (metrics.TaskCompletionRate * 0.3)

	finalScore := (baseScore * speedBonus) - overduePenalty

	if finalScore < 0 {
		finalScore = 0
	}
	if finalScore > 100 {
		finalScore = 100
	}

	metrics.EfficiencyScore = math.Round(finalScore*100) / 100
}

func (c *Core) GetEmployeeMetrics(ctx context.Context, emplID string) (*analytics.EmployeeMetrics, error) {
	cacheKey := fmt.Sprintf(employeeKeyFmt, emplID)
	if c.cache != nil {
		cached := &analytics.EmployeeMetrics{}
		found, err := c.cache.Get(ctx, cacheKey, cached)
		if err != nil {
			log.Printf("employee metrics cache get failed for %s: %v", emplID, err)
		} else if found {
			return cached, nil
		}
	}

	metrics, err := c.storage.GetEmployeeMetrics(ctx, nil, emplID)
	if err != nil {
		return nil, err
	}
	calculateScores(metrics)

	if c.cache != nil {
		if err := c.cache.Set(ctx, cacheKey, metrics, cacheTTL); err != nil {
			log.Printf("employee metrics cache set failed for %s: %v", emplID, err)
		}
	}

	return metrics, nil
}

func (c *Core) SaveEmployeeMetrics(ctx context.Context, metrics *analytics.EmployeeMetrics) error {
	return c.storage.SaveEmployeeMetrics(ctx, nil, metrics)
}

func (c *Core) UpdateEmployeeMetrics(ctx context.Context, emplID string, updateFunc func(*analytics.EmployeeMetrics)) error {
	tx, err := c.storage.BeginTransaction(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	metrics, err := c.storage.GetEmployeeMetrics(ctx, tx, emplID)
	if err != nil {
		now := time.Now()
		metrics = &analytics.EmployeeMetrics{
			EmployeeID: emplID,
			MetricDate: now,
		}
	}

	updateFunc(metrics)

	if err := c.storage.SaveEmployeeMetrics(ctx, tx, metrics); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	if c.cache != nil {
		cacheKey := fmt.Sprintf(employeeKeyFmt, emplID)
		calculateScores(metrics)
		if err := c.cache.Set(ctx, cacheKey, metrics, cacheTTL); err != nil {
			log.Printf("employee metrics cache refresh failed for %s: %v", emplID, err)
		}
	}

	return nil
}

func calculateProjectScores(metrics *analytics.ProjectMetrics) {
	if metrics.TotalTasks > 0 {
		metrics.DeliveryPerformance = (float64(metrics.CompletedTasks) / float64(metrics.TotalTasks)) * 100
	}
	if metrics.CompletedTasks > 0 {
		metrics.SchedulePerformance = (float64(metrics.OnTimeCompletedTasks) / float64(metrics.CompletedTasks)) * 100
	}

	metrics.Velocity = float64(metrics.CompletedTasks)

	baseScore := (metrics.SchedulePerformance * 0.6) + (metrics.DeliveryPerformance * 0.4)
	penalty := float64(metrics.OverdueTasks) * 5.0
	finalScore := baseScore - penalty
	if finalScore < 0 {
		finalScore = 0
	}
	metrics.HealthIndex = finalScore

	if metrics.HealthIndex > 80 {
		metrics.HealthStatus = analytics.HEALTH_STATUS_HEALTHY
	} else if metrics.HealthIndex > 50 {
		metrics.HealthStatus = analytics.HEALTH_STATUS_AT_RISK
	} else {
		metrics.HealthStatus = analytics.HEALTH_STATUS_CRITICAL
	}

	metrics.IsAtRisk = metrics.HealthIndex <= 50
}

func (c *Core) GetProjectMetrics(ctx context.Context, projectID string) (*analytics.ProjectMetrics, error) {
	cacheKey := fmt.Sprintf(projectKeyFmt, projectID)
	if c.cache != nil {
		cached := &analytics.ProjectMetrics{}
		found, err := c.cache.Get(ctx, cacheKey, cached)
		if err != nil {
			log.Printf("project metrics cache get failed for %s: %v", projectID, err)
		} else if found {
			return cached, nil
		}
	}

	metrics, err := c.storage.GetProjectMetrics(ctx, nil, projectID)
	if err != nil {
		return nil, err
	}
	calculateProjectScores(metrics)

	if c.cache != nil {
		if err := c.cache.Set(ctx, cacheKey, metrics, cacheTTL); err != nil {
			log.Printf("project metrics cache set failed for %s: %v", projectID, err)
		}
	}
	return metrics, nil
}

func (c *Core) UpdateProjectMetrics(ctx context.Context, projectID string, updateFunc func(*analytics.ProjectMetrics)) error {
	tx, err := c.storage.BeginTransaction(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	metrics, err := c.storage.GetProjectMetrics(ctx, tx, projectID)
	if err != nil {
		now := time.Now()
		metrics = &analytics.ProjectMetrics{
			ProjectID:  projectID,
			MetricDate: now,
		}
	}

	updateFunc(metrics)

	if err := c.storage.SaveProjectMetrics(ctx, tx, metrics); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	if c.cache != nil {
		cacheKey := fmt.Sprintf(projectKeyFmt, projectID)
		calculateProjectScores(metrics)
		if err := c.cache.Set(ctx, cacheKey, metrics, cacheTTL); err != nil {
			log.Printf("project metrics cache refresh failed for %s: %v", projectID, err)
		}
	}

	return nil
}

func (c *Core) ListEmployeeMetrics(ctx context.Context, req *analytics.ListEmployeeMetrics) ([]*analytics.EmployeeMetrics, int32, error) {
	metrics, total, err := c.storage.ListEmployeeMetrics(ctx, req)
	if err != nil {
		return nil, 0, err
	}

	for _, m := range metrics {
		calculateScores(m)
	}

	return metrics, total, nil
}

func (c *Core) GetTopPerformers(ctx context.Context, limit int32, startDate, endDate *time.Time) ([]*analytics.EmployeeMetrics, error) {
	metrics, err := c.storage.GetTopPerformers(ctx, limit, startDate, endDate)
	if err != nil {
		return nil, err
	}

	for _, m := range metrics {
		calculateScores(m)
	}

	return metrics, nil
}

func (c *Core) ListProjectMetrics(ctx context.Context, req *analytics.ListProjectMetrics) ([]*analytics.ProjectMetrics, int32, error) {
	metrics, total, err := c.storage.ListProjectMetrics(ctx, req)
	if err != nil {
		return nil, 0, err
	}

	for _, m := range metrics {
		calculateProjectScores(m)
	}

	return metrics, total, nil
}

func (c *Core) GetProductivityTrends(ctx context.Context, req *analytics.ProductivityTrends) (*analytics.ProductivityTrendsResp, error) {
	trends, err := c.storage.GetProductivityTrends(ctx, req)
	if err != nil {
		return nil, err
	}

	result := make([]analytics.ProductivityTrend, len(trends))
	for i, t := range trends {
		result[i] = *t
	}

	return &analytics.ProductivityTrendsResp{
		Prod:   result,
		Period: req.Period,
	}, nil
}

func (c *Core) GetCompletionRateTrends(ctx context.Context, req *analytics.ComletionRateTrends) (*analytics.CompletionRateTrendResp, error) {
	trends, err := c.storage.GetCompletionRateTrends(ctx, req)
	if err != nil {
		return nil, err
	}

	result := make([]analytics.CompletionRateTrend, len(trends))
	for i, t := range trends {
		result[i] = *t
	}

	return &analytics.CompletionRateTrendResp{
		CompTrend: result,
		Period:    req.Period,
	}, nil
}

func (c *Core) GetDashboardStats(ctx context.Context, startDate, endDate *time.Time) (*analytics.DashboardStats, error) {
	stats, err := c.storage.GetDashboardStats(ctx, startDate, endDate)
	if err != nil {
		return nil, err
	}

	return stats, nil
}
