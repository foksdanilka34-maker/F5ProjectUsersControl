package repo

import (
	"context"
	"errors"
	"fmt"
	"log"

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
		on_time_completed_tasks, total_priority_weight_completed, total_task_duration_seconds) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) 
		ON CONFLICT (employee_id, metric_date) 
		DO UPDATE SET 
			assigned_tasks = $3,
			completed_tasks = $4,
			in_progress_tasks = $5,
			overdue_tasks = $6,
			on_time_completed_tasks = $7,
			total_priority_weight_completed = $8,
			total_task_duration_seconds = $9,
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
			metrics.TotalPrioritWeight, metrics.TotalTaskDurationSeconds,
		)
	case *pgxpool.Pool:
		_, err = r.Exec(ctx, query,
			metrics.EmployeeID, metrics.MetricDate, metrics.AssignedTasks, metrics.CompletedTasks,
			metrics.InProgressTasks, metrics.OverdueTasks, metrics.OnTimeCompletionTask,
			metrics.TotalPrioritWeight, metrics.TotalTaskDurationSeconds,
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

func (s *Storage) GetEmployeeMetrics(ctx context.Context, emplID string) (*analytics.EmployeeMetrics, error) {
	query := `SELECT metric_date, assigned_tasks, completed_tasks, 
			in_progress_tasks, overdue_tasks, on_time_completed_tasks, total_priority_weight_completed,
			total_task_duration_seconds, created_at, updated_at
			FROM analytics.employee_metrics
			WHERE employee_id = $1
			ORDER BY metric_date DESC
			LIMIT 1`

	analMetrics := &analytics.EmployeeMetrics{}
	err := s.pool.QueryRow(ctx, query, emplID).Scan(
		&analMetrics.MetricDate,
		&analMetrics.AssignedTasks,
		&analMetrics.CompletedTasks,
		&analMetrics.InProgressTasks,
		&analMetrics.OverdueTasks,
		&analMetrics.OnTimeCompletionTask,
		&analMetrics.TotalPrioritWeight,
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

// func (s *Storage) ListEmployeeMetrics(ctx context.Context, pageSize, pageNumber int32, departmentID string, startDate, endDate time.Time) ([]*analytics.EmployeeMetrics, int32, error) {
// 	whereClause := "metric_date >= $1 AND metric_date <= $2"
// 	args := []interface{}{startDate, endDate}

// 	countQuery := "SELECT COUNT(DISTINCT employee_id) FROM analytics.employee_metrics WHERE " + whereClause
// 	var totalCount int32
// 	err := s.pool.QueryRow(ctx, countQuery, args...).Scan(&totalCount)
// 	if err != nil {
// 		return nil, 0, fmt.Errorf("failed to count employee metrics: %w", err)
// 	}

// 	offset := (pageNumber - 1) * pageSize
// 	args = append(args, pageSize, offset)
// 	query := `
// 		SELECT 
// 			id, employee_id, employee_name, metric_date,
// 			assigned_tasks, completed_tasks, in_progress_tasks, overdue_tasks,
// 			efficiency_score, task_completion_rate, on_time_completion_rate,
// 			created_at, updated_at
// 		FROM analytics.employee_metrics
// 		WHERE ` + whereClause + `
// 		ORDER BY metric_date DESC, efficiency_score DESC
// 		LIMIT $3 OFFSET $4
// 	`

// 	rows, err := s.pool.Query(ctx, query, args...)
// 	if err != nil {
// 		return nil, 0, fmt.Errorf("failed to query employee metrics: %w", err)
// 	}
// 	defer rows.Close()

// 	var metrics []*analytics.EmployeeMetrics
// 	for rows.Next() {
// 		m := &analytics.EmployeeMetrics{}
// 		err := rows.Scan(
// 			&m.ID, &m.EmployeeID, &m.EmployeeName, &m.MetricDate,
// 			&m.TasksAssigned, &m.TasksCompleted, &m.InProgressTasks, &m.OverdueTasks,
// 			&m.EfficiencyScore, &m.TaskCompletionRate, &m.OnTimeCompletionRate,
// 			&m.CreatedAt, &m.UpdatedAt,
// 		)
// 		if err != nil {
// 			return nil, 0, fmt.Errorf("failed to scan employee metrics: %w", err)
// 		}
// 		metrics = append(metrics, m)
// 	}

// 	return metrics, totalCount, rows.Err()
// }

// func (s *Storage) GetTopPerformers(ctx context.Context, limit int32, departmentID string, startDate, endDate time.Time) ([]*analytics.EmployeeMetrics, error) {
// 	query := `
// 		SELECT 
// 			id, employee_id, employee_name, metric_date,
// 			assigned_tasks, completed_tasks, in_progress_tasks, overdue_tasks,
// 			efficiency_score, task_completion_rate, on_time_completion_rate,
// 			created_at, updated_at
// 		FROM analytics.employee_metrics
// 		WHERE metric_date >= $1 AND metric_date <= $2
// 		ORDER BY efficiency_score DESC, metric_date DESC
// 		LIMIT $3
// 	`

// 	rows, err := s.pool.Query(ctx, query, startDate, endDate, limit)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to query top performers: %w", err)
// 	}
// 	defer rows.Close()

// 	var metrics []*analytics.EmployeeMetrics
// 	for rows.Next() {
// 		m := &analytics.EmployeeMetrics{}
// 		err := rows.Scan(
// 			&m.ID, &m.EmployeeID, &m.EmployeeName, &m.MetricDate,
// 			&m.TasksAssigned, &m.TasksCompleted, &m.InProgressTasks, &m.OverdueTasks,
// 			&m.EfficiencyScore, &m.TaskCompletionRate, &m.OnTimeCompletionRate,
// 			&m.CreatedAt, &m.UpdatedAt,
// 		)
// 		if err != nil {
// 			return nil, fmt.Errorf("failed to scan top performers: %w", err)
// 		}
// 		metrics = append(metrics, m)
// 	}

// 	return metrics, rows.Err()
// }

// func (s *Storage) SaveProjectMetrics(ctx context.Context, metrics *analytics.ProjectMetrics) error {
// 	query := `
// 		INSERT INTO analytics.project_metrics (
// 			project_id, project_name, manager_id, manager_name, metric_date,
// 			total_tasks, completed_tasks, in_progress_tasks, overdue_tasks,
// 			delivery_performance, schedule_performance, quality_performance, team_performance,
// 			health_index, risk_score, health_status, velocity, projected_end_date,
// 			team_capacity_utilization, team_size, avg_team_efficiency, is_at_risk, days_until_due
// 		) VALUES (
// 			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23
// 		)
// 		ON CONFLICT (project_id, metric_date) DO UPDATE SET
// 			project_name = EXCLUDED.project_name,
// 			manager_name = EXCLUDED.manager_name,
// 			total_tasks = EXCLUDED.total_tasks,
// 			completed_tasks = EXCLUDED.completed_tasks,
// 			in_progress_tasks = EXCLUDED.in_progress_tasks,
// 			overdue_tasks = EXCLUDED.overdue_tasks,
// 			delivery_performance = EXCLUDED.delivery_performance,
// 			schedule_performance = EXCLUDED.schedule_performance,
// 			quality_performance = EXCLUDED.quality_performance,
// 			team_performance = EXCLUDED.team_performance,
// 			health_index = EXCLUDED.health_index,
// 			risk_score = EXCLUDED.risk_score,
// 			health_status = EXCLUDED.health_status,
// 			velocity = EXCLUDED.velocity,
// 			projected_end_date = EXCLUDED.projected_end_date,
// 			team_capacity_utilization = EXCLUDED.team_capacity_utilization,
// 			team_size = EXCLUDED.team_size,
// 			avg_team_efficiency = EXCLUDED.avg_team_efficiency,
// 			is_at_risk = EXCLUDED.is_at_risk,
// 			days_until_due = EXCLUDED.days_until_due,
// 			updated_at = NOW()
// 		RETURNING id
// 	`

// 	var id string
// 	err := s.pool.QueryRow(ctx, query,
// 		metrics.ProjectID, metrics.ProjectName, metrics.ManagerID, metrics.ManagerName, metrics.MetricDate,
// 		metrics.TotalTasks, metrics.CompletedTasks, metrics.InProgressTasks, metrics.OverdueTasks,
// 		metrics.DeliveryPerformance, metrics.SchedulePerformance, metrics.QualityPerformance, metrics.TeamPerformance,
// 		metrics.HealthIndex, metrics.RiskScore, metrics.HealthStatus, metrics.Velocity, metrics.ProjectedEndDate,
// 		metrics.TeamCapacityUtilization, metrics.TeamSize, metrics.AvgTeamEfficiency, metrics.IsAtRisk, metrics.DaysUntilDue,
// 	).Scan(&id)

// 	if err != nil {
// 		return fmt.Errorf("failed to save project metrics: %w", err)
// 	}

// 	metrics.ID = id
// 	return nil
// }

// func (s *Storage) GetProjectMetrics(ctx context.Context, projectID string, startDate, endDate time.Time) ([]*analytics.ProjectMetrics, error) {
// 	query := `
// 		SELECT 
// 			id, project_id, project_name, manager_id, manager_name, metric_date,
// 			total_tasks, completed_tasks, in_progress_tasks, overdue_tasks,
// 			delivery_performance, schedule_performance, quality_performance, team_performance,
// 			health_index, risk_score, health_status, velocity, projected_end_date,
// 			team_capacity_utilization, team_size, avg_team_efficiency, is_at_risk, days_until_due,
// 			created_at, updated_at
// 		FROM analytics.project_metrics
// 		WHERE project_id = $1
// 			AND metric_date >= $2
// 			AND metric_date <= $3
// 		ORDER BY metric_date DESC
// 	`

// 	rows, err := s.pool.Query(ctx, query, projectID, startDate, endDate)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to query project metrics: %w", err)
// 	}
// 	defer rows.Close()

// 	var metrics []*analytics.ProjectMetrics
// 	for rows.Next() {
// 		m := &analytics.ProjectMetrics{}
// 		err := rows.Scan(
// 			&m.ID, &m.ProjectID, &m.ProjectName, &m.ManagerID, &m.ManagerName, &m.MetricDate,
// 			&m.TotalTasks, &m.CompletedTasks, &m.InProgressTasks, &m.OverdueTasks,
// 			&m.DeliveryPerformance, &m.SchedulePerformance, &m.QualityPerformance, &m.TeamPerformance,
// 			&m.HealthIndex, &m.RiskScore, &m.HealthStatus, &m.Velocity, &m.ProjectedEndDate,
// 			&m.TeamCapacityUtilization, &m.TeamSize, &m.AvgTeamEfficiency, &m.IsAtRisk, &m.DaysUntilDue,
// 			&m.CreatedAt, &m.UpdatedAt,
// 		)
// 		if err != nil {
// 			return nil, fmt.Errorf("failed to scan project metrics: %w", err)
// 		}
// 		metrics = append(metrics, m)
// 	}

// 	return metrics, rows.Err()
// }

// func (s *Storage) ListProjectMetrics(ctx context.Context, pageSize, pageNumber int32, managerID string, startDate, endDate time.Time) ([]*analytics.ProjectMetrics, int32, error) {
// 	whereClause := "metric_date >= $1 AND metric_date <= $2"
// 	args := []interface{}{startDate, endDate}

// 	countQuery := "SELECT COUNT(DISTINCT project_id) FROM analytics.project_metrics WHERE " + whereClause
// 	var totalCount int32
// 	err := s.pool.QueryRow(ctx, countQuery, args...).Scan(&totalCount)
// 	if err != nil {
// 		return nil, 0, fmt.Errorf("failed to count project metrics: %w", err)
// 	}

// 	offset := (pageNumber - 1) * pageSize
// 	args = append(args, pageSize, offset)
// 	query := `
// 		SELECT 
// 			id, project_id, project_name, manager_id, manager_name, metric_date,
// 			total_tasks, completed_tasks, in_progress_tasks, overdue_tasks,
// 			delivery_performance, schedule_performance, quality_performance, team_performance,
// 			health_index, risk_score, health_status, velocity, projected_end_date,
// 			team_capacity_utilization, team_size, avg_team_efficiency, is_at_risk, days_until_due,
// 			created_at, updated_at
// 		FROM analytics.project_metrics
// 		WHERE ` + whereClause + `
// 		ORDER BY metric_date DESC, health_index DESC
// 		LIMIT $3 OFFSET $4
// 	`

// 	rows, err := s.pool.Query(ctx, query, args...)
// 	if err != nil {
// 		return nil, 0, fmt.Errorf("failed to query project metrics: %w", err)
// 	}
// 	defer rows.Close()

// 	var metrics []*analytics.ProjectMetrics
// 	for rows.Next() {
// 		m := &analytics.ProjectMetrics{}
// 		err := rows.Scan(
// 			&m.ID, &m.ProjectID, &m.ProjectName, &m.ManagerID, &m.ManagerName, &m.MetricDate,
// 			&m.TotalTasks, &m.CompletedTasks, &m.InProgressTasks, &m.OverdueTasks,
// 			&m.DeliveryPerformance, &m.SchedulePerformance, &m.QualityPerformance, &m.TeamPerformance,
// 			&m.HealthIndex, &m.RiskScore, &m.HealthStatus, &m.Velocity, &m.ProjectedEndDate,
// 			&m.TeamCapacityUtilization, &m.TeamSize, &m.AvgTeamEfficiency, &m.IsAtRisk, &m.DaysUntilDue,
// 			&m.CreatedAt, &m.UpdatedAt,
// 		)
// 		if err != nil {
// 			return nil, 0, fmt.Errorf("failed to scan project metrics: %w", err)
// 		}
// 		metrics = append(metrics, m)
// 	}

// 	return metrics, totalCount, rows.Err()
// }

// func (s *Storage) CalculateProductivityTrends(ctx context.Context, period string, limit int32, departmentID, employeeID string) ([]map[string]interface{}, error) {
// 	query := `
// 		SELECT 
// 			DATE(metric_date) as date,
// 			AVG(efficiency_score) as avg_efficiency,
// 			COUNT(*) as total_employees_active,
// 			0 as total_tasks_completed
// 		FROM analytics.employee_metrics
// 		WHERE metric_date >= NOW() - INTERVAL '90 days'
// 		GROUP BY DATE(metric_date)
// 		ORDER BY DATE(metric_date) DESC
// 		LIMIT $1
// 	`

// 	rows, err := s.pool.Query(ctx, query, limit)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to query productivity trends: %w", err)
// 	}
// 	defer rows.Close()

// 	var trends []map[string]interface{}
// 	for rows.Next() {
// 		var date time.Time
// 		var avgEfficiency float64
// 		var totalEmployeesActive int32
// 		var totalTasksCompleted int32

// 		err := rows.Scan(&date, &avgEfficiency, &totalEmployeesActive, &totalTasksCompleted)
// 		if err != nil {
// 			return nil, fmt.Errorf("failed to scan productivity trend: %w", err)
// 		}

// 		trends = append(trends, map[string]interface{}{
// 			"date":                   date,
// 			"avg_efficiency":         avgEfficiency,
// 			"total_employees_active": totalEmployeesActive,
// 			"total_tasks_completed":  totalTasksCompleted,
// 		})
// 	}

// 	return trends, rows.Err()
// }

// func (s *Storage) CalculateCompletionRateTrends(ctx context.Context, period string, limit int32, projectID, departmentID string) ([]map[string]interface{}, error) {
// 	query := `
// 		SELECT 
// 			DATE(metric_date) as date,
// 			AVG(CASE WHEN completed_tasks > 0 
// 				THEN (completed_tasks - overdue_tasks)::float / completed_tasks 
// 				ELSE 0 END) as on_time_rate,
// 			AVG(CASE WHEN total_tasks > 0 
// 				THEN completed_tasks::float / total_tasks 
// 				ELSE 0 END) as overall_rate,
// 			SUM(completed_tasks) as completed_count,
// 			SUM(overdue_tasks) as overdue_count
// 		FROM analytics.project_metrics
// 		WHERE metric_date >= NOW() - INTERVAL '90 days'
// 		GROUP BY DATE(metric_date)
// 		ORDER BY DATE(metric_date) DESC
// 		LIMIT $1
// 	`

// 	rows, err := s.pool.Query(ctx, query, limit)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to query completion rate trends: %w", err)
// 	}
// 	defer rows.Close()

// 	var trends []map[string]interface{}
// 	for rows.Next() {
// 		var date time.Time
// 		var onTimeRate, overallRate float64
// 		var completedCount, overdueCount int32

// 		err := rows.Scan(&date, &onTimeRate, &overallRate, &completedCount, &overdueCount)
// 		if err != nil {
// 			return nil, fmt.Errorf("failed to scan completion rate trend: %w", err)
// 		}

// 		trends = append(trends, map[string]interface{}{
// 			"date":            date,
// 			"on_time_rate":    onTimeRate,
// 			"overall_rate":    overallRate,
// 			"completed_count": completedCount,
// 			"overdue_count":   overdueCount,
// 		})
// 	}

// 	return trends, rows.Err()
// }

// func (s *Storage) GetDashboardStats(ctx context.Context, startDate, endDate time.Time) (map[string]interface{}, error) {
// 	query := `
// 		WITH employee_stats AS (
// 			SELECT 
// 				COUNT(DISTINCT employee_id) AS total_employees,
// 				COUNT(DISTINCT employee_id) FILTER (WHERE metric_date >= NOW() - INTERVAL '7 days') AS active_employees,
// 				COALESCE(AVG(efficiency_score), 0) AS avg_company_efficiency,
// 				COALESCE(AVG(on_time_completion_rate), 0) AS avg_on_time_rate
// 			FROM analytics.employee_metrics
// 			WHERE metric_date >= $1 AND metric_date <= $2
// 		), project_stats AS (
// 			SELECT 
// 				COUNT(DISTINCT project_id) AS total_projects,
// 				COUNT(DISTINCT project_id) FILTER (WHERE in_progress_tasks > 0) AS active_projects,
// 				COALESCE(SUM(total_tasks), 0) AS total_tasks,
// 				COALESCE(SUM(completed_tasks), 0) AS completed_tasks,
// 				COALESCE(SUM(overdue_tasks), 0) AS overdue_tasks
// 			FROM analytics.project_metrics
// 			WHERE metric_date >= $1 AND metric_date <= $2
// 		)
// 		SELECT 
// 			employee_stats.total_employees,
// 			employee_stats.active_employees,
// 			project_stats.total_projects,
// 			project_stats.active_projects,
// 			project_stats.total_tasks,
// 			project_stats.completed_tasks,
// 			project_stats.overdue_tasks,
// 			employee_stats.avg_company_efficiency,
// 			employee_stats.avg_on_time_rate
// 		FROM employee_stats, project_stats
// 	`

// 	var totalEmployees, activeEmployees, totalProjects, activeProjects int32
// 	var totalTasks, completedTasks, overdueTasks int32
// 	var avgCompanyEfficiency, avgOnTimeRate float64

// 	err := s.pool.QueryRow(ctx, query, startDate, endDate).Scan(
// 		&totalEmployees, &activeEmployees, &totalProjects, &activeProjects,
// 		&totalTasks, &completedTasks, &overdueTasks,
// 		&avgCompanyEfficiency, &avgOnTimeRate,
// 	)

// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get dashboard stats: %w", err)
// 	}

// 	return map[string]interface{}{
// 		"total_employees":        totalEmployees,
// 		"active_employees":       activeEmployees,
// 		"total_projects":         totalProjects,
// 		"active_projects":        activeProjects,
// 		"total_tasks":            totalTasks,
// 		"completed_tasks":        completedTasks,
// 		"overdue_tasks":          overdueTasks,
// 		"avg_company_efficiency": avgCompanyEfficiency,
// 		"avg_on_time_rate":       avgOnTimeRate,
// 	}, nil
// }
