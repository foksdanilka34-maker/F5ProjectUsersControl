package repo

import (
	"context"
	"time"

	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/business/dto"
)

type AnalyticsRepo struct {
	db DBExecutor
}

func NewAnalyticsRepo(db DBExecutor) *AnalyticsRepo {
	return &AnalyticsRepo{db: db}
}

func (r *AnalyticsRepo) GetSummary(ctx context.Context) (*dto.DashboardStatsDTO, error) {
	query := `
		SELECT 
			(SELECT COUNT(*) FROM business.projects) as total_projects,
			(SELECT COUNT(*) FROM business.projects WHERE UPPER(status) = 'ACTIVE') as active_projects,
			(SELECT COUNT(*) FROM business.tasks) as total_tasks,
			(SELECT COUNT(*) FROM business.tasks WHERE UPPER(status) = 'DONE') as completed_tasks,
			(SELECT COUNT(*) FROM business.tasks WHERE UPPER(status) = 'DONE' AND (due_date IS NULL OR completed_at <= due_date)) as completed_on_time,
			(SELECT COUNT(*) FROM business.tasks WHERE UPPER(status) = 'DONE' AND due_date IS NOT NULL AND completed_at > due_date) as completed_late,
			(SELECT COUNT(*) FROM business.tasks WHERE due_date IS NOT NULL AND due_date < NOW() AND UPPER(status) != 'DONE') as overdue_tasks,
			(SELECT COUNT(DISTINCT user_id) FROM business.project_members) as total_employees,
			(SELECT COUNT(DISTINCT user_id) FROM business.project_members) as active_employees
	`
	var s dto.DashboardStatsDTO
	err := r.db.QueryRow(ctx, query).Scan(
		&s.TotalProjects, &s.ActiveProjects, &s.TotalTasks, &s.CompletedTasks,
		&s.CompletedOnTime, &s.CompletedLate, &s.OverdueTasks,
		&s.TotalEmployees, &s.ActiveEmployees,
	)
	if err != nil {
		return nil, err
	}

	if s.TotalTasks > 0 {
		s.AvgCompletionRate = float64(s.CompletedTasks) / float64(s.TotalTasks) * 100
	}
	if s.CompletedTasks > 0 {
		s.AvgOnTimeRate = float64(s.CompletedOnTime) / float64(s.CompletedTasks) * 100
	} else {
		s.AvgOnTimeRate = 100
	}

	return &s, nil
}

func (r *AnalyticsRepo) GetProjectAnalytics(ctx context.Context, projectID int64) (*dto.ProjectMetricsDTO, error) {
	query := `
		SELECT 
			p.id,
			p.owner_id,
			COUNT(t.id) as total_tasks,
			COUNT(t.id) FILTER (WHERE UPPER(t.status) = 'DONE') as completed_tasks,
			COUNT(t.id) FILTER (WHERE UPPER(t.status) = 'DONE' AND (t.due_date IS NULL OR t.completed_at <= t.due_date)) as completed_on_time,
			COUNT(t.id) FILTER (WHERE UPPER(t.status) = 'DONE' AND t.due_date IS NOT NULL AND t.completed_at > t.due_date) as completed_late,
			COUNT(t.id) FILTER (WHERE UPPER(t.status) = 'IN_PROGRESS') as in_progress_tasks,
			COUNT(t.id) FILTER (WHERE t.due_date IS NOT NULL AND t.due_date < NOW() AND UPPER(t.status) != 'DONE') as overdue_tasks,
			(SELECT COUNT(*) FROM business.project_members WHERE project_id = p.id) as team_size
		FROM business.projects p
		LEFT JOIN business.tasks t ON p.id = t.project_id
		WHERE p.id = $1
		GROUP BY p.id, p.owner_id
	`
	var m dto.ProjectMetricsDTO
	err := r.db.QueryRow(ctx, query, projectID).Scan(
		&m.ProjectID, &m.ManagerID, &m.TotalTasks, &m.CompletedTasks,
		&m.CompletedOnTime, &m.CompletedLate, &m.InProgressTasks, &m.OverdueTasks,
		&m.TeamSize,
	)
	if err != nil {
		return nil, err
	}

	if m.TotalTasks > 0 {
		m.ProgressPercent = float64(m.CompletedTasks) / float64(m.TotalTasks) * 100
	}
	if m.CompletedTasks > 0 {
		m.OnTimeRate = float64(m.CompletedOnTime) / float64(m.CompletedTasks) * 100
	} else {
		m.OnTimeRate = 100
	}

	m.HealthStatus = "HEALTH_STATUS_HEALTHY"
	if m.OverdueTasks > m.TotalTasks/2 {
		m.HealthStatus = "HEALTH_STATUS_CRITICAL"
	} else if m.OverdueTasks > 0 {
		m.HealthStatus = "HEALTH_STATUS_AT_RISK"
	}
	m.CalculatedAt = time.Now()

	return &m, nil
}

func (r *AnalyticsRepo) GetAllProjectAnalytics(ctx context.Context) ([]dto.ProjectMetricsDTO, error) {
	query := `
		SELECT 
			p.id,
			p.owner_id,
			COUNT(t.id) as total_tasks,
			COUNT(t.id) FILTER (WHERE UPPER(t.status) = 'DONE') as completed_tasks,
			COUNT(t.id) FILTER (WHERE UPPER(t.status) = 'DONE' AND (t.due_date IS NULL OR t.completed_at <= t.due_date)) as completed_on_time,
			COUNT(t.id) FILTER (WHERE UPPER(t.status) = 'DONE' AND t.due_date IS NOT NULL AND t.completed_at > t.due_date) as completed_late,
			COUNT(t.id) FILTER (WHERE UPPER(t.status) = 'IN_PROGRESS') as in_progress_tasks,
			COUNT(t.id) FILTER (WHERE t.due_date IS NOT NULL AND t.due_date < NOW() AND UPPER(t.status) != 'DONE') as overdue_tasks,
			(SELECT COUNT(*) FROM business.project_members WHERE project_id = p.id) as team_size
		FROM business.projects p
		LEFT JOIN business.tasks t ON p.id = t.project_id
		GROUP BY p.id, p.owner_id
		ORDER BY p.id ASC
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []dto.ProjectMetricsDTO
	for rows.Next() {
		var m dto.ProjectMetricsDTO
		if err := rows.Scan(
			&m.ProjectID, &m.ManagerID, &m.TotalTasks, &m.CompletedTasks,
			&m.CompletedOnTime, &m.CompletedLate, &m.InProgressTasks, &m.OverdueTasks,
			&m.TeamSize,
		); err != nil {
			return nil, err
		}

		if m.TotalTasks > 0 {
			m.ProgressPercent = float64(m.CompletedTasks) / float64(m.TotalTasks) * 100
		}
		if m.CompletedTasks > 0 {
			m.OnTimeRate = float64(m.CompletedOnTime) / float64(m.CompletedTasks) * 100
		} else {
			m.OnTimeRate = 100
		}

		m.HealthStatus = "HEALTH_STATUS_HEALTHY"
		if m.OverdueTasks > m.TotalTasks/2 {
			m.HealthStatus = "HEALTH_STATUS_CRITICAL"
		} else if m.OverdueTasks > 0 {
			m.HealthStatus = "HEALTH_STATUS_AT_RISK"
		}
		m.CalculatedAt = time.Now()

		list = append(list, m)
	}

	return list, nil
}

func (r *AnalyticsRepo) GetEmployeeAnalytics(ctx context.Context, userID int64) (*dto.EmployeeMetricsDTO, error) {
	query := `
		WITH priority_weights AS (
			SELECT 
				t.id,
				UPPER(t.status) as status,
				t.due_date,
				t.completed_at,
				CASE LOWER(t.priority)
					WHEN 'critical' THEN 4
					WHEN 'high' THEN 3
					WHEN 'medium' THEN 2
					WHEN 'low' THEN 1
					ELSE 2
				END as weight
			FROM business.tasks t
			WHERE t.assignee_id = $1
		)
		SELECT 
			COUNT(id) as assigned_tasks,
			COUNT(id) FILTER (WHERE status = 'DONE') as completed_tasks,
			COUNT(id) FILTER (WHERE status = 'DONE' AND (due_date IS NULL OR completed_at <= due_date)) as completed_on_time,
			COUNT(id) FILTER (WHERE status = 'DONE' AND due_date IS NOT NULL AND completed_at > due_date) as completed_late,
			COUNT(id) FILTER (WHERE status = 'IN_PROGRESS') as in_progress_tasks,
			COUNT(id) FILTER (WHERE due_date IS NOT NULL AND due_date < NOW() AND status != 'DONE') as overdue_tasks,
			COALESCE(SUM(weight) FILTER (WHERE status = 'DONE' AND (due_date IS NULL OR completed_at <= due_date)), 0) as weighted_on_time,
			COALESCE(SUM(weight) FILTER (WHERE status = 'DONE'), 0) as weighted_total
		FROM priority_weights
	`
	var m dto.EmployeeMetricsDTO
	m.EmployeeID = userID
	var weightedOnTime, weightedTotal float64

	err := r.db.QueryRow(ctx, query, userID).Scan(
		&m.AssignedTasks, &m.CompletedTasks, &m.CompletedOnTime, &m.CompletedLate,
		&m.InProgressTasks, &m.OverdueTasks, &weightedOnTime, &weightedTotal,
	)
	if err != nil {
		return nil, err
	}

	var completionRate float64
	if m.AssignedTasks > 0 {
		completionRate = float64(m.CompletedTasks) / float64(m.AssignedTasks) * 100
	}

	var weightedOnTimeRate float64
	if weightedTotal > 0 {
		weightedOnTimeRate = weightedOnTime / weightedTotal * 100
	}

	if m.CompletedTasks > 0 && m.AssignedTasks > 0 {
		m.CompletionRate = completionRate * weightedOnTimeRate / 100
	} else {
		m.CompletionRate = 0
	}
	m.OnTimeRate = weightedOnTimeRate

	return &m, nil
}

func (r *AnalyticsRepo) GetAllEmployeeAnalytics(ctx context.Context) ([]dto.EmployeeMetricsDTO, error) {
	query := `
		WITH emp AS (
			SELECT DISTINCT user_id FROM business.project_members
		),
		priority_weights AS (
			SELECT
				t.assignee_id,
				t.id,
				UPPER(t.status) as status,
				t.due_date,
				t.completed_at,
				CASE LOWER(t.priority)
					WHEN 'critical' THEN 4
					WHEN 'high' THEN 3
					WHEN 'medium' THEN 2
					WHEN 'low' THEN 1
					ELSE 2
				END as weight
			FROM business.tasks t
			WHERE t.assignee_id IS NOT NULL
		)
		SELECT
			e.user_id,
			COALESCE(COUNT(pw.id), 0) as assigned_tasks,
			COALESCE(COUNT(pw.id) FILTER (WHERE pw.status = 'DONE'), 0) as completed_tasks,
			COALESCE(COUNT(pw.id) FILTER (WHERE pw.status = 'DONE' AND (pw.due_date IS NULL OR pw.completed_at <= pw.due_date)), 0) as completed_on_time,
			COALESCE(COUNT(pw.id) FILTER (WHERE pw.status = 'DONE' AND pw.due_date IS NOT NULL AND pw.completed_at > pw.due_date), 0) as completed_late,
			COALESCE(COUNT(pw.id) FILTER (WHERE pw.status = 'IN_PROGRESS'), 0) as in_progress_tasks,
			COALESCE(COUNT(pw.id) FILTER (WHERE pw.due_date IS NOT NULL AND pw.due_date < NOW() AND pw.status != 'DONE'), 0) as overdue_tasks,
			COALESCE(SUM(pw.weight) FILTER (WHERE pw.status = 'DONE' AND (pw.due_date IS NULL OR pw.completed_at <= pw.due_date)), 0) as weighted_on_time,
			COALESCE(SUM(pw.weight) FILTER (WHERE pw.status = 'DONE'), 0) as weighted_total
		FROM emp e
		LEFT JOIN priority_weights pw ON pw.assignee_id = e.user_id
		GROUP BY e.user_id
		ORDER BY e.user_id ASC
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []dto.EmployeeMetricsDTO
	for rows.Next() {
		var m dto.EmployeeMetricsDTO
		var weightedOnTime, weightedTotal float64

		if err := rows.Scan(
			&m.EmployeeID, &m.AssignedTasks, &m.CompletedTasks, &m.CompletedOnTime, &m.CompletedLate,
			&m.InProgressTasks, &m.OverdueTasks, &weightedOnTime, &weightedTotal,
		); err != nil {
			return nil, err
		}

		var completionRate float64
		if m.AssignedTasks > 0 {
			completionRate = float64(m.CompletedTasks) / float64(m.AssignedTasks) * 100
		}

		var weightedOnTimeRate float64
		if weightedTotal > 0 {
			weightedOnTimeRate = weightedOnTime / weightedTotal * 100
		}

		if m.CompletedTasks > 0 && m.AssignedTasks > 0 {
			m.CompletionRate = completionRate * weightedOnTimeRate / 100
		} else {
			m.CompletionRate = 0
		}
		m.OnTimeRate = weightedOnTimeRate

		list = append(list, m)
	}

	return list, nil
}

func (r *AnalyticsRepo) GetProductivityTrends(ctx context.Context, days int) ([]dto.ProductivityTrendEntryDTO, error) {
	query := `
		WITH dates AS (
			SELECT generate_series(
				CURRENT_DATE - make_interval(days => $1),
				CURRENT_DATE,
				'1 day'::interval
			)::date as day
		),
		daily AS (
			SELECT
				DATE(t.completed_at) as day,
				COUNT(*) as completed,
				COUNT(*) FILTER (WHERE t.due_date IS NULL OR t.completed_at <= t.due_date) as on_time,
				COUNT(*) as total_completed
			FROM business.tasks t
			WHERE UPPER(t.status) = 'DONE'
				AND t.completed_at IS NOT NULL
				AND t.completed_at >= CURRENT_DATE - make_interval(days => $1)
			GROUP BY DATE(t.completed_at)
		)
		SELECT
			d.day,
			COALESCE(dl.completed, 0) as tasks_completed,
			CASE WHEN COALESCE(dl.total_completed, 0) > 0
				THEN (dl.on_time::float / dl.total_completed * 100)
				ELSE 0
			END as avg_completion_rate
		FROM dates d
		LEFT JOIN daily dl ON d.day = dl.day
		ORDER BY d.day ASC
	`
	rows, err := r.db.Query(ctx, query, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []dto.ProductivityTrendEntryDTO
	for rows.Next() {
		var day time.Time
		var completed int
		var rate float64
		if err := rows.Scan(&day, &completed, &rate); err != nil {
			return nil, err
		}
		list = append(list, dto.ProductivityTrendEntryDTO{
			Date:              day.Format("2006-01-02"),
			TasksCompleted:    completed,
			AvgCompletionRate: rate,
		})
	}

	return list, nil
}
