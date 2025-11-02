-- Remove due_date constraints from projects and tasks tables
-- This allows more flexible date management at the application level

ALTER TABLE project.projects
DROP CONSTRAINT IF EXISTS check_projects_due_date;

ALTER TABLE project.tasks
DROP CONSTRAINT IF EXISTS check_tasks_due_date;
