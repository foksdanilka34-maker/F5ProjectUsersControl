package core

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/foksdanilka34-maker/F5ProjectUsersControl/AnalyticsService/internal/app/analytics"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/AnalyticsService/internal/app/analytics/repo"
)

type Core struct {
	storage *repo.Storage
	cache   *repo.RedisCache
	mu      sync.RWMutex
}

func NewCore(storage *repo.Storage, cache *repo.RedisCache) *Core {
	return &Core{
		storage: storage,
		cache:   cache,
	}
}

func (c *Core) GetEmployeeMetrics(ctx context.Context, employeeID string, startDate, endDate time.Time) ([]*analytics.EmployeeMetrics, error) {
	return c.storage.GetEmployeeMetrics(ctx, employeeID, startDate, endDate)
}

func (c *Core) ListEmployeeMetrics(ctx context.Context, pageSize, pageNumber int32, departmentID string, startDate, endDate time.Time) ([]*analytics.EmployeeMetrics, int32, error) {
	return c.storage.ListEmployeeMetrics(ctx, pageSize, pageNumber, departmentID, startDate, endDate)
}

func (c *Core) GetTopPerformers(ctx context.Context, limit int32, departmentID string, startDate, endDate time.Time) ([]*analytics.EmployeeMetrics, error) {
	return c.storage.GetTopPerformers(ctx, limit, departmentID, startDate, endDate)
}

func (c *Core) SaveEmployeeMetrics(ctx context.Context, metrics *analytics.EmployeeMetrics) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.storage.SaveEmployeeMetrics(ctx, metrics); err != nil {
		return err
	}

	ttl := time.Hour * 24
	if err := c.cache.SetEmployeeMetricsCache(ctx, metrics, ttl); err != nil {
		log.Printf("failed to cache employee metrics: %v", err)
	}

	return nil
}

func (c *Core) GetProjectMetrics(ctx context.Context, projectID string, startDate, endDate time.Time) ([]*analytics.ProjectMetrics, error) {
	return c.storage.GetProjectMetrics(ctx, projectID, startDate, endDate)
}

func (c *Core) ListProjectMetrics(ctx context.Context, pageSize, pageNumber int32, managerID string, startDate, endDate time.Time) ([]*analytics.ProjectMetrics, int32, error) {
	return c.storage.ListProjectMetrics(ctx, pageSize, pageNumber, managerID, startDate, endDate)
}

func (c *Core) SaveProjectMetrics(ctx context.Context, metrics *analytics.ProjectMetrics) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.storage.SaveProjectMetrics(ctx, metrics); err != nil {
		return err
	}

	ttl := time.Hour * 24
	if err := c.cache.SetProjectMetricsCache(ctx, metrics, ttl); err != nil {
		log.Printf("failed to cache project metrics: %v", err)
	}

	return nil
}

func (c *Core) GetDepartmentMetrics(ctx context.Context, departmentID string, startDate, endDate time.Time) ([]*analytics.DepartmentMetrics, error) {
	return c.storage.GetDepartmentMetrics(ctx, departmentID, startDate, endDate)
}

func (c *Core) ListDepartmentMetrics(ctx context.Context, pageSize, pageNumber int32, startDate, endDate time.Time) ([]*analytics.DepartmentMetrics, int32, error) {
	return c.storage.ListDepartmentMetrics(ctx, pageSize, pageNumber, startDate, endDate)
}

func (c *Core) SaveDepartmentMetrics(ctx context.Context, metrics *analytics.DepartmentMetrics) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.storage.SaveDepartmentMetrics(ctx, metrics); err != nil {
		return err
	}

	ttl := time.Hour * 24
	if err := c.cache.SetDepartmentMetricsCache(ctx, metrics, ttl); err != nil {
		log.Printf("failed to cache department metrics: %v", err)
	}

	return nil
}

func (c *Core) GetDailySnapshot(ctx context.Context, date time.Time) (*analytics.DailySnapshot, error) {
	snapshot, err := c.cache.GetDailySnapshotCache(ctx, date)
	if err != nil {
		log.Printf("failed to get snapshot from cache: %v", err)
	}
	if snapshot != nil {
		return snapshot, nil
	}

	return c.storage.GetDailySnapshot(ctx, date)
}

func (c *Core) GetDailySnapshots(ctx context.Context, startDate, endDate time.Time) ([]*analytics.DailySnapshot, error) {
	return c.storage.GetDailySnapshots(ctx, startDate, endDate)
}

func (c *Core) SaveDailySnapshot(ctx context.Context, snapshot *analytics.DailySnapshot) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.storage.SaveDailySnapshot(ctx, snapshot); err != nil {
		return err
	}

	ttl := time.Hour * 24
	if err := c.cache.SetDailySnapshotCache(ctx, snapshot, ttl); err != nil {
		log.Printf("failed to cache daily snapshot: %v", err)
	}

	return nil
}

func (c *Core) CalculateProductivityTrends(ctx context.Context, period string, limit int32, departmentID, employeeID string) ([]map[string]any, error) {
	startDate, endDate := c.getDateRangeByPeriod(period, limit)

	snapshots, err := c.storage.GetDailySnapshots(ctx, startDate, endDate)
	if err != nil {
		return nil, err
	}

	trends := []map[string]any{}
	for _, snapshot := range snapshots {
		trend := map[string]any{
			"date":                   snapshot.SnapshotDate,
			"avg_efficiency":         snapshot.AvgCompanyEfficiency,
			"total_tasks_completed":  snapshot.CompletedTasks,
			"total_employees_active": snapshot.ActiveEmployees,
		}
		trends = append(trends, trend)
	}

	return trends, nil
}

func (c *Core) CalculateCompletionRateTrends(ctx context.Context, period string, limit int32, projectID, departmentID string) ([]map[string]any, error) {
	startDate, endDate := c.getDateRangeByPeriod(period, limit)

	snapshots, err := c.storage.GetDailySnapshots(ctx, startDate, endDate)
	if err != nil {
		return nil, err
	}

	trends := []map[string]any{}
	for _, snapshot := range snapshots {
		trend := map[string]any{
			"date":            snapshot.SnapshotDate,
			"on_time_rate":    snapshot.AvgOnTimeRate,
			"overall_rate":    calculateOverallCompletionRate(snapshot.CompletedTasks, snapshot.TotalTasks),
			"completed_count": snapshot.CompletedTasks,
			"overdue_count":   snapshot.OverdueTasks,
		}
		trends = append(trends, trend)
	}

	return trends, nil
}

func (c *Core) GetDashboardStats(ctx context.Context, startDate, endDate time.Time) (map[string]any, error) {
	snapshots, err := c.storage.GetDailySnapshots(ctx, startDate, endDate)
	if err != nil {
		return nil, err
	}

	if len(snapshots) == 0 {
		return nil, nil
	}

	latest := snapshots[0]

	topEmployees, err := c.storage.GetTopPerformers(ctx, 5, "", startDate, endDate)
	if err != nil {
		log.Printf("failed to get top employees: %v", err)
	}

	stats := map[string]any{
		"total_employees":  latest.TotalEmployees,
		"active_employees": latest.ActiveEmployees,
		"total_projects":   latest.TotalProjects,
		"active_projects":  latest.ActiveProjects,
		"total_tasks":      latest.TotalTasks,
		"completed_tasks":  latest.CompletedTasks,
		"overdue_tasks":    latest.OverdueTasks,
		"avg_efficiency":   latest.AvgCompanyEfficiency,
		"avg_on_time_rate": latest.AvgOnTimeRate,
		"calculated_at":    latest.UpdatedAt,
		"top_employees":    topEmployees,
	}

	return stats, nil
}

func (c *Core) getDateRangeByPeriod(period string, limit int32) (time.Time, time.Time) {
	now := time.Now()
	endDate := now
	var startDate time.Time

	switch period {
	case "DAILY":
		startDate = now.AddDate(0, 0, -int(limit))
	case "WEEKLY":
		startDate = now.AddDate(0, 0, -7*int(limit))
	case "MONTHLY":
		startDate = now.AddDate(0, -int(limit), 0)
	default:
		startDate = now.AddDate(0, 0, -30)
	}

	return startDate, endDate
}

func calculateOverallCompletionRate(completed, total int32) float64 {
	if total == 0 {
		return 0
	}
	return float64(completed) / float64(total) * 100
}
