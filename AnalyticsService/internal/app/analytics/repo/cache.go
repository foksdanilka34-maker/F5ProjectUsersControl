package repo

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/foksdanilka34-maker/F5ProjectUsersControl/AnalyticsService/internal/app/analytics"
	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(client *redis.Client) *RedisCache {
	return &RedisCache{client: client}
}

func (r *RedisCache) GetEmployeeMetricsCache(ctx context.Context, employeeID string, date time.Time) (*analytics.EmployeeMetrics, error) {
	key := fmt.Sprintf("emp_metrics:%s:%s", employeeID, date.Format("2006-01-02"))

	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("redis get error: %w", err)
	}

	var metrics analytics.EmployeeMetrics
	if err := json.Unmarshal([]byte(val), &metrics); err != nil {
		return nil, fmt.Errorf("unmarshal error: %w", err)
	}

	return &metrics, nil
}

func (r *RedisCache) SetEmployeeMetricsCache(ctx context.Context, metrics *analytics.EmployeeMetrics, ttl time.Duration) error {
	key := fmt.Sprintf("emp_metrics:%s:%s", metrics.EmployeeID, metrics.MetricDate.Format("2006-01-02"))

	data, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}

	if err := r.client.Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("redis set error: %w", err)
	}

	return nil
}

func (r *RedisCache) GetProjectMetricsCache(ctx context.Context, projectID string, date time.Time) (*analytics.ProjectMetrics, error) {
	key := fmt.Sprintf("proj_metrics:%s:%s", projectID, date.Format("2006-01-02"))

	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("redis get error: %w", err)
	}

	var metrics analytics.ProjectMetrics
	if err := json.Unmarshal([]byte(val), &metrics); err != nil {
		return nil, fmt.Errorf("unmarshal error: %w", err)
	}

	return &metrics, nil
}

func (r *RedisCache) SetProjectMetricsCache(ctx context.Context, metrics *analytics.ProjectMetrics, ttl time.Duration) error {
	key := fmt.Sprintf("proj_metrics:%s:%s", metrics.ProjectID, metrics.MetricDate.Format("2006-01-02"))

	data, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}

	if err := r.client.Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("redis set error: %w", err)
	}

	return nil
}

func (r *RedisCache) GetDepartmentMetricsCache(ctx context.Context, departmentID string, date time.Time) (*analytics.DepartmentMetrics, error) {
	key := fmt.Sprintf("dept_metrics:%s:%s", departmentID, date.Format("2006-01-02"))

	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("redis get error: %w", err)
	}

	var metrics analytics.DepartmentMetrics
	if err := json.Unmarshal([]byte(val), &metrics); err != nil {
		return nil, fmt.Errorf("unmarshal error: %w", err)
	}

	return &metrics, nil
}

func (r *RedisCache) SetDepartmentMetricsCache(ctx context.Context, metrics *analytics.DepartmentMetrics, ttl time.Duration) error {
	key := fmt.Sprintf("dept_metrics:%s:%s", metrics.DepartmentID, metrics.MetricDate.Format("2006-01-02"))

	data, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}

	if err := r.client.Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("redis set error: %w", err)
	}

	return nil
}

func (r *RedisCache) GetDailySnapshotCache(ctx context.Context, date time.Time) (*analytics.DailySnapshot, error) {
	key := fmt.Sprintf("daily_snapshot:%s", date.Format("2006-01-02"))

	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("redis get error: %w", err)
	}

	var snapshot analytics.DailySnapshot
	if err := json.Unmarshal([]byte(val), &snapshot); err != nil {
		return nil, fmt.Errorf("unmarshal error: %w", err)
	}

	return &snapshot, nil
}

func (r *RedisCache) SetDailySnapshotCache(ctx context.Context, snapshot *analytics.DailySnapshot, ttl time.Duration) error {
	key := fmt.Sprintf("daily_snapshot:%s", snapshot.SnapshotDate.Format("2006-01-02"))

	data, err := json.Marshal(snapshot)
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}

	if err := r.client.Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("redis set error: %w", err)
	}

	return nil
}

func (r *RedisCache) InvalidateEmployeeMetricsCache(ctx context.Context, employeeID string) error {
	pattern := fmt.Sprintf("emp_metrics:%s:*", employeeID)
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to get keys: %w", err)
	}

	if len(keys) > 0 {
		if err := r.client.Del(ctx, keys...).Err(); err != nil {
			return fmt.Errorf("failed to delete keys: %w", err)
		}
	}

	return nil
}

func (r *RedisCache) InvalidateProjectMetricsCache(ctx context.Context, projectID string) error {
	pattern := fmt.Sprintf("proj_metrics:%s:*", projectID)
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to get keys: %w", err)
	}

	if len(keys) > 0 {
		if err := r.client.Del(ctx, keys...).Err(); err != nil {
			return fmt.Errorf("failed to delete keys: %w", err)
		}
	}

	return nil
}

func (r *RedisCache) InvalidateDepartmentMetricsCache(ctx context.Context, departmentID string) error {
	pattern := fmt.Sprintf("dept_metrics:%s:*", departmentID)
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to get keys: %w", err)
	}

	if len(keys) > 0 {
		if err := r.client.Del(ctx, keys...).Err(); err != nil {
			return fmt.Errorf("failed to delete keys: %w", err)
		}
	}

	return nil
}
