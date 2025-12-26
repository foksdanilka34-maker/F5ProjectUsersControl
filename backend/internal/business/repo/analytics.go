package repo

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Analytics models
type ProjectAnalytics struct {
	ProjectID         int64
	ProjectName       string
	TotalTasks        int32
	CompletedTasks    int32
	CompletedOnTime   int32 // Завершено вовремя (completed_at <= due_date)
	CompletedLate     int32 // Завершено с опозданием (completed_at > due_date)
	InProgressTasks   int32
	OverdueTasks      int32 // Текущие просроченные (не завершены)
	AvgCompletionTime float64
	MemberCount       int32
}

type EmployeeAnalytics struct {
	UserID              int64
	AssignedTasks       int32
	CompletedTasks      int32
	CompletedOnTime     int32
	CompletedLate       int32
	InProgressTasks     int32
	OverdueTasks        int32
	AvgCompletionTime   float64
	ProjectCount        int32
	// Взвешенные метрики (с учётом приоритета)
	WeightedOnTime      float64 // Сумма весов задач выполненных вовремя
	WeightedTotal       float64 // Сумма весов всех завершённых задач
}

type TimeSeriesPoint struct {
	Date  time.Time
	Count int
}

type TaskDistribution struct {
	Status string
	Count  int
}

type PriorityDistribution struct {
	Priority string
	Count    int
}

type AnalyticsRepo struct {
	db *pgxpool.Pool
}

func NewAnalyticsRepo(db *pgxpool.Pool) *AnalyticsRepo {
	return &AnalyticsRepo{db: db}
}

func (r *AnalyticsRepo) GetSummary(ctx context.Context) (*AnalyticsSummary, error) {
	query := `
		SELECT 
			(SELECT COUNT(*) FROM business.projects) as total_projects,
			(SELECT COUNT(*) FROM business.projects WHERE status = 'ACTIVE') as active_projects,
			(SELECT COUNT(*) FROM business.tasks) as total_tasks,
			(SELECT COUNT(*) FROM business.tasks WHERE status = 'DONE') as completed_tasks,
			-- Завершено вовремя: completed_at <= due_date (или без дедлайна)
			(SELECT COUNT(*) FROM business.tasks WHERE status = 'DONE' AND (due_date IS NULL OR completed_at <= due_date)) as completed_on_time,
			-- Завершено с опозданием: completed_at > due_date
			(SELECT COUNT(*) FROM business.tasks WHERE status = 'DONE' AND due_date IS NOT NULL AND completed_at > due_date) as completed_late,
			-- Текущие просроченные: не завершены и дедлайн прошёл
			(SELECT COUNT(*) FROM business.tasks WHERE due_date IS NOT NULL AND due_date < NOW() AND status != 'DONE') as overdue_tasks,
			(SELECT COUNT(DISTINCT user_id) FROM business.project_members) as total_employees,
			(SELECT COUNT(DISTINCT user_id) FROM business.project_members) as active_employees
	`
	var s AnalyticsSummary
	err := r.db.QueryRow(ctx, query).Scan(
		&s.TotalProjects, &s.ActiveProjects, &s.TotalTasks, &s.CompletedTasks,
		&s.CompletedOnTime, &s.CompletedLate, &s.OverdueTasks,
		&s.TotalEmployees, &s.ActiveEmployees,
	)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *AnalyticsRepo) GetProjectAnalytics(ctx context.Context, projectID int64) (*ProjectAnalytics, error) {
	query := `
		SELECT 
			p.id,
			p.name,
			COUNT(t.id) as total_tasks,
			COUNT(t.id) FILTER (WHERE t.status = 'DONE') as completed_tasks,
			-- Завершено вовремя: completed_at <= due_date (или без дедлайна)
			COUNT(t.id) FILTER (WHERE t.status = 'DONE' AND (t.due_date IS NULL OR t.completed_at <= t.due_date)) as completed_on_time,
			-- Завершено с опозданием: completed_at > due_date
			COUNT(t.id) FILTER (WHERE t.status = 'DONE' AND t.due_date IS NOT NULL AND t.completed_at > t.due_date) as completed_late,
			COUNT(t.id) FILTER (WHERE t.status = 'IN_PROGRESS') as in_progress_tasks,
			-- Текущие просроченные: не завершены и дедлайн прошёл
			COUNT(t.id) FILTER (WHERE t.due_date IS NOT NULL AND t.due_date < NOW() AND t.status != 'DONE') as overdue_tasks,
			-- Среднее время выполнения: completed_at - created_at (в днях)
			COALESCE(AVG(EXTRACT(EPOCH FROM (t.completed_at - t.created_at)) / 86400) FILTER (WHERE t.status = 'DONE' AND t.completed_at IS NOT NULL), 0) as avg_completion_time,
			(SELECT COUNT(*) FROM business.project_members WHERE project_id = p.id) as member_count
		FROM business.projects p
		LEFT JOIN business.tasks t ON p.id = t.project_id
		WHERE p.id = $1
		GROUP BY p.id, p.name
	`
	var a ProjectAnalytics
	err := r.db.QueryRow(ctx, query, projectID).Scan(
		&a.ProjectID, &a.ProjectName, &a.TotalTasks, &a.CompletedTasks,
		&a.CompletedOnTime, &a.CompletedLate, &a.InProgressTasks, &a.OverdueTasks,
		&a.AvgCompletionTime, &a.MemberCount,
	)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *AnalyticsRepo) GetEmployeeAnalytics(ctx context.Context, userID int64) (*EmployeeAnalytics, error) {
	// Веса приоритетов: critical=4, high=3, medium=2, low=1
	query := `
		WITH priority_weights AS (
			SELECT 
				t.id,
				t.status,
				t.due_date,
				t.completed_at,
				t.created_at,
				CASE t.priority 
					WHEN 4 THEN 4  -- critical
					WHEN 3 THEN 3  -- high
					WHEN 2 THEN 2  -- medium
					WHEN 1 THEN 1  -- low
					ELSE 2         -- default medium
				END as weight
			FROM business.tasks t
			WHERE t.assignee_id = $1
		)
		SELECT 
			COUNT(id) as assigned_tasks,
			COUNT(id) FILTER (WHERE status = 'DONE') as completed_tasks,
			-- Завершено вовремя: completed_at <= due_date (или без дедлайна)
			COUNT(id) FILTER (WHERE status = 'DONE' AND (due_date IS NULL OR completed_at <= due_date)) as completed_on_time,
			-- Завершено с опозданием: completed_at > due_date
			COUNT(id) FILTER (WHERE status = 'DONE' AND due_date IS NOT NULL AND completed_at > due_date) as completed_late,
			COUNT(id) FILTER (WHERE status = 'IN_PROGRESS') as in_progress_tasks,
			-- Текущие просроченные: не завершены и дедлайн прошёл
			COUNT(id) FILTER (WHERE due_date IS NOT NULL AND due_date < NOW() AND status != 'DONE') as overdue_tasks,
			COALESCE(AVG(EXTRACT(EPOCH FROM (completed_at - created_at)) / 86400) FILTER (WHERE status = 'DONE' AND completed_at IS NOT NULL), 0) as avg_completion_time,
			-- Взвешенная сумма задач выполненных вовремя
			COALESCE(SUM(weight) FILTER (WHERE status = 'DONE' AND (due_date IS NULL OR completed_at <= due_date)), 0) as weighted_on_time,
			-- Взвешенная сумма всех завершённых задач
			COALESCE(SUM(weight) FILTER (WHERE status = 'DONE'), 0) as weighted_total
		FROM priority_weights
	`
	var a EmployeeAnalytics
	a.UserID = userID
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&a.AssignedTasks, &a.CompletedTasks, &a.CompletedOnTime, &a.CompletedLate,
		&a.InProgressTasks, &a.OverdueTasks, &a.AvgCompletionTime,
		&a.WeightedOnTime, &a.WeightedTotal,
	)
	if err != nil {
		return nil, err
	}
	
	// Get project count separately
	var projectCount int32
	err = r.db.QueryRow(ctx, `SELECT COUNT(DISTINCT project_id) FROM business.project_members WHERE user_id = $1`, userID).Scan(&projectCount)
	if err == nil {
		a.ProjectCount = projectCount
	}
	
	return &a, nil
}

func (r *AnalyticsRepo) GetTasksTimeSeries(ctx context.Context, startDate, endDate time.Time, projectID int64) ([]TimeSeriesPoint, error) {
	query := `
		SELECT DATE(created_at) as date, COUNT(*) as count
		FROM business.tasks
		WHERE created_at BETWEEN $1 AND $2
	`
	args := []interface{}{startDate, endDate}
	if projectID > 0 {
		query += " AND project_id = $3"
		args = append(args, projectID)
	}
	query += " GROUP BY DATE(created_at) ORDER BY date"

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var points []TimeSeriesPoint
	for rows.Next() {
		var p TimeSeriesPoint
		if err := rows.Scan(&p.Date, &p.Count); err != nil {
			return nil, err
		}
		points = append(points, p)
	}
	return points, nil
}

func (r *AnalyticsRepo) GetCompletedTasksTimeSeries(ctx context.Context, startDate, endDate time.Time, projectID int64) ([]TimeSeriesPoint, error) {
	query := `
		SELECT DATE(updated_at) as date, COUNT(*) as count
		FROM business.tasks
		WHERE status = 'DONE' AND updated_at BETWEEN $1 AND $2
	`
	args := []interface{}{startDate, endDate}
	if projectID > 0 {
		query += " AND project_id = $3"
		args = append(args, projectID)
	}
	query += " GROUP BY DATE(updated_at) ORDER BY date"

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var points []TimeSeriesPoint
	for rows.Next() {
		var p TimeSeriesPoint
		if err := rows.Scan(&p.Date, &p.Count); err != nil {
			return nil, err
		}
		points = append(points, p)
	}
	return points, nil
}

func (r *AnalyticsRepo) GetTaskDistribution(ctx context.Context, projectID int64) ([]TaskDistribution, error) {
	query := `SELECT status, COUNT(*) FROM business.tasks`
	args := []interface{}{}
	if projectID > 0 {
		query += " WHERE project_id = $1"
		args = append(args, projectID)
	}
	query += " GROUP BY status"

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dist []TaskDistribution
	for rows.Next() {
		var d TaskDistribution
		if err := rows.Scan(&d.Status, &d.Count); err != nil {
			return nil, err
		}
		dist = append(dist, d)
	}
	return dist, nil
}

func (r *AnalyticsRepo) GetPriorityDistribution(ctx context.Context, projectID int64) ([]PriorityDistribution, error) {
	query := `SELECT priority, COUNT(*) FROM business.tasks`
	args := []interface{}{}
	if projectID > 0 {
		query += " WHERE project_id = $1"
		args = append(args, projectID)
	}
	query += " GROUP BY priority"

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dist []PriorityDistribution
	for rows.Next() {
		var d PriorityDistribution
		if err := rows.Scan(&d.Priority, &d.Count); err != nil {
			return nil, err
		}
		dist = append(dist, d)
	}
	return dist, nil
}
