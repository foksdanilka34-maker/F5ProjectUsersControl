CREATE TABLE IF NOT EXISTS business.extensions (
    id                BIGSERIAL PRIMARY KEY,
    key               VARCHAR(64) UNIQUE NOT NULL,
    name              VARCHAR(255) NOT NULL,
    description       TEXT,
    base_url          VARCHAR(500) NOT NULL,
    shared_secret_enc BYTEA NOT NULL,
    task_panel_url    VARCHAR(500),
    project_tab_url   VARCHAR(500),
    project_tab_label VARCHAR(100),
    events            TEXT[] NOT NULL DEFAULT '{}',
    is_active         BOOLEAN NOT NULL DEFAULT TRUE,
    created_at        TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS business.project_extensions (
    project_id   BIGINT NOT NULL REFERENCES business.projects(id) ON DELETE CASCADE,
    extension_id BIGINT NOT NULL REFERENCES business.extensions(id) ON DELETE CASCADE,
    is_enabled   BOOLEAN NOT NULL DEFAULT TRUE,
    installed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    PRIMARY KEY (project_id, extension_id)
);

CREATE INDEX IF NOT EXISTS idx_project_extensions_project_id ON business.project_extensions(project_id);

CREATE TABLE IF NOT EXISTS business.task_entity_properties (
    id           BIGSERIAL PRIMARY KEY,
    task_id      BIGINT NOT NULL REFERENCES business.tasks(id) ON DELETE CASCADE,
    extension_id BIGINT NOT NULL REFERENCES business.extensions(id) ON DELETE CASCADE,
    property_key VARCHAR(100) NOT NULL,
    value        JSONB NOT NULL,
    updated_at   TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE (task_id, extension_id, property_key)
);

CREATE INDEX IF NOT EXISTS idx_task_entity_properties_task_id ON business.task_entity_properties(task_id);
