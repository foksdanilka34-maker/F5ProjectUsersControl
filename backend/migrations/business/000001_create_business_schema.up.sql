-- Business Service Schema

CREATE SCHEMA IF NOT EXISTS business;

-- Users meta table (synced from IdentityService via NATS)
CREATE TABLE business.users_meta (
    user_id BIGINT PRIMARY KEY,
    full_name VARCHAR(500) NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Projects table
CREATE TABLE business.projects (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    owner_id BIGINT NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'ACTIVE',
    deadline TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_projects_owner_id ON business.projects(owner_id);
CREATE INDEX idx_projects_status ON business.projects(status);

-- Project members junction table
CREATE TABLE business.project_members (
    project_id BIGINT NOT NULL REFERENCES business.projects(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'member',
    joined_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    PRIMARY KEY (project_id, user_id)
);

CREATE INDEX idx_project_members_user_id ON business.project_members(user_id);

-- Tasks table
CREATE TABLE business.tasks (
    id BIGSERIAL PRIMARY KEY,
    project_id BIGINT NOT NULL REFERENCES business.projects(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    assignee_id BIGINT,
    creator_id BIGINT,
    status VARCHAR(50) NOT NULL DEFAULT 'TODO',
    priority INTEGER NOT NULL DEFAULT 2,
    order_index INTEGER NOT NULL DEFAULT 0,
    due_date TIMESTAMP WITH TIME ZONE,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_tasks_project_id ON business.tasks(project_id);
CREATE INDEX idx_tasks_assignee_id ON business.tasks(assignee_id);
CREATE INDEX idx_tasks_status ON business.tasks(status);
CREATE INDEX idx_tasks_due_date ON business.tasks(due_date);

-- Task status constraint
ALTER TABLE business.tasks ADD CONSTRAINT chk_task_status 
    CHECK (status IN ('TODO', 'IN_PROGRESS', 'IN_REVIEW', 'DONE'));

-- Task priority constraint (1=low, 2=medium, 3=high, 4=critical)
ALTER TABLE business.tasks ADD CONSTRAINT chk_task_priority 
    CHECK (priority BETWEEN 1 AND 4);

-- Project status constraint
ALTER TABLE business.projects ADD CONSTRAINT chk_project_status 
    CHECK (status IN ('ACTIVE', 'ON_HOLD', 'COMPLETED', 'ARCHIVED'));
