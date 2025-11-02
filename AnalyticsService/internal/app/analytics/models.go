package analytics

import (
	"time"
)

type EmployeeMetrics struct {
	ID                       string
	EmployeeID               string
	EmployeeName             string
	DepartmentID             string
	PositionID               string
	MetricDate               time.Time
	TasksCompleted           int32
	TasksAssigned            int32
	AvgCompletionTimeHours   float64
	OnTimeCompletionRate     float64 
	AvgTaskPriority          float64
	SkillsUsed               []string
	ProjectsInvolved         int32
	EfficiencyScore          float64 
	CreatedAt                time.Time
	UpdatedAt                time.Time
}

type ProjectMetrics struct {
	ID                     string
	ProjectID              string
	ProjectName            string
	ManagerID              string
	ManagerName            string
	MetricDate             time.Time
	TotalTasks             int32
	CompletedTasks         int32
	InProgressTasks        int32
	OverdueTasks           int32
	CompletionRate         float64 
	OnTimeCompletionRate   float64 
	TeamSize               int32
	AvgTaskDurationHours   float64
	ProjectHealthScore     float64 
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

type DepartmentMetrics struct {
	ID                        string
	DepartmentID              string
	DepartmentName            string
	MetricDate                time.Time
	TotalEmployees            int32
	ActiveProjects            int32
	TotalTasks                int32
	CompletedTasks            int32
	AvgEmployeeEfficiency     float64
	DepartmentCompletionRate  float64 
	DepartmentOnTimeRate      float64 
	DepartmentHealthScore     float64 
	CreatedAt                 time.Time
	UpdatedAt                 time.Time
}

type DailySnapshot struct {
	ID                 string
	SnapshotDate       time.Time
	TotalEmployees     int32
	ActiveEmployees    int32
	TotalProjects      int32
	ActiveProjects     int32
	TotalTasks         int32
	CompletedTasks     int32
	OverdueTasks       int32
	AvgCompanyEfficiency float64
	AvgOnTimeRate      float64
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type CachedMetrics struct {
	ID        string
	CacheKey  string
	CacheValue []byte
	TTLSeconds int32
	CreatedAt time.Time
	ExpiresAt time.Time
}
