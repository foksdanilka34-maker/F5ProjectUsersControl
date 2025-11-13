package core

import (
	"context"
	"math"
	"sync"
	"time"

	"github.com/foksdanilka34-maker/F5ProjectUsersControl/AnalyticsService/internal/app/analytics"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/AnalyticsService/internal/app/analytics/repo"
)

type Core struct {
	storage *repo.Storage
	mx      *sync.RWMutex
}

func NewCore(storage *repo.Storage) *Core {
	return &Core{
		storage: storage,
		mx:      &sync.RWMutex{},
	}
}

type CoreLogic interface {
	GetEmployeeMetrics(ctx context.Context, emplID string) (*analytics.EmployeeMetrics, error)
	SaveEmployeeMetrics(ctx context.Context, metrics *analytics.EmployeeMetrics) error
	UpdateEmployeeMetrics(ctx context.Context, emplID string, updateFunc func(*analytics.EmployeeMetrics)) error
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

	metrics.EfficiencyScore = math.Round(finalScore*100) / 100
}

func (c *Core) GetEmployeeMetrics(ctx context.Context, emplID string) (*analytics.EmployeeMetrics, error) {
	metrics, err := c.storage.GetEmployeeMetrics(ctx, nil, emplID)
	if err != nil {
		return nil, err
	}
	calculateScores(metrics)

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

	return tx.Commit(ctx)
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
	metrics, err := c.storage.GetProjectMetrics(ctx, nil, projectID)
	if err != nil {
		return nil, err
	}
	calculateProjectScores(metrics)
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

	return tx.Commit(ctx)
}

// func (c *Core) ListEmployeeMetrics(ctx context.Context, pageSize, pageNumber int32, departmentID string, startDate, endDate time.Time) ([]*analytics.EmployeeMetrics, int32, error) {
// 	return c.storage.ListEmployeeMetrics(ctx, pageSize, pageNumber, departmentID, startDate, endDate)
// }

// func (c *Core) GetTopPerformers(ctx context.Context, limit int32, departmentID string, startDate, endDate time.Time) ([]*analytics.EmployeeMetrics, error) {
// 	return c.storage.GetTopPerformers(ctx, limit, departmentID, startDate, endDate)
// }

// func (c *Core) GetProjectMetrics(ctx context.Context, projectID string, startDate, endDate time.Time) ([]*analytics.ProjectMetrics, error) {
// 	return c.storage.GetProjectMetrics(ctx, projectID, startDate, endDate)
// }

// func (c *Core) ListProjectMetrics(ctx context.Context, pageSize, pageNumber int32, managerID string, startDate, endDate time.Time) ([]*analytics.ProjectMetrics, int32, error) {
// 	return c.storage.ListProjectMetrics(ctx, pageSize, pageNumber, managerID, startDate, endDate)
// }

// func (c *Core) SaveProjectMetrics(ctx context.Context, metrics *analytics.ProjectMetrics) error {
// 	c.mu.Lock()
// 	defer c.mu.Unlock()

// 	if err := c.storage.SaveProjectMetrics(ctx, metrics); err != nil {
// 		return err
// 	}

// 	ttl := time.Hour * 24
// 	if err := c.cache.SetProjectMetricsCache(ctx, metrics, ttl); err != nil {
// 		log.Printf("failed to cache project metrics: %v", err)
// 	}

// 	return nil
// }

// func (c *Core) CalculateProductivityTrends(ctx context.Context, period string, limit int32, departmentID, employeeID string) ([]map[string]any, error) {
// 	return c.storage.CalculateProductivityTrends(ctx, period, limit, departmentID, employeeID)
// }

// func (c *Core) CalculateCompletionRateTrends(ctx context.Context, period string, limit int32, projectID, departmentID string) ([]map[string]any, error) {
// 	return c.storage.CalculateCompletionRateTrends(ctx, period, limit, projectID, departmentID)
// }

// func (c *Core) GetDashboardStats(ctx context.Context, startDate, endDate time.Time) (map[string]any, error) {
// 	return c.storage.GetDashboardStats(ctx, startDate, endDate)
// }
