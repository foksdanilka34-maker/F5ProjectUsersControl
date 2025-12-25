-- Logs Service Schema

CREATE SCHEMA IF NOT EXISTS logs;

-- Log entries table
CREATE TABLE logs.entries (
    id UUID PRIMARY KEY,
    service VARCHAR(100) NOT NULL,
    level VARCHAR(20) NOT NULL,
    message TEXT NOT NULL,
    user_id UUID,
    action VARCHAR(100),
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_logs_service ON logs.entries(service);
CREATE INDEX idx_logs_level ON logs.entries(level);
CREATE INDEX idx_logs_user_id ON logs.entries(user_id);
CREATE INDEX idx_logs_created_at ON logs.entries(created_at DESC);
CREATE INDEX idx_logs_action ON logs.entries(action);

-- Log level constraint
ALTER TABLE logs.entries ADD CONSTRAINT chk_log_level 
    CHECK (level IN ('DEBUG', 'INFO', 'WARN', 'ERROR'));

-- Retention policy - partition by month (optional, for production)
-- This can be implemented with pg_partman extension

-- Cleanup function to delete old logs (older than 90 days)
CREATE OR REPLACE FUNCTION logs.cleanup_old_entries() RETURNS void AS $$
BEGIN
    DELETE FROM logs.entries WHERE created_at < NOW() - INTERVAL '90 days';
END;
$$ LANGUAGE plpgsql;
