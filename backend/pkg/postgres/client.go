package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	Schema   string // Схема для search_path, если не указана - используется имя базы
}

func Connect(ctx context.Context, cfg *Config) (*pgxpool.Pool, error) {
	schema := cfg.Schema
	if schema == "" {
		schema = cfg.Database
	}
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable&search_path=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database, schema)

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	poolConfig.MinConns = 2                       // Keep warm connections
	poolConfig.MaxConns = 10                      // Max connections
	poolConfig.MaxConnLifetime = 30 * time.Minute // Connection lifetime
	poolConfig.MaxConnIdleTime = 5 * time.Minute  // Idle timeout
	poolConfig.HealthCheckPeriod = 30 * time.Second

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return pool, nil
}


