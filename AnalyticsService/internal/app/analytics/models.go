package analytics

import (
	"time"
)

type Period int32

const (
	PERIOD_UNSPECIFIED Period = iota
	PERIOD_DAILY
	PERIOD_WEEKLY
	PERIOD_MONTHLY
)

func (p Period) String() string {
	switch p {
	case PERIOD_DAILY:
		return "daily"
	case PERIOD_WEEKLY:
		return "weekly"
	case PERIOD_MONTHLY:
		return "monthly"
	default:
		return "unspecified"
	}
}

type HealthStatus int32

const (
	HEALTH_STATUS_UNSPECIFIED HealthStatus = iota
	HEALTH_STATUS_HEALTHY
	HEALTH_STATUS_AT_RISK
	HEALTH_STATUS_CRITICAL
)

func (hs HealthStatus) String() string {
	switch hs {
	case HEALTH_STATUS_HEALTHY:
		return "healthy"
	case HEALTH_STATUS_AT_RISK:
		return "at_risk"
	case HEALTH_STATUS_CRITICAL:
		return "critical"
	default:
		return "unspecified"
	}
}

type TaskPriority string

type EmployeeMetrics struct {
	ID                       string
	EmployeeID               string
	MetricDate               time.Time
	AssignedTasks            int32
	CompletedTasks           int32
	InProgressTasks          int32
	OverdueTasks             int32
	OnTimeCompletionTask     int32
	TotalTaskDurationSeconds int64
	CreatedAt                time.Time
	UpdatedAt                time.Time

	TaskCompletionRate   float64
	OnTimeCompletionRate float64
	EfficiencyScore      float64
}

type ProjectMetrics struct {
	ID                                string
	ProjectID                         string
	ManagerID                         string
	MetricDate                        time.Time
	TotalTasks                        int32
	CompletedTasks                    int32
	InProgressTasks                   int32
	OverdueTasks                      int32
	OnTimeCompletedTasks              int32
	TeamSize                          int32
	TotalTaskDurationSecondsCompleted int64
	TotalPriorityWeightCompleted      int32
	CreatedAt                         time.Time
	UpdatedAt                         time.Time

	DeliveryPerformance     float64
	SchedulePerformance     float64
	QualityPerformance      float64
	TeamPerformance         float64
	HealthIndex             float64
	RiskScore               float64
	HealthStatus            HealthStatus
	Velocity                float64
	ProjectedEndDate        *time.Time
	TeamCapacityUtilization float64
	AvgTeamEfficiency       float64
	IsAtRisk                bool
	DaysUntilDue            *int32
}

type ListEmployeeMetrics struct {
	PageSize     int32
	PageNumber   int32
	DepartmentID string

	StartDate *time.Time
	EndDate   *time.Time
}

type ListProjectMetrics struct {
	PageSize   *int32
	PageNumber *int32
	ManagerID  string

	StartDate *time.Time
	EndDate   *time.Time
}

type ProductivityTrends struct {
	Period Period
	Limit  int32

	DepartmentID string
	EmployeeID   *string
}

type ComletionRateTrends struct {
	Period Period
	Limit  int32

	ProjectID  string
	EmployeeID string
}

type DashboardStats struct {
	TotalEmployees       int32
	ActiveEmployees      int32
	TotalProjects        int32
	ActiveProjects       int32
	TotalTasks           int32
	CompletedTasks       int32
	OverDueTasks         int32
	AvgCompanyEfficiency float32
	AvgOnTimeRate        float32
	TopEmployees         []TopEmployee
	ProblematicProjects  []BottomProject
	CalculatedAt         time.Time
}

type TopEmployee struct {
	ID              string
	EfficiencyScore float32
	TaskCompleted   int32
}

type BottomProject struct {
	ProjectID   string
	HealthScore float32
	OnTimeRate  float32
}

type ProductivityTrend struct {
	Date                 time.Time
	AvgEfficiency        float32
	TotalTasksCompleted  int32
	TotalEmployeesActive int32
}

type CompletionRateTrend struct {
	Date           time.Time
	OnTimeRate     float32
	OverallRate    float32
	CompletedCount int32
	OverDueCount   int32
}

type ProductivityTrendsResp struct {
	Prod   []ProductivityTrend
	Period Period
}

type CompletionRateTrendResp struct {
	CompTrend []CompletionRateTrend
	Period    Period
}
