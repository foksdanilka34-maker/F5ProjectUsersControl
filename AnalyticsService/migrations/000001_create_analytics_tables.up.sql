CREATE SCHEMA IF NOT EXISTS analytics;

CREATE TABLE IF NOT EXISTS analytics.employee_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id UUID NOT NULL,
    
    metric_date TIMESTAMPTZ NOT NULL,
    
    assigned_tasks INT NOT NULL DEFAULT 0,
    completed_tasks INT NOT NULL DEFAULT 0,
    in_progress_tasks INT NOT NULL DEFAULT 0,
    overdue_tasks INT NOT NULL DEFAULT 0,
    
    on_time_completed_tasks INT NOT NULL DEFAULT 0,
    total_task_duration_seconds BIGINT NOT NULL DEFAULT 0,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    UNIQUE(employee_id, metric_date)
);

CREATE TABLE IF NOT EXISTS analytics.project_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL,
    manager_id UUID,
    
    metric_date TIMESTAMPTZ NOT NULL,
    
    total_tasks INT NOT NULL DEFAULT 0,
    completed_tasks INT NOT NULL DEFAULT 0,
    in_progress_tasks INT NOT NULL DEFAULT 0,
    overdue_tasks INT NOT NULL DEFAULT 0,
    on_time_completed_tasks INT NOT NULL DEFAULT 0,
    
    team_size INT NOT NULL DEFAULT 0,
    total_task_duration_seconds BIGINT NOT NULL DEFAULT 0,
    total_priority_weight_completed INT NOT NULL DEFAULT 0,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    UNIQUE(project_id, metric_date)
);