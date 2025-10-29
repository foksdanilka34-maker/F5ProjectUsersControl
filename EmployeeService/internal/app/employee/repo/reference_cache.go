package employee

import (
	"context"
	"errors"
	"time"

	emp "github.com/foksdanilka34-maker/F5ProjectUsersControl/EmployeeService/internal/app/employee"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/EmployeeService/internal/app"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const (
	DepartmentHashKey = "departments:hash"
	PositionHashKey   = "positions:hash"
	SkillHashKey      = "skills:hash"

	ReferenceListTTL = 1 * time.Hour
	ReferenceItemTTL = 30 * time.Minute
)

type ReferenceCache struct {
	redis *redis.Client
}

func NewReferenceCache(r *redis.Client) *ReferenceCache {
	return &ReferenceCache{
		redis: r,
	}
}

func (c *ReferenceCache) SetDepartmentsList(ctx context.Context, departments []*emp.Department) error {
	if len(departments) == 0 {
		return nil
	}

	pipe := c.redis.Pipeline()
	pipe.Del(ctx, DepartmentHashKey)

	for _, dept := range departments {
		pipe.HSet(ctx, DepartmentHashKey, dept.ID, dept.Name)
	}

	pipe.Expire(ctx, DepartmentHashKey, ReferenceListTTL)

	_, err := pipe.Exec(ctx)
	if err != nil {
		app.Logger.Error("error setting departments hash", zap.Error(err))
		return err
	}
	return nil
}

func (c *ReferenceCache) GetDepartmentsList(ctx context.Context) ([]*emp.Department, error) {
	data, err := c.redis.HGetAll(ctx, DepartmentHashKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}

	if len(data) == 0 {
		return nil, nil
	}

	departments := make([]*emp.Department, 0, len(data))
	for id, name := range data {
		departments = append(departments, &emp.Department{
			ID:   id,
			Name: name,
		})
	}

	return departments, nil
}

func (c *ReferenceCache) SetDepartment(ctx context.Context, department *emp.Department) error {
	return c.redis.HSet(ctx, DepartmentHashKey, department.ID, department.Name).Err()
}

func (c *ReferenceCache) GetDepartment(ctx context.Context, id string) (*emp.Department, error) {
	name, err := c.redis.HGet(ctx, DepartmentHashKey, id).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}

	return &emp.Department{
		ID:   id,
		Name: name,
	}, nil
}

func (c *ReferenceCache) InvalidateDepartments(ctx context.Context) error {
	return c.redis.Del(ctx, DepartmentHashKey).Err()
}

func (c *ReferenceCache) InvalidateDepartment(ctx context.Context, id string) error {
	return c.redis.HDel(ctx, DepartmentHashKey, id).Err()
}

// ============= Position Cache (используем Hash) =============

func (c *ReferenceCache) SetPositionsList(ctx context.Context, positions []*emp.Position) error {
	if len(positions) == 0 {
		return nil
	}

	pipe := c.redis.Pipeline()
	pipe.Del(ctx, PositionHashKey)

	for _, pos := range positions {
		pipe.HSet(ctx, PositionHashKey, pos.ID, pos.Name)
	}

	pipe.Expire(ctx, PositionHashKey, ReferenceListTTL)

	_, err := pipe.Exec(ctx)
	if err != nil {
		app.Logger.Error("error setting positions hash", zap.Error(err))
		return err
	}
	return nil
}

func (c *ReferenceCache) GetPositionsList(ctx context.Context) ([]*emp.Position, error) {
	data, err := c.redis.HGetAll(ctx, PositionHashKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}

	if len(data) == 0 {
		return nil, nil
	}

	positions := make([]*emp.Position, 0, len(data))
	for id, name := range data {
		positions = append(positions, &emp.Position{
			ID:   id,
			Name: name,
		})
	}

	return positions, nil
}

func (c *ReferenceCache) SetPosition(ctx context.Context, position *emp.Position) error {
	return c.redis.HSet(ctx, PositionHashKey, position.ID, position.Name).Err()
}

func (c *ReferenceCache) GetPosition(ctx context.Context, id string) (*emp.Position, error) {
	name, err := c.redis.HGet(ctx, PositionHashKey, id).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}

	return &emp.Position{
		ID:   id,
		Name: name,
	}, nil
}

func (c *ReferenceCache) InvalidatePositions(ctx context.Context) error {
	return c.redis.Del(ctx, PositionHashKey).Err()
}

func (c *ReferenceCache) InvalidatePosition(ctx context.Context, id string) error {
	return c.redis.HDel(ctx, PositionHashKey, id).Err()
}

func (c *ReferenceCache) SetSkillsList(ctx context.Context, skills []*emp.Skill) error {
	if len(skills) == 0 {
		return nil
	}

	pipe := c.redis.Pipeline()
	pipe.Del(ctx, SkillHashKey)

	for _, skill := range skills {
		pipe.HSet(ctx, SkillHashKey, skill.ID, skill.Name)
	}

	pipe.Expire(ctx, SkillHashKey, ReferenceListTTL)

	_, err := pipe.Exec(ctx)
	if err != nil {
		app.Logger.Error("error setting skills hash", zap.Error(err))
		return err
	}
	return nil
}

func (c *ReferenceCache) GetSkillsList(ctx context.Context) ([]*emp.Skill, error) {
	data, err := c.redis.HGetAll(ctx, SkillHashKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}

	if len(data) == 0 {
		return nil, nil
	}

	skills := make([]*emp.Skill, 0, len(data))
	for id, name := range data {
		skills = append(skills, &emp.Skill{
			ID:   id,
			Name: name,
		})
	}

	return skills, nil
}

func (c *ReferenceCache) InvalidateSkills(ctx context.Context) error {
	return c.redis.Del(ctx, SkillHashKey).Err()
}
