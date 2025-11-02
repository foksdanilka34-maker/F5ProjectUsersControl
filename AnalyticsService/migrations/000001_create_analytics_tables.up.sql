CREATE SCHEMA IF NOT EXISTS analytics;

CREATE TABLE IF NOT EXISTS analytics.employee_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id UUID NOT NULL,
    employee_name VARCHAR(255),
    department_id UUID,
    position_id UUID,
    
    metric_date DATE NOT NULL,
    
    tasks_completed INT NOT NULL DEFAULT 0,
    tasks_assigned INT NOT NULL DEFAULT 0,
    avg_completion_time_hours FLOAT NOT NULL DEFAULT 0,
    on_time_completion_rate FLOAT NOT NULL DEFAULT 0,
    avg_task_priority FLOAT NOT NULL DEFAULT 0,
    
    skills_used TEXT[],
    projects_involved INT NOT NULL DEFAULT 0,
    
    efficiency_score FLOAT NOT NULL DEFAULT 0,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    UNIQUE(employee_id, metric_date)
);

CREATE TABLE IF NOT EXISTS analytics.project_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL,
    project_name VARCHAR(255),
    manager_id UUID,
    manager_name VARCHAR(255),
    
    metric_date DATE NOT NULL,
    
    total_tasks INT NOT NULL DEFAULT 0,
    completed_tasks INT NOT NULL DEFAULT 0,
    in_progress_tasks INT NOT NULL DEFAULT 0,
    overdue_tasks INT NOT NULL DEFAULT 0,
    
    completion_rate FLOAT NOT NULL DEFAULT 0, 
    on_time_completion_rate FLOAT NOT NULL DEFAULT 0, 
    
    team_size INT NOT NULL DEFAULT 0,
    avg_task_duration_hours FLOAT NOT NULL DEFAULT 0,
    
    project_health_score FLOAT NOT NULL DEFAULT 0, 
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    UNIQUE(project_id, metric_date)
);

CREATE TABLE IF NOT EXISTS analytics.department_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    department_id UUID NOT NULL,
    department_name VARCHAR(255),
    
    metric_date DATE NOT NULL,
    
    total_employees INT NOT NULL DEFAULT 0,
    active_projects INT NOT NULL DEFAULT 0,
    total_tasks INT NOT NULL DEFAULT 0,
    completed_tasks INT NOT NULL DEFAULT 0,
    
    avg_employee_efficiency FLOAT NOT NULL DEFAULT 0,
    department_completion_rate FLOAT NOT NULL DEFAULT 0, 
    department_on_time_rate FLOAT NOT NULL DEFAULT 0, 
    
    department_health_score FLOAT NOT NULL DEFAULT 0, 
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    UNIQUE(department_id, metric_date)
);

CREATE TABLE IF NOT EXISTS analytics.daily_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    snapshot_date DATE NOT NULL UNIQUE,
    
    total_employees INT NOT NULL DEFAULT 0,
    active_employees INT NOT NULL DEFAULT 0,
    
    total_projects INT NOT NULL DEFAULT 0,
    active_projects INT NOT NULL DEFAULT 0,
    
    total_tasks INT NOT NULL DEFAULT 0,
    completed_tasks INT NOT NULL DEFAULT 0,
    overdue_tasks INT NOT NULL DEFAULT 0,
    
    avg_company_efficiency FLOAT NOT NULL DEFAULT 0,
    avg_on_time_rate FLOAT NOT NULL DEFAULT 0,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS analytics.metrics_cache (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cache_key VARCHAR(500) NOT NULL UNIQUE,
    cache_value BYTEA NOT NULL,
    ttl_seconds INT NOT NULL DEFAULT 3600,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL DEFAULT (NOW() + INTERVAL '1 hour')
);

CREATE INDEX idx_employee_metrics_employee_date 
    ON analytics.employee_metrics(employee_id, metric_date DESC);

CREATE INDEX idx_employee_metrics_date 
    ON analytics.employee_metrics(metric_date DESC);

CREATE INDEX idx_employee_metrics_efficiency 
    ON analytics.employee_metrics(efficiency_score DESC, metric_date DESC);

CREATE INDEX idx_project_metrics_project_date 
    ON analytics.project_metrics(project_id, metric_date DESC);

CREATE INDEX idx_project_metrics_date 
    ON analytics.project_metrics(metric_date DESC);

CREATE INDEX idx_project_metrics_health 
    ON analytics.project_metrics(project_health_score DESC, metric_date DESC);

CREATE INDEX idx_department_metrics_department_date 
    ON analytics.department_metrics(department_id, metric_date DESC);

CREATE INDEX idx_department_metrics_date 
    ON analytics.department_metrics(metric_date DESC);

CREATE INDEX idx_daily_snapshots_date 
    ON analytics.daily_snapshots(snapshot_date DESC);

CREATE INDEX idx_metrics_cache_expires_at 
    ON analytics.metrics_cache(expires_at);
