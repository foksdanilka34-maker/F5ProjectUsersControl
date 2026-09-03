CREATE TABLE IF NOT EXISTS business.gitlab_integrations (
    project_id            BIGINT PRIMARY KEY REFERENCES business.projects(id) ON DELETE CASCADE,
    base_url              VARCHAR(255) NOT NULL DEFAULT 'https://gitlab.com',
    gitlab_project_id     BIGINT NOT NULL,
    access_token_enc      BYTEA,
    webhook_secret        VARCHAR(128) NOT NULL,
    default_branch        VARCHAR(100) NOT NULL DEFAULT 'main',
    task_key_prefix       VARCHAR(16) NOT NULL DEFAULT 'F5',
    auto_move_in_progress BOOLEAN NOT NULL DEFAULT TRUE,
    auto_move_review      BOOLEAN NOT NULL DEFAULT TRUE,
    auto_close_on_merge   BOOLEAN NOT NULL DEFAULT TRUE,
    is_active             BOOLEAN NOT NULL DEFAULT TRUE,
    created_at            TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS business.task_git_links (
    id          BIGSERIAL PRIMARY KEY,
    task_id     BIGINT NOT NULL REFERENCES business.tasks(id) ON DELETE CASCADE,
    kind        VARCHAR(20) NOT NULL CHECK (kind IN ('BRANCH', 'MERGE_REQUEST', 'COMMIT')),
    external_id VARCHAR(255) NOT NULL,
    title       TEXT,
    state       VARCHAR(30),
    web_url     TEXT,
    author_name VARCHAR(255),
    created_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE (task_id, kind, external_id)
);

CREATE INDEX IF NOT EXISTS idx_task_git_links_task_id ON business.task_git_links(task_id);

CREATE TABLE IF NOT EXISTS business.gitlab_pipelines (
    id           BIGSERIAL PRIMARY KEY,
    project_id   BIGINT NOT NULL REFERENCES business.projects(id) ON DELETE CASCADE,
    task_id      BIGINT REFERENCES business.tasks(id) ON DELETE CASCADE,
    pipeline_id  BIGINT NOT NULL,
    ref          VARCHAR(255) NOT NULL,
    sha          VARCHAR(64) NOT NULL,
    status       VARCHAR(30) NOT NULL,
    duration_sec INTEGER,
    web_url      TEXT,
    started_at   TIMESTAMP WITH TIME ZONE,
    finished_at  TIMESTAMP WITH TIME ZONE,
    updated_at   TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, pipeline_id)
);

CREATE INDEX IF NOT EXISTS idx_gitlab_pipelines_task_id ON business.gitlab_pipelines(task_id);

CREATE TABLE IF NOT EXISTS business.gitlab_webhook_events (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id    BIGINT NOT NULL,
    event_type    VARCHAR(50) NOT NULL,
    delivery_id   VARCHAR(120) UNIQUE,
    payload       JSONB NOT NULL,
    status        VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    retry_count   INTEGER NOT NULL DEFAULT 0,
    error_message TEXT,
    created_at    TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    processed_at  TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_gitlab_webhook_events_pending
    ON business.gitlab_webhook_events(created_at) WHERE status = 'PENDING';
