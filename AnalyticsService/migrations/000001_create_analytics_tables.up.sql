CREATE SCHEMA IF NOT EXISTS analytics;

CREATE TABLE IF NOT EXISTS analytics.employee_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id UUID NOT NULL,
    
    metric_date DATE NOT NULL,
    
    assigned_tasks INT NOT NULL DEFAULT 0,
    completed_tasks INT NOT NULL DEFAULT 0,
    in_progress_tasks INT NOT NULL DEFAULT 0,
    overdue_tasks INT NOT NULL DEFAULT 0,
    efficiency_score FLOAT NOT NULL DEFAULT 0,
    task_completion_rate FLOAT NOT NULL DEFAULT 0,
    on_time_completion_rate FLOAT NOT NULL DEFAULT 0,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    UNIQUE(employee_id, metric_date)
);

CREATE TABLE IF NOT EXISTS analytics.project_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL,
    manager_id UUID,
    
    metric_date DATE NOT NULL,
    
    total_tasks INT NOT NULL DEFAULT 0,
    completed_tasks INT NOT NULL DEFAULT 0,
    in_progress_tasks INT NOT NULL DEFAULT 0,
    overdue_tasks INT NOT NULL DEFAULT 0,
    
    delivery_performance FLOAT NOT NULL DEFAULT 0,
    schedule_performance FLOAT NOT NULL DEFAULT 0,
    quality_performance FLOAT NOT NULL DEFAULT 0,
    team_performance FLOAT NOT NULL DEFAULT 0,
    
    health_index FLOAT NOT NULL DEFAULT 0,
    risk_score FLOAT NOT NULL DEFAULT 0,
    health_status VARCHAR(20),
    
    velocity FLOAT NOT NULL DEFAULT 0,
    projected_end_date DATE,
    team_capacity_utilization FLOAT NOT NULL DEFAULT 0,
    team_size INT NOT NULL DEFAULT 0,
    avg_team_efficiency FLOAT NOT NULL DEFAULT 0,
    
    is_at_risk BOOLEAN DEFAULT false,
    days_until_due INT,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    UNIQUE(project_id, metric_date)
);

CREATE INDEX idx_employee_metrics_employee_date 
    ON analytics.employee_metrics(employee_id, metric_date DESC);

CREATE INDEX idx_employee_metrics_efficiency 
    ON analytics.employee_metrics(efficiency_score DESC);

CREATE INDEX idx_project_metrics_project_date 
    ON analytics.project_metrics(project_id, metric_date DESC);

CREATE INDEX idx_project_metrics_health 
    ON analytics.project_metrics(health_index DESC);
