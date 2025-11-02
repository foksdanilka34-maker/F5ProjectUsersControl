DROP TABLE IF EXISTS project.project_metrics;

DROP INDEX IF EXISTS project.idx_task_status_history_changed_at;
DROP INDEX IF EXISTS project.idx_task_status_history_task_id;

DROP TABLE IF EXISTS project.task_status_history;

ALTER TABLE project.tasks
DROP CONSTRAINT IF EXISTS check_task_completion_after_start;

ALTER TABLE project.tasks
DROP COLUMN IF EXISTS completed_at,
DROP COLUMN IF EXISTS started_at;
