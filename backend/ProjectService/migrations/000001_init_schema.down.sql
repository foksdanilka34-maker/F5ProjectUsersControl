DROP TABLE IF EXISTS project.project_metrics CASCADE;
DROP TABLE IF EXISTS project.task_status_history CASCADE;
DROP TABLE IF EXISTS project.project_members CASCADE;
DROP TABLE IF EXISTS project.tasks CASCADE;
DROP TABLE IF EXISTS project.projects CASCADE;
DROP TABLE IF EXISTS project.users_meta CASCADE;

DROP TYPE IF EXISTS project.project_status;
DROP TYPE IF EXISTS project.task_priority;
DROP TYPE IF EXISTS project.task_status;

DROP SCHEMA IF EXISTS project CASCADE;
