package repo

import (
	"context"
	"fmt"
	"time"

	"github.com/foksdanilka34-maker/F5ProjectUsersControl/AnalyticsService/internal/app/analytics"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	pool *pgxpool.Pool
}

func NewStorage(pool *pgxpool.Pool) *Storage {
	return &Storage{pool: pool}
}

func (s *Storage) SaveEmployeeMetrics(ctx context.Context, metrics *analytics.EmployeeMetrics) error {
	query := `
		INSERT INTO analytics.employee_metrics (
			employee_id, employee_name, department_id, position_id,
			metric_date, tasks_completed, tasks_assigned,
			avg_completion_time_hours, on_time_completion_rate,
			avg_task_priority, skills_used, projects_involved,
			efficiency_score
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		)
		ON CONFLICT (employee_id, metric_date) DO UPDATE SET
			employee_name = EXCLUDED.employee_name,
			department_id = EXCLUDED.department_id,
			position_id = EXCLUDED.position_id,
			tasks_completed = EXCLUDED.tasks_completed,
			tasks_assigned = EXCLUDED.tasks_assigned,
			avg_completion_time_hours = EXCLUDED.avg_completion_time_hours,
			on_time_completion_rate = EXCLUDED.on_time_completion_rate,
			avg_task_priority = EXCLUDED.avg_task_priority,
			skills_used = EXCLUDED.skills_used,
			projects_involved = EXCLUDED.projects_involved,
			efficiency_score = EXCLUDED.efficiency_score,
			updated_at = NOW()
		RETURNING id
	`

	var id string
	err := s.pool.QueryRow(ctx, query,
		metrics.EmployeeID, metrics.EmployeeName, metrics.DepartmentID, metrics.PositionID,
		metrics.MetricDate, metrics.TasksCompleted, metrics.TasksAssigned,
		metrics.AvgCompletionTimeHours, metrics.OnTimeCompletionRate,
		metrics.AvgTaskPriority, metrics.SkillsUsed, metrics.ProjectsInvolved,
		metrics.EfficiencyScore,
	).Scan(&id)

	if err != nil {
		return fmt.Errorf("failed to save employee metrics: %w", err)
	}

	metrics.ID = id
	return nil
}

func (s *Storage) GetEmployeeMetrics(ctx context.Context, employeeID string, startDate, endDate time.Time) ([]*analytics.EmployeeMetrics, error) {
	query := `
		SELECT 
			id, employee_id, employee_name, department_id, position_id,
			metric_date, tasks_completed, tasks_assigned,
			avg_completion_time_hours, on_time_completion_rate,
			avg_task_priority, skills_used, projects_involved,
			efficiency_score, created_at, updated_at
		FROM analytics.employee_metrics
		WHERE employee_id = $1
			AND metric_date >= $2
			AND metric_date <= $3
		ORDER BY metric_date DESC
	`

	rows, err := s.pool.Query(ctx, query, employeeID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to query employee metrics: %w", err)
	}
	defer rows.Close()

	var metrics []*analytics.EmployeeMetrics
	for rows.Next() {
		m := &analytics.EmployeeMetrics{}
		err := rows.Scan(
			&m.ID, &m.EmployeeID, &m.EmployeeName, &m.DepartmentID, &m.PositionID,
			&m.MetricDate, &m.TasksCompleted, &m.TasksAssigned,
			&m.AvgCompletionTimeHours, &m.OnTimeCompletionRate,
			&m.AvgTaskPriority, &m.SkillsUsed, &m.ProjectsInvolved,
			&m.EfficiencyScore, &m.CreatedAt, &m.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan employee metrics: %w", err)
		}
		metrics = append(metrics, m)
	}

	return metrics, rows.Err()
}

func (s *Storage) ListEmployeeMetrics(ctx context.Context, pageSize, pageNumber int32, departmentID string, startDate, endDate time.Time) ([]*analytics.EmployeeMetrics, int32, error) {
	whereClause := "metric_date >= $1 AND metric_date <= $2"
	args := []interface{}{startDate, endDate}

	if departmentID != "" {
		args = append(args, departmentID)
		whereClause += fmt.Sprintf(" AND department_id = $%d", len(args))
	}

	countQuery := "SELECT COUNT(*) FROM analytics.employee_metrics WHERE " + whereClause
	var totalCount int32
	err := s.pool.QueryRow(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count employee metrics: %w", err)
	}

	offset := (pageNumber - 1) * pageSize
	args = append(args, pageSize, offset)
	query := `
		SELECT 
			id, employee_id, employee_name, department_id, position_id,
			metric_date, tasks_completed, tasks_assigned,
			avg_completion_time_hours, on_time_completion_rate,
			avg_task_priority, skills_used, projects_involved,
			efficiency_score, created_at, updated_at
		FROM analytics.employee_metrics
		WHERE ` + whereClause + `
		ORDER BY metric_date DESC, efficiency_score DESC
		LIMIT $` + fmt.Sprintf("%d", len(args)-1) + ` OFFSET $` + fmt.Sprintf("%d", len(args)) + ``

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query employee metrics: %w", err)
	}
	defer rows.Close()

	var metrics []*analytics.EmployeeMetrics
	for rows.Next() {
		m := &analytics.EmployeeMetrics{}
		err := rows.Scan(
			&m.ID, &m.EmployeeID, &m.EmployeeName, &m.DepartmentID, &m.PositionID,
			&m.MetricDate, &m.TasksCompleted, &m.TasksAssigned,
			&m.AvgCompletionTimeHours, &m.OnTimeCompletionRate,
			&m.AvgTaskPriority, &m.SkillsUsed, &m.ProjectsInvolved,
			&m.EfficiencyScore, &m.CreatedAt, &m.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan employee metrics: %w", err)
		}
		metrics = append(metrics, m)
	}

	return metrics, totalCount, rows.Err()
}

func (s *Storage) GetTopPerformers(ctx context.Context, limit int32, departmentID string, startDate, endDate time.Time) ([]*analytics.EmployeeMetrics, error) {
	whereClause := "metric_date >= $1 AND metric_date <= $2"
	args := []any{startDate, endDate, limit}

	if departmentID != "" {
		args = []any{startDate, endDate, departmentID, limit}
		whereClause += " AND department_id = $3"
		query := `
			SELECT 
				id, employee_id, employee_name, department_id, position_id,
				metric_date, tasks_completed, tasks_assigned,
				avg_completion_time_hours, on_time_completion_rate,
				avg_task_priority, skills_used, projects_involved,
				efficiency_score, created_at, updated_at
			FROM analytics.employee_metrics
			WHERE ` + whereClause + `
			ORDER BY efficiency_score DESC, metric_date DESC
			LIMIT $4
		`
		return s.scanEmployeeMetricsList(ctx, query, args...)
	}

	query := `
		SELECT 
			id, employee_id, employee_name, department_id, position_id,
			metric_date, tasks_completed, tasks_assigned,
			avg_completion_time_hours, on_time_completion_rate,
			avg_task_priority, skills_used, projects_involved,
			efficiency_score, created_at, updated_at
		FROM analytics.employee_metrics
		WHERE ` + whereClause + `
		ORDER BY efficiency_score DESC, metric_date DESC
		LIMIT $3
	`
	return s.scanEmployeeMetricsList(ctx, query, args...)
}

func (s *Storage) scanEmployeeMetricsList(ctx context.Context, query string, args ...any) ([]*analytics.EmployeeMetrics, error) {
	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query: %w", err)
	}
	defer rows.Close()

	var metrics []*analytics.EmployeeMetrics
	for rows.Next() {
		m := &analytics.EmployeeMetrics{}
		err := rows.Scan(
			&m.ID, &m.EmployeeID, &m.EmployeeName, &m.DepartmentID, &m.PositionID,
			&m.MetricDate, &m.TasksCompleted, &m.TasksAssigned,
			&m.AvgCompletionTimeHours, &m.OnTimeCompletionRate,
			&m.AvgTaskPriority, &m.SkillsUsed, &m.ProjectsInvolved,
			&m.EfficiencyScore, &m.CreatedAt, &m.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan: %w", err)
		}
		metrics = append(metrics, m)
	}

	return metrics, rows.Err()
}

func (s *Storage) SaveProjectMetrics(ctx context.Context, metrics *analytics.ProjectMetrics) error {
	query := `
		INSERT INTO analytics.project_metrics (
			project_id, project_name, manager_id, manager_name,
			metric_date, total_tasks, completed_tasks,
			in_progress_tasks, overdue_tasks,
			completion_rate, on_time_completion_rate,
			team_size, avg_task_duration_hours,
			project_health_score
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
		)
		ON CONFLICT (project_id, metric_date) DO UPDATE SET
			project_name = EXCLUDED.project_name,
			manager_id = EXCLUDED.manager_id,
			manager_name = EXCLUDED.manager_name,
			total_tasks = EXCLUDED.total_tasks,
			completed_tasks = EXCLUDED.completed_tasks,
			in_progress_tasks = EXCLUDED.in_progress_tasks,
			overdue_tasks = EXCLUDED.overdue_tasks,
			completion_rate = EXCLUDED.completion_rate,
			on_time_completion_rate = EXCLUDED.on_time_completion_rate,
			team_size = EXCLUDED.team_size,
			avg_task_duration_hours = EXCLUDED.avg_task_duration_hours,
			project_health_score = EXCLUDED.project_health_score,
			updated_at = NOW()
		RETURNING id
	`

	var id string
	err := s.pool.QueryRow(ctx, query,
		metrics.ProjectID, metrics.ProjectName, metrics.ManagerID, metrics.ManagerName,
		metrics.MetricDate, metrics.TotalTasks, metrics.CompletedTasks,
		metrics.InProgressTasks, metrics.OverdueTasks,
		metrics.CompletionRate, metrics.OnTimeCompletionRate,
		metrics.TeamSize, metrics.AvgTaskDurationHours,
		metrics.ProjectHealthScore,
	).Scan(&id)

	if err != nil {
		return fmt.Errorf("failed to save project metrics: %w", err)
	}

	metrics.ID = id
	return nil
}

func (s *Storage) GetProjectMetrics(ctx context.Context, projectID string, startDate, endDate time.Time) ([]*analytics.ProjectMetrics, error) {
	query := `
		SELECT 
			id, project_id, project_name, manager_id, manager_name,
			metric_date, total_tasks, completed_tasks,
			in_progress_tasks, overdue_tasks,
			completion_rate, on_time_completion_rate,
			team_size, avg_task_duration_hours,
			project_health_score, created_at, updated_at
		FROM analytics.project_metrics
		WHERE project_id = $1
			AND metric_date >= $2
			AND metric_date <= $3
		ORDER BY metric_date DESC
	`

	rows, err := s.pool.Query(ctx, query, projectID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to query project metrics: %w", err)
	}
	defer rows.Close()

	var metrics []*analytics.ProjectMetrics
	for rows.Next() {
		m := &analytics.ProjectMetrics{}
		err := rows.Scan(
			&m.ID, &m.ProjectID, &m.ProjectName, &m.ManagerID, &m.ManagerName,
			&m.MetricDate, &m.TotalTasks, &m.CompletedTasks,
			&m.InProgressTasks, &m.OverdueTasks,
			&m.CompletionRate, &m.OnTimeCompletionRate,
			&m.TeamSize, &m.AvgTaskDurationHours,
			&m.ProjectHealthScore, &m.CreatedAt, &m.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan project metrics: %w", err)
		}
		metrics = append(metrics, m)
	}

	return metrics, rows.Err()
}

func (s *Storage) ListProjectMetrics(ctx context.Context, pageSize, pageNumber int32, managerID string, startDate, endDate time.Time) ([]*analytics.ProjectMetrics, int32, error) {
	whereClause := "metric_date >= $1 AND metric_date <= $2"
	args := []interface{}{startDate, endDate}

	if managerID != "" {
		args = append(args, managerID)
		whereClause += fmt.Sprintf(" AND manager_id = $%d", len(args))
	}

	countQuery := "SELECT COUNT(*) FROM analytics.project_metrics WHERE " + whereClause
	var totalCount int32
	err := s.pool.QueryRow(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count project metrics: %w", err)
	}

	offset := (pageNumber - 1) * pageSize
	args = append(args, pageSize, offset)
	query := `
		SELECT 
			id, project_id, project_name, manager_id, manager_name,
			metric_date, total_tasks, completed_tasks,
			in_progress_tasks, overdue_tasks,
			completion_rate, on_time_completion_rate,
			team_size, avg_task_duration_hours,
			project_health_score, created_at, updated_at
		FROM analytics.project_metrics
		WHERE ` + whereClause + `
		ORDER BY metric_date DESC, project_health_score DESC
		LIMIT $` + fmt.Sprintf("%d", len(args)-1) + ` OFFSET $` + fmt.Sprintf("%d", len(args)) + ``

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query project metrics: %w", err)
	}
	defer rows.Close()

	var metrics []*analytics.ProjectMetrics
	for rows.Next() {
		m := &analytics.ProjectMetrics{}
		err := rows.Scan(
			&m.ID, &m.ProjectID, &m.ProjectName, &m.ManagerID, &m.ManagerName,
			&m.MetricDate, &m.TotalTasks, &m.CompletedTasks,
			&m.InProgressTasks, &m.OverdueTasks,
			&m.CompletionRate, &m.OnTimeCompletionRate,
			&m.TeamSize, &m.AvgTaskDurationHours,
			&m.ProjectHealthScore, &m.CreatedAt, &m.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan project metrics: %w", err)
		}
		metrics = append(metrics, m)
	}

	return metrics, totalCount, rows.Err()
}

func (s *Storage) SaveDepartmentMetrics(ctx context.Context, metrics *analytics.DepartmentMetrics) error {
	query := `
		INSERT INTO analytics.department_metrics (
			department_id, department_name, metric_date,
			total_employees, active_projects, total_tasks, completed_tasks,
			avg_employee_efficiency, department_completion_rate,
			department_on_time_rate, department_health_score
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
		)
		ON CONFLICT (department_id, metric_date) DO UPDATE SET
			department_name = EXCLUDED.department_name,
			total_employees = EXCLUDED.total_employees,
			active_projects = EXCLUDED.active_projects,
			total_tasks = EXCLUDED.total_tasks,
			completed_tasks = EXCLUDED.completed_tasks,
			avg_employee_efficiency = EXCLUDED.avg_employee_efficiency,
			department_completion_rate = EXCLUDED.department_completion_rate,
			department_on_time_rate = EXCLUDED.department_on_time_rate,
			department_health_score = EXCLUDED.department_health_score,
			updated_at = NOW()
		RETURNING id
	`

	var id string
	err := s.pool.QueryRow(ctx, query,
		metrics.DepartmentID, metrics.DepartmentName, metrics.MetricDate,
		metrics.TotalEmployees, metrics.ActiveProjects, metrics.TotalTasks, metrics.CompletedTasks,
		metrics.AvgEmployeeEfficiency, metrics.DepartmentCompletionRate,
		metrics.DepartmentOnTimeRate, metrics.DepartmentHealthScore,
	).Scan(&id)

	if err != nil {
		return fmt.Errorf("failed to save department metrics: %w", err)
	}

	metrics.ID = id
	return nil
}

func (s *Storage) GetDepartmentMetrics(ctx context.Context, departmentID string, startDate, endDate time.Time) ([]*analytics.DepartmentMetrics, error) {
	query := `
		SELECT 
			id, department_id, department_name, metric_date,
			total_employees, active_projects, total_tasks, completed_tasks,
			avg_employee_efficiency, department_completion_rate,
			department_on_time_rate, department_health_score,
			created_at, updated_at
		FROM analytics.department_metrics
		WHERE department_id = $1
			AND metric_date >= $2
			AND metric_date <= $3
		ORDER BY metric_date DESC
	`

	rows, err := s.pool.Query(ctx, query, departmentID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to query department metrics: %w", err)
	}
	defer rows.Close()

	var metrics []*analytics.DepartmentMetrics
	for rows.Next() {
		m := &analytics.DepartmentMetrics{}
		err := rows.Scan(
			&m.ID, &m.DepartmentID, &m.DepartmentName, &m.MetricDate,
			&m.TotalEmployees, &m.ActiveProjects, &m.TotalTasks, &m.CompletedTasks,
			&m.AvgEmployeeEfficiency, &m.DepartmentCompletionRate,
			&m.DepartmentOnTimeRate, &m.DepartmentHealthScore,
			&m.CreatedAt, &m.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan department metrics: %w", err)
		}
		metrics = append(metrics, m)
	}

	return metrics, rows.Err()
}

func (s *Storage) ListDepartmentMetrics(ctx context.Context, pageSize, pageNumber int32, startDate, endDate time.Time) ([]*analytics.DepartmentMetrics, int32, error) {
	whereClause := "metric_date >= $1 AND metric_date <= $2"
	args := []interface{}{startDate, endDate}

	countQuery := "SELECT COUNT(*) FROM analytics.department_metrics WHERE " + whereClause
	var totalCount int32
	err := s.pool.QueryRow(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count department metrics: %w", err)
	}

	offset := (pageNumber - 1) * pageSize
	args = append(args, pageSize, offset)
	query := `
		SELECT 
			id, department_id, department_name, metric_date,
			total_employees, active_projects, total_tasks, completed_tasks,
			avg_employee_efficiency, department_completion_rate,
			department_on_time_rate, department_health_score,
			created_at, updated_at
		FROM analytics.department_metrics
		WHERE ` + whereClause + `
		ORDER BY metric_date DESC, department_health_score DESC
		LIMIT $` + fmt.Sprintf("%d", len(args)-1) + ` OFFSET $` + fmt.Sprintf("%d", len(args)) + ``

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query department metrics: %w", err)
	}
	defer rows.Close()

	var metrics []*analytics.DepartmentMetrics
	for rows.Next() {
		m := &analytics.DepartmentMetrics{}
		err := rows.Scan(
			&m.ID, &m.DepartmentID, &m.DepartmentName, &m.MetricDate,
			&m.TotalEmployees, &m.ActiveProjects, &m.TotalTasks, &m.CompletedTasks,
			&m.AvgEmployeeEfficiency, &m.DepartmentCompletionRate,
			&m.DepartmentOnTimeRate, &m.DepartmentHealthScore,
			&m.CreatedAt, &m.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan department metrics: %w", err)
		}
		metrics = append(metrics, m)
	}

	return metrics, totalCount, rows.Err()
}

func (s *Storage) SaveDailySnapshot(ctx context.Context, snapshot *analytics.DailySnapshot) error {
	query := `
		INSERT INTO analytics.daily_snapshots (
			snapshot_date, total_employees, active_employees,
			total_projects, active_projects, total_tasks,
			completed_tasks, overdue_tasks,
			avg_company_efficiency, avg_on_time_rate
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		)
		ON CONFLICT (snapshot_date) DO UPDATE SET
			total_employees = EXCLUDED.total_employees,
			active_employees = EXCLUDED.active_employees,
			total_projects = EXCLUDED.total_projects,
			active_projects = EXCLUDED.active_projects,
			total_tasks = EXCLUDED.total_tasks,
			completed_tasks = EXCLUDED.completed_tasks,
			overdue_tasks = EXCLUDED.overdue_tasks,
			avg_company_efficiency = EXCLUDED.avg_company_efficiency,
			avg_on_time_rate = EXCLUDED.avg_on_time_rate,
			updated_at = NOW()
		RETURNING id
	`

	var id string
	err := s.pool.QueryRow(ctx, query,
		snapshot.SnapshotDate, snapshot.TotalEmployees, snapshot.ActiveEmployees,
		snapshot.TotalProjects, snapshot.ActiveProjects, snapshot.TotalTasks,
		snapshot.CompletedTasks, snapshot.OverdueTasks,
		snapshot.AvgCompanyEfficiency, snapshot.AvgOnTimeRate,
	).Scan(&id)

	if err != nil {
		return fmt.Errorf("failed to save daily snapshot: %w", err)
	}

	snapshot.ID = id
	return nil
}

func (s *Storage) GetDailySnapshot(ctx context.Context, date time.Time) (*analytics.DailySnapshot, error) {
	query := `
		SELECT 
			id, snapshot_date, total_employees, active_employees,
			total_projects, active_projects, total_tasks,
			completed_tasks, overdue_tasks,
			avg_company_efficiency, avg_on_time_rate,
			created_at, updated_at
		FROM analytics.daily_snapshots
		WHERE snapshot_date = $1
	`

	snapshot := &analytics.DailySnapshot{}
	err := s.pool.QueryRow(ctx, query, date).Scan(
		&snapshot.ID, &snapshot.SnapshotDate, &snapshot.TotalEmployees, &snapshot.ActiveEmployees,
		&snapshot.TotalProjects, &snapshot.ActiveProjects, &snapshot.TotalTasks,
		&snapshot.CompletedTasks, &snapshot.OverdueTasks,
		&snapshot.AvgCompanyEfficiency, &snapshot.AvgOnTimeRate,
		&snapshot.CreatedAt, &snapshot.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query daily snapshot: %w", err)
	}

	return snapshot, nil
}

func (s *Storage) GetDailySnapshots(ctx context.Context, startDate, endDate time.Time) ([]*analytics.DailySnapshot, error) {
	query := `
		SELECT 
			id, snapshot_date, total_employees, active_employees,
			total_projects, active_projects, total_tasks,
			completed_tasks, overdue_tasks,
			avg_company_efficiency, avg_on_time_rate,
			created_at, updated_at
		FROM analytics.daily_snapshots
		WHERE snapshot_date >= $1 AND snapshot_date <= $2
		ORDER BY snapshot_date DESC
	`

	rows, err := s.pool.Query(ctx, query, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to query daily snapshots: %w", err)
	}
	defer rows.Close()

	var snapshots []*analytics.DailySnapshot
	for rows.Next() {
		snapshot := &analytics.DailySnapshot{}
		err := rows.Scan(
			&snapshot.ID, &snapshot.SnapshotDate, &snapshot.TotalEmployees, &snapshot.ActiveEmployees,
			&snapshot.TotalProjects, &snapshot.ActiveProjects, &snapshot.TotalTasks,
			&snapshot.CompletedTasks, &snapshot.OverdueTasks,
			&snapshot.AvgCompanyEfficiency, &snapshot.AvgOnTimeRate,
			&snapshot.CreatedAt, &snapshot.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan daily snapshot: %w", err)
		}
		snapshots = append(snapshots, snapshot)
	}

	return snapshots, rows.Err()
}
