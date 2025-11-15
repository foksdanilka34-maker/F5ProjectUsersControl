DROP SCHEMA IF EXISTS analytics CASCADE;

ALTER TABLE analytics.employee_metrics ADD COLUMN IF NOT EXISTS efficiency_score FLOAT NOT NULL DEFAULT 0;
ALTER TABLE analytics.employee_metrics ADD COLUMN IF NOT EXISTS task_completion_rate FLOAT NOT NULL DEFAULT 0;
ALTER TABLE analytics.employee_metrics ADD COLUMN IF NOT EXISTS on_time_completion_rate FLOAT NOT NULL DEFAULT 0;

ALTER TABLE analytics.employee_metrics DROP COLUMN IF EXISTS on_time_completed_tasks;
ALTER TABLE analytics.employee_metrics DROP COLUMN IF EXISTS total_priority_weight_completed;
ALTER TABLE analytics.employee_metrics DROP COLUMN IF EXISTS total_task_duration_seconds;
