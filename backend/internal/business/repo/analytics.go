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
	InProgressTasks   int32
	OverdueTasks      int32
	AvgCompletionTime float64
	MemberCount       int32
}

type EmployeeAnalytics struct {
	UserID            int64
	AssignedTasks     int32
	CompletedTasks    int32
	InProgressTasks   int32
	OverdueTasks      int32
	AvgCompletionTime float64
	ProjectCount      int32
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
			(SELECT COUNT(*) FROM projects) as total_projects,
			(SELECT COUNT(*) FROM projects WHERE status = 'active') as active_projects,
			(SELECT COUNT(*) FROM tasks) as total_tasks,
			(SELECT COUNT(*) FROM tasks WHERE status = 'done') as completed_tasks,
			(SELECT COUNT(*) FROM tasks WHERE due_date < NOW() AND status != 'done') as overdue_tasks
	`
	var s AnalyticsSummary
	err := r.db.QueryRow(ctx, query).Scan(
		&s.TotalProjects, &s.ActiveProjects, &s.TotalTasks, &s.CompletedTasks, &s.OverdueTasks,
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
			COUNT(t.id) FILTER (WHERE t.status = 'done') as completed_tasks,
			COUNT(t.id) FILTER (WHERE t.status = 'in_progress') as in_progress_tasks,
			COUNT(t.id) FILTER (WHERE t.due_date < NOW() AND t.status != 'done') as overdue_tasks,
			COALESCE(AVG(EXTRACT(EPOCH FROM (t.updated_at - t.created_at)) / 86400) FILTER (WHERE t.status = 'done'), 0) as avg_completion_time,
			(SELECT COUNT(*) FROM project_members WHERE project_id = p.id) as member_count
		FROM projects p
		LEFT JOIN tasks t ON p.id = t.project_id
		WHERE p.id = $1
		GROUP BY p.id, p.name
	`
	var a ProjectAnalytics
	err := r.db.QueryRow(ctx, query, projectID).Scan(
		&a.ProjectID, &a.ProjectName, &a.TotalTasks, &a.CompletedTasks, &a.InProgressTasks, &a.OverdueTasks, &a.AvgCompletionTime, &a.MemberCount,
	)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *AnalyticsRepo) GetEmployeeAnalytics(ctx context.Context, userID int64) (*EmployeeAnalytics, error) {
	query := `
		SELECT 
			$1 as user_id,
			COUNT(t.id) as assigned_tasks,
			COUNT(t.id) FILTER (WHERE t.status = 'done') as completed_tasks,
			COUNT(t.id) FILTER (WHERE t.status = 'in_progress') as in_progress_tasks,
			COUNT(t.id) FILTER (WHERE t.due_date < NOW() AND t.status != 'done') as overdue_tasks,
			COALESCE(AVG(EXTRACT(EPOCH FROM (t.updated_at - t.created_at)) / 86400) FILTER (WHERE t.status = 'done'), 0) as avg_completion_time,
			(SELECT COUNT(DISTINCT project_id) FROM project_members WHERE user_id = $1) as project_count
		FROM tasks t
		WHERE t.assignee_id = $1
	`
	var a EmployeeAnalytics
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&a.UserID, &a.AssignedTasks, &a.CompletedTasks, &a.InProgressTasks, &a.OverdueTasks, &a.AvgCompletionTime, &a.ProjectCount,
	)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *AnalyticsRepo) GetTasksTimeSeries(ctx context.Context, startDate, endDate time.Time, projectID int64) ([]TimeSeriesPoint, error) {
	query := `
		SELECT DATE(created_at) as date, COUNT(*) as count
		FROM tasks
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
		FROM tasks
		WHERE status = 'done' AND updated_at BETWEEN $1 AND $2
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
	query := `SELECT status, COUNT(*) FROM tasks`
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
	query := `SELECT priority, COUNT(*) FROM tasks`
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
