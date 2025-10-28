CREATE SCHEMA IF NOT EXISTS project;

CREATE TYPE project.task_status AS ENUM ('TASK_STATUS_UNSPECIFIED', 'TODO', 'IN_PROGRESS', 'REVIEW', 'DONE');
CREATE TYPE project.task_priority AS ENUM ('PRIORITY_UNSPECIFIED', 'LOW', 'MEDIUM', 'HIGH', 'CRITICAL');
CREATE TYPE project.project_status AS ENUM ('PROJECT_STATUS_UNSPECIFIED', 'ACTIVE', 'ON_HOLD', 'ARCHIVED');

CREATE TABLE IF NOT EXISTS project.projects (
    project_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_name VARCHAR(255) NOT NULL,
    project_description TEXT,

    manager_id UUID,
    project_status project.project_status NOT NULL DEFAULT 'ACTIVE',

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    due_date TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS project.tasks (
    task_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES project.projects(project_id) ON DELETE CASCADE,
    task_name VARCHAR(255) NOT NULL,
    task_description TEXT,

    task_priority project.task_priority NOT NULL DEFAULT 'PRIORITY_UNSPECIFIED',
    task_status project.task_status NOT NULL DEFAULT 'TASK_STATUS_UNSPECIFIED',

    creator_id UUID NOT NULL,
    assign_id UUID,

    order_index INT NOT NULL DEFAULT 0,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    due_date TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS project.users_meta (
    user_id UUID UNIQUE NOT NULL,
    user_name VARCHAR(255) NOT NULL,
    user_photo VARCHAR(255),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS project.project_members (
    project_id UUID NOT NULL REFERENCES project.projects(project_id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    role VARCHAR(50) NOT NULL,
    added_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (project_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_projects_manager_status_created_at
    ON project.projects (manager_id, project_status, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_tasks_project_status_order
    ON project.tasks (project_id, task_status, order_index);
