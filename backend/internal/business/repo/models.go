package repo

import "time"

// Project - проект
type Project struct {
	ID          int64
	Name        string
	Description string
	Status      string // active, completed, archived
	StartDate   *time.Time
	EndDate     *time.Time
	OwnerID     int64
	CreatedAt   time.Time
	UpdatedAt   time.Time

	// Joined fields
	OwnerName   string
	TaskCount   int
	MemberCount int
	TaskStats   *TaskStats
}

// TaskStats - статистика задач проекта
type TaskStats struct {
	Total      int
	Todo       int
	InProgress int
	Done       int
}

// ProjectMember - участник проекта
type ProjectMember struct {
	ProjectID int64
	UserID    int64
	Role      string // owner, manager, member
	JoinedAt  time.Time

	// Joined fields
	UserName string
}

// Task - задача
type Task struct {
	ID          int64
	ProjectID   int64
	Title       string
	Description string
	Status      string // todo, in_progress, review, done
	Priority    string // low, medium, high, critical
	AssigneeID  *int64
	CreatorID   int64
	OrderIndex  int32
	DueDate     *time.Time
	StartedAt   *time.Time
	CompletedAt *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time

	// Joined fields
	AssigneeName string
	CreatorName  string
	ProjectName  string
}

// TaskComment - комментарий к задаче
type TaskComment struct {
	ID        int64
	TaskID    int64
	UserID    int64
	Content   string
	CreatedAt time.Time

	// Joined fields
	UserName string
}

// TaskHistory - история изменений задачи
type TaskHistory struct {
	ID        int64
	TaskID    int64
	UserID    int64
	Field     string
	OldValue  string
	NewValue  string
	ChangedAt time.Time

	// Joined fields
	UserName string
}

// AnalyticsSummary - сводная аналитика
type AnalyticsSummary struct {
	TotalProjects     int32
	ActiveProjects    int32
	TotalTasks        int32
	CompletedTasks    int32
	InProgressTasks   int32
	OverdueTasks      int32
	TotalEmployees    int32
	ActiveEmployees   int32
	AvgTasksPerMember float32
}

// EmployeeMetrics - метрики сотрудника
type EmployeeMetrics struct {
	EmployeeID      int64
	AssignedTasks   int32
	CompletedTasks  int32
	InProgressTasks int32
	OverdueTasks    int32
	CompletionRate  float64
	OnTimeRate      float64
}

// ProjectMetrics - метрики проекта
type ProjectMetrics struct {
	ProjectID       int64
	ManagerID       int64
	TotalTasks      int32
	CompletedTasks  int32
	InProgressTasks int32
	OverdueTasks    int32
	TeamSize        int32
	ProgressPercent float64
	OnTimeRate      float64
	HealthStatus    string
}
