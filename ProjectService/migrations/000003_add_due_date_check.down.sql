
DROP INDEX IF EXISTS project.idx_tasks_due_date;
DROP INDEX IF EXISTS project.idx_projects_due_date;

ALTER TABLE project.tasks
DROP CONSTRAINT IF EXISTS check_tasks_due_date;

ALTER TABLE project.projects
DROP CONSTRAINT IF EXISTS check_projects_due_date;
