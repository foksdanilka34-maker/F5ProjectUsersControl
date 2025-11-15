package repo

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/foksdanilka34-maker/F5ProjectUsersControl/AnalyticsService/internal/app/analytics"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	pool *pgxpool.Pool
}

func NewStorage(pool *pgxpool.Pool) *Storage {
	return &Storage{
		pool: pool,
	}
}

func (s *Storage) BeginTransaction(ctx context.Context) (pgx.Tx, error) {
	return s.pool.Begin(ctx)
}

func (s *Storage) SaveEmployeeMetrics(ctx context.Context, tx pgx.Tx, metrics *analytics.EmployeeMetrics) error {
	query := `INSERT INTO analytics.employee_metrics 
		(employee_id, metric_date, assigned_tasks, completed_tasks, in_progress_tasks, overdue_tasks,
		on_time_completed_tasks, total_task_duration_seconds) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8) 
		ON CONFLICT (employee_id, metric_date) 
		DO UPDATE SET 
			assigned_tasks = EXCLUDED.assigned_tasks,
			completed_tasks = EXCLUDED.completed_tasks,
			in_progress_tasks = EXCLUDED.in_progress_tasks,
			overdue_tasks = EXCLUDED.overdue_tasks,
			on_time_completed_tasks = EXCLUDED.on_time_completed_tasks,
			total_task_duration_seconds = EXCLUDED.total_task_duration_seconds,
			updated_at = NOW()`

	var runner any
	if tx != nil {
		runner = tx
	} else {
		runner = s.pool
	}

	var err error
	switch r := runner.(type) {
	case pgx.Tx:
		_, err = r.Exec(ctx, query,
			metrics.EmployeeID, metrics.MetricDate, metrics.AssignedTasks, metrics.CompletedTasks,
			metrics.InProgressTasks, metrics.OverdueTasks, metrics.OnTimeCompletionTask,
			metrics.TotalTaskDurationSeconds,
		)
	case *pgxpool.Pool:
		_, err = r.Exec(ctx, query,
			metrics.EmployeeID, metrics.MetricDate, metrics.AssignedTasks, metrics.CompletedTasks,
			metrics.InProgressTasks, metrics.OverdueTasks, metrics.OnTimeCompletionTask,
			metrics.TotalTaskDurationSeconds,
		)
	default:
		err = errors.New("unexpected runner type")
	}

	if err != nil {
		log.Printf("SYSTEM ERROR saving employee metrics (upsert): %v", err)
		return fmt.Errorf("failed to save employee metrics: %w", err)
	}

	return nil
}

func (s *Storage) GetEmployeeMetrics(ctx context.Context, tx pgx.Tx, emplID string) (*analytics.EmployeeMetrics, error) {
	query := `SELECT metric_date, assigned_tasks, completed_tasks, 
			in_progress_tasks, overdue_tasks, on_time_completed_tasks,
			total_task_duration_seconds, created_at, updated_at
			FROM analytics.employee_metrics
			WHERE employee_id = $1
			ORDER BY metric_date DESC
			LIMIT 1`

	var runner any
	if tx != nil {
		runner = tx
	} else {
		runner = s.pool
	}

	var row pgx.Row
	switch r := runner.(type) {
	case pgx.Tx:
		row = r.QueryRow(ctx, query, emplID)
	case *pgxpool.Pool:
		row = r.QueryRow(ctx, query, emplID)
	default:
		return nil, errors.New("unexpected runner type")
	}

	analMetrics := &analytics.EmployeeMetrics{}
	err := row.Scan(
		&analMetrics.MetricDate,
		&analMetrics.AssignedTasks,
		&analMetrics.CompletedTasks,
		&analMetrics.InProgressTasks,
		&analMetrics.OverdueTasks,
		&analMetrics.OnTimeCompletionTask,
		&analMetrics.TotalTaskDurationSeconds,
		&analMetrics.CreatedAt,
		&analMetrics.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to query employee metrics: %w", err)
	}
	analMetrics.EmployeeID = emplID

	return analMetrics, nil
}

func (s *Storage) SaveProjectMetrics(ctx context.Context, tx pgx.Tx, metrics *analytics.ProjectMetrics) error {
	query := `INSERT INTO analytics.project_metrics 
		(project_id, manager_id, metric_date, total_tasks, completed_tasks, in_progress_tasks, overdue_tasks, 
		on_time_completed_tasks, team_size, total_task_duration_seconds, total_priority_weight_completed) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) 
		ON CONFLICT (project_id, metric_date) 
		DO UPDATE SET 
			manager_id = EXCLUDED.manager_id,
			total_tasks = EXCLUDED.total_tasks,
			completed_tasks = EXCLUDED.completed_tasks,
			in_progress_tasks = EXCLUDED.in_progress_tasks,
			overdue_tasks = EXCLUDED.overdue_tasks,
			on_time_completed_tasks = EXCLUDED.on_time_completed_tasks,
			team_size = EXCLUDED.team_size,
			total_task_duration_seconds = EXCLUDED.total_task_duration_seconds,
			total_priority_weight_completed = EXCLUDED.total_priority_weight_completed,
			updated_at = NOW()`

	var runner any
	if tx != nil {
		runner = tx
	} else {
		runner = s.pool
	}

	var err error
	switch r := runner.(type) {
	case pgx.Tx:
		_, err = r.Exec(ctx, query,
			metrics.ProjectID, metrics.ManagerID, metrics.MetricDate, metrics.TotalTasks, metrics.CompletedTasks,
			metrics.InProgressTasks, metrics.OverdueTasks, metrics.OnTimeCompletedTasks, metrics.TeamSize,
			metrics.TotalTaskDurationSecondsCompleted, metrics.TotalPriorityWeightCompleted,
		)
	case *pgxpool.Pool:
		_, err = r.Exec(ctx, query,
			metrics.ProjectID, metrics.ManagerID, metrics.MetricDate, metrics.TotalTasks, metrics.CompletedTasks,
			metrics.InProgressTasks, metrics.OverdueTasks, metrics.OnTimeCompletedTasks, metrics.TeamSize,
			metrics.TotalTaskDurationSecondsCompleted, metrics.TotalPriorityWeightCompleted,
		)
	default:
		err = errors.New("unexpected runner type")
	}

	if err != nil {
		log.Printf("SYSTEM ERROR saving project metrics (upsert): %v", err)
		return fmt.Errorf("failed to save project metrics: %w", err)
	}

	return nil
}

func (s *Storage) GetProjectMetrics(ctx context.Context, tx pgx.Tx, projectID string) (*analytics.ProjectMetrics, error) {
	query := `SELECT manager_id, metric_date, total_tasks, completed_tasks, 
			in_progress_tasks, overdue_tasks, on_time_completed_tasks, team_size,
			total_task_duration_seconds, total_priority_weight_completed, created_at, updated_at
			FROM analytics.project_metrics
			WHERE project_id = $1
			ORDER BY metric_date DESC
			LIMIT 1`

	var runner any
	if tx != nil {
		runner = tx
	} else {
		runner = s.pool
	}

	var row pgx.Row
	switch r := runner.(type) {
	case pgx.Tx:
		row = r.QueryRow(ctx, query, projectID)
	case *pgxpool.Pool:
		row = r.QueryRow(ctx, query, projectID)
	default:
		return nil, errors.New("unexpected runner type")
	}

	analMetrics := &analytics.ProjectMetrics{}
	err := row.Scan(
		&analMetrics.ManagerID,
		&analMetrics.MetricDate,
		&analMetrics.TotalTasks,
		&analMetrics.CompletedTasks,
		&analMetrics.InProgressTasks,
		&analMetrics.OverdueTasks,
		&analMetrics.OnTimeCompletedTasks,
		&analMetrics.TeamSize,
		&analMetrics.TotalTaskDurationSecondsCompleted,
		&analMetrics.TotalPriorityWeightCompleted,
		&analMetrics.CreatedAt,
		&analMetrics.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to query project metrics: %w", err)
	}
	analMetrics.ProjectID = projectID

	return analMetrics, nil
}

func (s *Storage) ListEmployeeMetrics(ctx context.Context, req *analytics.ListEmployeeMetrics) ([]*analytics.EmployeeMetrics, int32, error) {
	countQuery := `SELECT COUNT(DISTINCT employee_id) FROM analytics.employee_metrics WHERE 1=1`
	query := `SELECT employee_id, metric_date, assigned_tasks, completed_tasks, 
			in_progress_tasks, overdue_tasks, on_time_completed_tasks,
			total_task_duration_seconds, created_at, updated_at
			FROM analytics.employee_metrics
			WHERE 1=1`

	args := []any{}
	argIndex := 1

	if req.StartDate != nil {
		query += fmt.Sprintf(" AND metric_date >= $%d", argIndex)
		countQuery += fmt.Sprintf(" AND metric_date >= $%d", argIndex)
		args = append(args, *req.StartDate)
		argIndex++
	}

	if req.EndDate != nil {
		query += fmt.Sprintf(" AND metric_date <= $%d", argIndex)
		countQuery += fmt.Sprintf(" AND metric_date <= $%d", argIndex)
		args = append(args, *req.EndDate)
		argIndex++
	}

	var totalCount int32
	err := s.pool.QueryRow(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count employee metrics: %w", err)
	}

	query += " ORDER BY metric_date DESC"

	if req.PageSize > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, req.PageSize)
		argIndex++

		if req.PageNumber > 0 {
			offset := (req.PageNumber - 1) * req.PageSize
			query += fmt.Sprintf(" OFFSET $%d", argIndex)
			args = append(args, offset)
			argIndex++
		}
	}

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list employee metrics: %w", err)
	}
	defer rows.Close()

	var metrics []*analytics.EmployeeMetrics
	for rows.Next() {
		m := &analytics.EmployeeMetrics{}
		err := rows.Scan(
			&m.EmployeeID,
			&m.MetricDate,
			&m.AssignedTasks,
			&m.CompletedTasks,
			&m.InProgressTasks,
			&m.OverdueTasks,
			&m.OnTimeCompletionTask,
			&m.TotalTaskDurationSeconds,
			&m.CreatedAt,
			&m.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan employee metrics: %w", err)
		}
		metrics = append(metrics, m)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating employee metrics: %w", err)
	}

	return metrics, totalCount, nil
}

func (s *Storage) ListProjectMetrics(ctx context.Context, req *analytics.ListProjectMetrics) ([]*analytics.ProjectMetrics, int32, error) {
	countQuery := `SELECT COUNT(DISTINCT project_id) FROM analytics.project_metrics WHERE 1=1`
	query := `SELECT project_id, manager_id, metric_date, total_tasks, completed_tasks, 
			in_progress_tasks, overdue_tasks, on_time_completed_tasks, team_size,
			total_task_duration_seconds, total_priority_weight_completed, created_at, updated_at
			FROM analytics.project_metrics
			WHERE 1=1`

	args := []any{}
	argIndex := 1

	if req.ManagerID != "" {
		query += fmt.Sprintf(" AND manager_id = $%d", argIndex)
		countQuery += fmt.Sprintf(" AND manager_id = $%d", argIndex)
		args = append(args, req.ManagerID)
		argIndex++
	}

	if req.StartDate != nil {
		query += fmt.Sprintf(" AND metric_date >= $%d", argIndex)
		countQuery += fmt.Sprintf(" AND metric_date >= $%d", argIndex)
		args = append(args, *req.StartDate)
		argIndex++
	}

	if req.EndDate != nil {
		query += fmt.Sprintf(" AND metric_date <= $%d", argIndex)
		countQuery += fmt.Sprintf(" AND metric_date <= $%d", argIndex)
		args = append(args, *req.EndDate)
		argIndex++
	}

	var totalCount int32
	err := s.pool.QueryRow(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count project metrics: %w", err)
	}

	query += " ORDER BY metric_date DESC"

	if req.PageSize != nil && *req.PageSize > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, *req.PageSize)
		argIndex++

		if req.PageNumber != nil && *req.PageNumber > 0 {
			offset := (*req.PageNumber - 1) * *req.PageSize
			query += fmt.Sprintf(" OFFSET $%d", argIndex)
			args = append(args, offset)
			argIndex++
		}
	}

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list project metrics: %w", err)
	}
	defer rows.Close()

	var metrics []*analytics.ProjectMetrics
	for rows.Next() {
		m := &analytics.ProjectMetrics{}
		err := rows.Scan(
			&m.ProjectID,
			&m.ManagerID,
			&m.MetricDate,
			&m.TotalTasks,
			&m.CompletedTasks,
			&m.InProgressTasks,
			&m.OverdueTasks,
			&m.OnTimeCompletedTasks,
			&m.TeamSize,
			&m.TotalTaskDurationSecondsCompleted,
			&m.TotalPriorityWeightCompleted,
			&m.CreatedAt,
			&m.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan project metrics: %w", err)
		}
		metrics = append(metrics, m)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating project metrics: %w", err)
	}

	return metrics, totalCount, nil
}

func (s *Storage) GetTopPerformers(ctx context.Context, limit int32, startDate, endDate *time.Time) ([]*analytics.EmployeeMetrics, error) {
	query := `
		WITH ranked_metrics AS (
			SELECT 
				employee_id,
				metric_date,
				assigned_tasks,
				completed_tasks,
				in_progress_tasks,
				overdue_tasks,
				on_time_completed_tasks,
				total_task_duration_seconds,
				created_at,
				updated_at,
				ROW_NUMBER() OVER (PARTITION BY employee_id ORDER BY metric_date DESC) as rn
			FROM analytics.employee_metrics
			WHERE 1=1`

	args := []any{}
	argIndex := 1

	if startDate != nil {
		query += fmt.Sprintf(" AND metric_date >= $%d", argIndex)
		args = append(args, *startDate)
		argIndex++
	}

	if endDate != nil {
		query += fmt.Sprintf(" AND metric_date <= $%d", argIndex)
		args = append(args, *endDate)
		argIndex++
	}

	query += `
		)
		SELECT 
			employee_id,
			metric_date,
			assigned_tasks,
			completed_tasks,
			in_progress_tasks,
			overdue_tasks,
			on_time_completed_tasks,
			total_task_duration_seconds,
			created_at,
			updated_at
		FROM ranked_metrics
		WHERE rn = 1 AND assigned_tasks > 0
		ORDER BY 
			(CAST(completed_tasks AS FLOAT) / NULLIF(assigned_tasks, 0)) DESC,
			completed_tasks DESC`

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, limit)
	}

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get top performers: %w", err)
	}
	defer rows.Close()

	var metrics []*analytics.EmployeeMetrics
	for rows.Next() {
		m := &analytics.EmployeeMetrics{}
		err := rows.Scan(
			&m.EmployeeID,
			&m.MetricDate,
			&m.AssignedTasks,
			&m.CompletedTasks,
			&m.InProgressTasks,
			&m.OverdueTasks,
			&m.OnTimeCompletionTask,
			&m.TotalTaskDurationSeconds,
			&m.CreatedAt,
			&m.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan top performers: %w", err)
		}
		metrics = append(metrics, m)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating top performers: %w", err)
	}

	return metrics, nil
}

func (s *Storage) GetProductivityTrends(ctx context.Context, req *analytics.ProductivityTrends) ([]*analytics.ProductivityTrend, error) {
	periodFormat := ""
	switch req.Period {
	case analytics.PERIOD_DAILY:
		periodFormat = "day"
	case analytics.PERIOD_WEEKLY:
		periodFormat = "week"
	case analytics.PERIOD_MONTHLY:
		periodFormat = "month"
	default:
		periodFormat = "day"
	}

	query := fmt.Sprintf(`
		SELECT 
			date_trunc('%s', metric_date) as date,
			AVG(CASE 
				WHEN assigned_tasks > 0 
				THEN CAST(completed_tasks AS FLOAT) / assigned_tasks 
				ELSE 0 
			END) * 100 as avg_efficiency,
			SUM(completed_tasks) as total_tasks_completed,
			COUNT(DISTINCT employee_id) as total_employees_active
		FROM analytics.employee_metrics
		WHERE assigned_tasks > 0`, periodFormat)

	args := []any{}
	argIndex := 1

	if req.EmployeeID != nil {
		query += fmt.Sprintf(" AND employee_id = $%d", argIndex)
		args = append(args, *req.EmployeeID)
		argIndex++
	}

	query += fmt.Sprintf(" GROUP BY date_trunc('%s', metric_date) ORDER BY date DESC", periodFormat)

	if req.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, req.Limit)
	}

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get productivity trends: %w", err)
	}
	defer rows.Close()

	var trends []*analytics.ProductivityTrend
	for rows.Next() {
		t := &analytics.ProductivityTrend{}
		err := rows.Scan(
			&t.Date,
			&t.AvgEfficiency,
			&t.TotalTasksCompleted,
			&t.TotalEmployeesActive,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan productivity trend: %w", err)
		}
		trends = append(trends, t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating productivity trends: %w", err)
	}

	return trends, nil
}

func (s *Storage) GetCompletionRateTrends(ctx context.Context, req *analytics.ComletionRateTrends) ([]*analytics.CompletionRateTrend, error) {
	periodFormat := ""
	switch req.Period {
	case analytics.PERIOD_DAILY:
		periodFormat = "day"
	case analytics.PERIOD_WEEKLY:
		periodFormat = "week"
	case analytics.PERIOD_MONTHLY:
		periodFormat = "month"
	default:
		periodFormat = "day"
	}

	var query string
	args := []any{}
	argIndex := 1

	if req.ProjectID != "" {
		query = fmt.Sprintf(`
			SELECT 
				date_trunc('%s', metric_date) as date,
				AVG(CASE 
					WHEN completed_tasks > 0 
					THEN (CAST(on_time_completed_tasks AS FLOAT) / completed_tasks) * 100 
					ELSE 0 
				END) as on_time_rate,
				AVG(CASE 
					WHEN total_tasks > 0 
					THEN (CAST(completed_tasks AS FLOAT) / total_tasks) * 100 
					ELSE 0 
				END) as overall_rate,
				SUM(completed_tasks) as completed_count,
				SUM(overdue_tasks) as overdue_count
			FROM analytics.project_metrics
			WHERE project_id = $1
			GROUP BY date_trunc('%s', metric_date)
			ORDER BY date DESC`, periodFormat, periodFormat)
		args = append(args, req.ProjectID)
		argIndex++
	} else {
		query = fmt.Sprintf(`
			SELECT 
				date_trunc('%s', metric_date) as date,
				AVG(CASE 
					WHEN completed_tasks > 0 
					THEN (CAST(on_time_completed_tasks AS FLOAT) / completed_tasks) * 100 
					ELSE 0 
				END) as on_time_rate,
				AVG(CASE 
					WHEN assigned_tasks > 0 
					THEN (CAST(completed_tasks AS FLOAT) / assigned_tasks) * 100 
					ELSE 0 
				END) as overall_rate,
				SUM(completed_tasks) as completed_count,
				SUM(overdue_tasks) as overdue_count
			FROM analytics.employee_metrics
			WHERE assigned_tasks > 0
			GROUP BY date_trunc('%s', metric_date)
			ORDER BY date DESC`, periodFormat, periodFormat)
	}

	if req.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, req.Limit)
	}

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get completion rate trends: %w", err)
	}
	defer rows.Close()

	var trends []*analytics.CompletionRateTrend
	for rows.Next() {
		t := &analytics.CompletionRateTrend{}
		err := rows.Scan(
			&t.Date,
			&t.OnTimeRate,
			&t.OverallRate,
			&t.CompletedCount,
			&t.OverDueCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan completion rate trend: %w", err)
		}
		trends = append(trends, t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating completion rate trends: %w", err)
	}

	return trends, nil
}

func (s *Storage) GetDashboardStats(ctx context.Context, startDate, endDate *time.Time) (*analytics.DashboardStats, error) {
	stats := &analytics.DashboardStats{
		CalculatedAt: time.Now(),
	}

	employeeQuery := `
		SELECT 
			COUNT(DISTINCT employee_id) as total_employees,
			COUNT(DISTINCT CASE WHEN assigned_tasks > 0 THEN employee_id END) as active_employees
		FROM analytics.employee_metrics
		WHERE 1=1`

	employeeArgs := []any{}
	argIndex := 1

	if startDate != nil {
		employeeQuery += fmt.Sprintf(" AND metric_date >= $%d", argIndex)
		employeeArgs = append(employeeArgs, *startDate)
		argIndex++
	}

	if endDate != nil {
		employeeQuery += fmt.Sprintf(" AND metric_date <= $%d", argIndex)
		employeeArgs = append(employeeArgs, *endDate)
	}

	err := s.pool.QueryRow(ctx, employeeQuery, employeeArgs...).Scan(
		&stats.TotalEmployees,
		&stats.ActiveEmployees,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get employee counts: %w", err)
	}

	projectQuery := `
		SELECT 
			COUNT(DISTINCT project_id) as total_projects,
			COUNT(DISTINCT CASE WHEN total_tasks > 0 THEN project_id END) as active_projects
		FROM analytics.project_metrics
		WHERE 1=1`

	projectArgs := []any{}
	argIndex = 1

	if startDate != nil {
		projectQuery += fmt.Sprintf(" AND metric_date >= $%d", argIndex)
		projectArgs = append(projectArgs, *startDate)
		argIndex++
	}

	if endDate != nil {
		projectQuery += fmt.Sprintf(" AND metric_date <= $%d", argIndex)
		projectArgs = append(projectArgs, *endDate)
	}

	err = s.pool.QueryRow(ctx, projectQuery, projectArgs...).Scan(
		&stats.TotalProjects,
		&stats.ActiveProjects,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get project counts: %w", err)
	}

	taskQuery := `
		SELECT 
			SUM(assigned_tasks) as total_tasks,
			SUM(completed_tasks) as completed_tasks,
			SUM(overdue_tasks) as overdue_tasks,
			AVG(CASE 
				WHEN assigned_tasks > 0 
				THEN (CAST(completed_tasks AS FLOAT) / assigned_tasks) * 100 
				ELSE 0 
			END) as avg_efficiency,
			AVG(CASE 
				WHEN completed_tasks > 0 
				THEN (CAST(on_time_completed_tasks AS FLOAT) / completed_tasks) * 100 
				ELSE 0 
			END) as avg_on_time_rate
		FROM analytics.employee_metrics
		WHERE assigned_tasks > 0`

	taskArgs := []any{}
	argIndex = 1

	if startDate != nil {
		taskQuery += fmt.Sprintf(" AND metric_date >= $%d", argIndex)
		taskArgs = append(taskArgs, *startDate)
		argIndex++
	}

	if endDate != nil {
		taskQuery += fmt.Sprintf(" AND metric_date <= $%d", argIndex)
		taskArgs = append(taskArgs, *endDate)
	}

	var avgEfficiency, avgOnTimeRate float64
	err = s.pool.QueryRow(ctx, taskQuery, taskArgs...).Scan(
		&stats.TotalTasks,
		&stats.CompletedTasks,
		&stats.OverDueTasks,
		&avgEfficiency,
		&avgOnTimeRate,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get task stats: %w", err)
	}

	stats.AvgCompanyEfficiency = float32(avgEfficiency)
	stats.AvgOnTimeRate = float32(avgOnTimeRate)

	topEmployeesQuery := `
		WITH ranked_metrics AS (
			SELECT 
				employee_id,
				assigned_tasks,
				completed_tasks,
				on_time_completed_tasks,
				total_task_duration_seconds,
				overdue_tasks,
				ROW_NUMBER() OVER (PARTITION BY employee_id ORDER BY metric_date DESC) as rn
			FROM analytics.employee_metrics
			WHERE assigned_tasks > 0`

	topArgs := []any{}
	argIndex = 1

	if startDate != nil {
		topEmployeesQuery += fmt.Sprintf(" AND metric_date >= $%d", argIndex)
		topArgs = append(topArgs, *startDate)
		argIndex++
	}

	if endDate != nil {
		topEmployeesQuery += fmt.Sprintf(" AND metric_date <= $%d", argIndex)
		topArgs = append(topArgs, *endDate)
		argIndex++
	}

	topEmployeesQuery += `
		)
		SELECT 
			employee_id,
			completed_tasks,
			assigned_tasks,
			on_time_completed_tasks,
			total_task_duration_seconds,
			overdue_tasks
		FROM ranked_metrics
		WHERE rn = 1
		ORDER BY completed_tasks DESC
		LIMIT 10`

	rows, err := s.pool.Query(ctx, topEmployeesQuery, topArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to get top employees: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var employeeID string
		var completedTasks, assignedTasks, onTimeCompleted, overdueTasks int32
		var totalDuration int64

		err := rows.Scan(&employeeID, &completedTasks, &assignedTasks, &onTimeCompleted, &totalDuration, &overdueTasks)
		if err != nil {
			return nil, fmt.Errorf("failed to scan top employee: %w", err)
		}

		taskCompletionRate := float64(0)
		if assignedTasks > 0 {
			taskCompletionRate = (float64(completedTasks) / float64(assignedTasks)) * 100
		}

		onTimeRate := float64(0)
		if completedTasks > 0 {
			onTimeRate = (float64(onTimeCompleted) / float64(completedTasks)) * 100
		}

		speedBonus := 1.0
		if completedTasks > 0 {
			avgDuration := float64(totalDuration) / float64(completedTasks)
			speedBonus = 28800.0 / (avgDuration + 1)
			if speedBonus > 1.5 {
				speedBonus = 1.5
			}
		}

		overduePenalty := float64(overdueTasks) * 5.0
		baseScore := (onTimeRate * 0.7) + (taskCompletionRate * 0.3)
		efficiencyScore := (baseScore * speedBonus) - overduePenalty
		if efficiencyScore < 0 {
			efficiencyScore = 0
		}

		stats.TopEmployees = append(stats.TopEmployees, analytics.TopEmployee{
			ID:              employeeID,
			EfficiencyScore: float32(efficiencyScore),
			TaskCompleted:   completedTasks,
		})
	}

	problematicQuery := `
		WITH ranked_metrics AS (
			SELECT 
				project_id,
				total_tasks,
				completed_tasks,
				on_time_completed_tasks,
				overdue_tasks,
				ROW_NUMBER() OVER (PARTITION BY project_id ORDER BY metric_date DESC) as rn
			FROM analytics.project_metrics
			WHERE total_tasks > 0`

	probArgs := []any{}
	argIndex = 1

	if startDate != nil {
		problematicQuery += fmt.Sprintf(" AND metric_date >= $%d", argIndex)
		probArgs = append(probArgs, *startDate)
		argIndex++
	}

	if endDate != nil {
		problematicQuery += fmt.Sprintf(" AND metric_date <= $%d", argIndex)
		probArgs = append(probArgs, *endDate)
		argIndex++
	}

	problematicQuery += `
		)
		SELECT 
			project_id,
			total_tasks,
			completed_tasks,
			on_time_completed_tasks,
			overdue_tasks
		FROM ranked_metrics
		WHERE rn = 1
		ORDER BY 
			(CASE 
				WHEN total_tasks > 0 
				THEN (CAST(completed_tasks AS FLOAT) / total_tasks) * 100 
				ELSE 0 
			END) ASC
		LIMIT 10`

	probRows, err := s.pool.Query(ctx, problematicQuery, probArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to get problematic projects: %w", err)
	}
	defer probRows.Close()

	for probRows.Next() {
		var projectID string
		var totalTasks, completedTasks, onTimeCompleted, overdueTasks int32

		err := probRows.Scan(&projectID, &totalTasks, &completedTasks, &onTimeCompleted, &overdueTasks)
		if err != nil {
			return nil, fmt.Errorf("failed to scan problematic project: %w", err)
		}

		deliveryPerf := float64(0)
		if totalTasks > 0 {
			deliveryPerf = (float64(completedTasks) / float64(totalTasks)) * 100
		}

		schedulePerf := float64(0)
		if completedTasks > 0 {
			schedulePerf = (float64(onTimeCompleted) / float64(completedTasks)) * 100
		}

		healthIndex := (schedulePerf * 0.6) + (deliveryPerf * 0.4) - (float64(overdueTasks) * 5.0)
		if healthIndex < 0 {
			healthIndex = 0
		}

		stats.ProblematicProjects = append(stats.ProblematicProjects, analytics.BottomProject{
			ProjectID:   projectID,
			HealthScore: float32(healthIndex),
			OnTimeRate:  float32(schedulePerf),
		})
	}

	return stats, nil
}
