package core

import (
	"context"
	"sync"

	"github.com/foksdanilka34-maker/F5ProjectUsersControl/AnalyticsService/internal/app/analytics"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/AnalyticsService/internal/app/analytics/repo"
)

type Core struct {
	storage *repo.Storage
	mx *sync.RWMutex
}

func NewCore(storage *repo.Storage)*Core {
	return &Core{
		storage: storage,
		mx: &sync.RWMutex{},
	}
}

type CoreLogic interface {
	GetEmployeeMetrics(ctx context.Context, emplID string) (*analytics.EmployeeMetrics, error)
	SaveEmployeeMetrics(ctx context.Context, metrics *analytics.EmployeeMetrics) error
}

func (c *Core) GetEmployeeMetrics(ctx context.Context, emplID string) (*analytics.EmployeeMetrics, error) {
	return c.storage.GetEmployeeMetrics(ctx, emplID)
}

func (c *Core) SaveEmployeeMetrics(ctx context.Context, metrics *analytics.EmployeeMetrics) error {
	return c.storage.SaveEmployeeMetrics(ctx, nil, metrics)
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