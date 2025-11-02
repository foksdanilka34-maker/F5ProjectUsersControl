ALTER TABLE project.projects
ADD CONSTRAINT check_projects_due_date 
CHECK (due_date IS NULL OR (due_date >= created_at AND due_date >= CURRENT_DATE));

ALTER TABLE project.tasks
ADD CONSTRAINT check_tasks_due_date 
CHECK (due_date >= created_at AND due_date >= CURRENT_DATE);

CREATE INDEX IF NOT EXISTS idx_projects_due_date 
ON project.projects (due_date) 
WHERE due_date IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_tasks_due_date 
ON project.tasks (due_date);
