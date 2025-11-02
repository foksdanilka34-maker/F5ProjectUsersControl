ALTER TABLE project.projects
DROP CONSTRAINT IF EXISTS check_projects_due_date;

ALTER TABLE project.tasks
DROP CONSTRAINT IF EXISTS check_tasks_due_date;

ALTER TABLE project.projects
ADD CONSTRAINT check_projects_due_date 
CHECK (due_date IS NULL OR due_date >= created_at);

ALTER TABLE project.tasks
ADD CONSTRAINT check_tasks_due_date 
CHECK (due_date >= created_at);
