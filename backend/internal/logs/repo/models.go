package repo

import "time"

type LogEntry struct {
	ID        string
	Service   string
	Level     string // debug, info, warn, error
	Message   string
	UserID    *string
	RequestID *string
	Metadata  map[string]string
	Timestamp time.Time
}

type LogFilter struct {
	Service   string
	Level     string
	UserID    string
	RequestID string
	StartTime *time.Time
	EndTime   *time.Time
	PageSize  int
	Offset    int
}


