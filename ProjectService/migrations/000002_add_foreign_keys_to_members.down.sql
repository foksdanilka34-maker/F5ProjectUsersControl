ALTER TABLE project.project_members
    DROP CONSTRAINT IF EXISTS fk_project_members_user;

ALTER TABLE project.tasks
    DROP CONSTRAINT IF EXISTS fk_tasks_assignee;

ALTER TABLE project.tasks
    DROP CONSTRAINT IF EXISTS fk_tasks_creator;

ALTER TABLE project.projects
    DROP CONSTRAINT IF EXISTS fk_projects_manager;
