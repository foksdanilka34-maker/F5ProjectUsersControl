package dto

import "time"

type DashboardStatsDTO struct {
	TotalEmployees      int                      `json:"total_employees"`
	ActiveEmployees     int                      `json:"active_employees"`
	TotalProjects       int                      `json:"total_projects"`
	ActiveProjects      int                      `json:"active_projects"`
	TotalTasks          int                      `json:"total_tasks"`
	CompletedTasks      int                      `json:"completed_tasks"`
	OverdueTasks        int                      `json:"overdue_tasks"`
	CompletedOnTime     int                      `json:"completed_on_time"`
	CompletedLate       int                      `json:"completed_late"`
	AvgCompletionRate   float64                  `json:"avg_completion_rate"`
	AvgOnTimeRate       float64                  `json:"avg_on_time_rate"`
	TopEmployees        []TopEmployeeDTO         `json:"top_employees"`
	ProblematicProjects []ProblematicProjectDTO  `json:"problematic_projects"`
	CalculatedAt        time.Time                `json:"calculated_at"`
}

type TopEmployeeDTO struct {
	EmployeeID     int64   `json:"employee_id"`
	CompletionRate float64 `json:"completion_rate"`
	TasksCompleted int     `json:"tasks_completed"`
}

type ProblematicProjectDTO struct {
	ProjectID    int64   `json:"project_id"`
	OnTimeRate   float64 `json:"on_time_rate"`
	HealthStatus string  `json:"health_status"`
}

type ProjectMetricsDTO struct {
	ProjectID       int64     `json:"project_id"`
	ManagerID       int64     `json:"manager_id"`
	TotalTasks      int       `json:"total_tasks"`
	CompletedTasks  int       `json:"completed_tasks"`
	CompletedOnTime int       `json:"completed_on_time"`
	CompletedLate   int       `json:"completed_late"`
	InProgressTasks int       `json:"in_progress_tasks"`
	OverdueTasks    int       `json:"overdue_tasks"`
	TeamSize        int       `json:"team_size"`
	ProgressPercent float64   `json:"progress_percent"`
	OnTimeRate      float64   `json:"on_time_rate"`
	HealthStatus    string    `json:"health_status"`
	CalculatedAt    time.Time `json:"calculated_at"`
}

type EmployeeMetricsDTO struct {
	EmployeeID      int64   `json:"employee_id"`
	AssignedTasks   int     `json:"assigned_tasks"`
	CompletedTasks  int     `json:"completed_tasks"`
	CompletedOnTime int     `json:"completed_on_time"`
	CompletedLate   int     `json:"completed_late"`
	InProgressTasks int     `json:"in_progress_tasks"`
	OverdueTasks    int     `json:"overdue_tasks"`
	CompletionRate  float64 `json:"completion_rate"`
	OnTimeRate      float64 `json:"on_time_rate"`
}

type ProductivityTrendEntryDTO struct {
	Date              string  `json:"date"`
	TasksCompleted    int     `json:"tasks_completed"`
	AvgCompletionRate float64 `json:"avg_completion_rate"`
}

type ProductivityTrendsDTO struct {
	Entries []ProductivityTrendEntryDTO `json:"entries"`
	Period  string                      `json:"period"`
}
