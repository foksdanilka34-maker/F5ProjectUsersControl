ALTER TABLE project.tasks
ADD COLUMN started_at TIMESTAMPTZ,
ADD COLUMN completed_at TIMESTAMPTZ;

ALTER TABLE project.tasks
ADD CONSTRAINT check_task_completion_after_start
CHECK (completed_at IS NULL OR started_at IS NULL OR completed_at >= started_at);

CREATE TABLE IF NOT EXISTS project.task_status_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id UUID NOT NULL REFERENCES project.tasks(task_id) ON DELETE CASCADE,
    from_status project.task_status,
    to_status project.task_status NOT NULL,
    changed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    actor_id UUID,
    
    CONSTRAINT check_status_change CHECK (from_status IS NULL OR from_status != to_status)
);

CREATE INDEX IF NOT EXISTS idx_task_status_history_task_id 
ON project.task_status_history (task_id, changed_at DESC);

CREATE INDEX IF NOT EXISTS idx_task_status_history_changed_at 
ON project.task_status_history (changed_at DESC);

CREATE TABLE IF NOT EXISTS project.project_metrics (
    project_id UUID NOT NULL REFERENCES project.projects(project_id) ON DELETE CASCADE,
    metric_date DATE NOT NULL DEFAULT CURRENT_DATE,
    
    total_tasks INT NOT NULL DEFAULT 0,
    completed_tasks INT NOT NULL DEFAULT 0,
    overdue_tasks INT NOT NULL DEFAULT 0,
    in_progress_tasks INT NOT NULL DEFAULT 0,
    
    avg_completion_time_hours NUMERIC(10, 2),
    on_time_completion_rate NUMERIC(5, 2),
    
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    PRIMARY KEY (project_id, metric_date)
);

CREATE INDEX IF NOT EXISTS idx_project_metrics_date 
ON project.project_metrics (metric_date DESC);

UPDATE project.tasks
SET completed_at = updated_at
WHERE task_status = 'DONE' AND completed_at IS NULL;

UPDATE project.tasks
SET started_at = created_at
WHERE task_status IN ('IN_PROGRESS', 'REVIEW', 'DONE') AND started_at IS NULL;
