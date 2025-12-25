package core

import (
	"context"
	"time"

	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/logs/repo"
	"github.com/google/uuid"
)

// LogRepository - интерфейс репозитория логов
type LogRepository interface {
	Create(ctx context.Context, entry *repo.LogEntry) error
	List(ctx context.Context, filter *repo.LogFilter) ([]*repo.LogEntry, int, error)
	DeleteOlderThan(ctx context.Context, days int) (int64, error)
}

// LogService - сервис логирования
type LogService struct {
	repo LogRepository
}

func NewLogService(repo LogRepository) *LogService {
	return &LogService{repo: repo}
}

type CreateLogRequest struct {
	Service   string
	Level     string
	Message   string
	UserID    *string
	RequestID *string
	Metadata  map[string]string
}

func (s *LogService) Log(ctx context.Context, req *CreateLogRequest) error {
	entry := &repo.LogEntry{
		ID:        uuid.New().String(),
		Service:   req.Service,
		Level:     req.Level,
		Message:   req.Message,
		UserID:    req.UserID,
		RequestID: req.RequestID,
		Metadata:  req.Metadata,
		Timestamp: time.Now(),
	}

	return s.repo.Create(ctx, entry)
}

func (s *LogService) Info(ctx context.Context, service, message string, metadata map[string]string) error {
	return s.Log(ctx, &CreateLogRequest{
		Service:  service,
		Level:    "info",
		Message:  message,
		Metadata: metadata,
	})
}

func (s *LogService) Warn(ctx context.Context, service, message string, metadata map[string]string) error {
	return s.Log(ctx, &CreateLogRequest{
		Service:  service,
		Level:    "warn",
		Message:  message,
		Metadata: metadata,
	})
}

func (s *LogService) Error(ctx context.Context, service, message string, metadata map[string]string) error {
	return s.Log(ctx, &CreateLogRequest{
		Service:  service,
		Level:    "error",
		Message:  message,
		Metadata: metadata,
	})
}

func (s *LogService) Debug(ctx context.Context, service, message string, metadata map[string]string) error {
	return s.Log(ctx, &CreateLogRequest{
		Service:  service,
		Level:    "debug",
		Message:  message,
		Metadata: metadata,
	})
}

type ListLogsFilter struct {
	Service    string
	Level      string
	UserID     string
	RequestID  string
	StartTime  *time.Time
	EndTime    *time.Time
	PageSize   int
	PageNumber int
}

func (s *LogService) ListLogs(ctx context.Context, filter *ListLogsFilter) ([]*repo.LogEntry, int, error) {
	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = 50
	}
	pageNumber := filter.PageNumber
	if pageNumber <= 0 {
		pageNumber = 1
	}
	offset := (pageNumber - 1) * pageSize

	repoFilter := &repo.LogFilter{
		Service:   filter.Service,
		Level:     filter.Level,
		UserID:    filter.UserID,
		RequestID: filter.RequestID,
		StartTime: filter.StartTime,
		EndTime:   filter.EndTime,
		PageSize:  pageSize,
		Offset:    offset,
	}

	return s.repo.List(ctx, repoFilter)
}

func (s *LogService) CleanupOldLogs(ctx context.Context, retentionDays int) (int64, error) {
	if retentionDays <= 0 {
		retentionDays = 30 // default 30 days
	}
	return s.repo.DeleteOlderThan(ctx, retentionDays)
}

// HandleLogEntry - обработчик для NATS сообщений логов
type NATSLogEntry struct {
	Service   string `json:"service"`
	Level     string `json:"level"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
	Data      any    `json:"data,omitempty"`
}

func (s *LogService) HandleLogEntry(ctx context.Context, entry *NATSLogEntry) error {
	return s.Log(ctx, &CreateLogRequest{
		Service: entry.Service,
		Level:   entry.Level,
		Message: entry.Message,
	})
}
