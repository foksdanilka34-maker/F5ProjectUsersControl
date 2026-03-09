-- Add photo_url to users_meta
ALTER TABLE business.users_meta ADD COLUMN IF NOT EXISTS photo_url VARCHAR(500);

-- Task comments table
CREATE TABLE IF NOT EXISTS business.task_comments (
    id BIGSERIAL PRIMARY KEY,
    task_id BIGINT NOT NULL REFERENCES business.tasks(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_task_comments_task_id ON business.task_comments(task_id);

-- Task history table
CREATE TABLE IF NOT EXISTS business.task_history (
    id BIGSERIAL PRIMARY KEY,
    task_id BIGINT NOT NULL REFERENCES business.tasks(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL,
    field VARCHAR(100) NOT NULL,
    old_value TEXT,
    new_value TEXT,
    changed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_task_history_task_id ON business.task_history(task_id);
