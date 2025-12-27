-- Rollback: Revert status and priority normalization

-- Remove new constraints
ALTER TABLE business.tasks DROP CONSTRAINT IF EXISTS chk_task_status;
ALTER TABLE business.tasks DROP CONSTRAINT IF EXISTS chk_task_priority;

-- Convert priority back to integer
ALTER TABLE business.tasks ADD COLUMN priority_int INTEGER;

UPDATE business.tasks SET priority_int = CASE LOWER(priority)
    WHEN 'low' THEN 1
    WHEN 'medium' THEN 2
    WHEN 'high' THEN 3
    WHEN 'critical' THEN 4
    ELSE 2
END;

ALTER TABLE business.tasks DROP COLUMN priority;
ALTER TABLE business.tasks RENAME COLUMN priority_int TO priority;
ALTER TABLE business.tasks ALTER COLUMN priority SET NOT NULL;
ALTER TABLE business.tasks ALTER COLUMN priority SET DEFAULT 2;

-- Restore old constraints
ALTER TABLE business.tasks ADD CONSTRAINT chk_task_status 
    CHECK (status IN ('TODO', 'IN_PROGRESS', 'IN_REVIEW', 'DONE', 'todo', 'in_progress', 'review', 'done'));

ALTER TABLE business.tasks ADD CONSTRAINT chk_task_priority 
    CHECK (priority BETWEEN 1 AND 4);

-- Remove comments
COMMENT ON COLUMN business.tasks.status IS NULL;
COMMENT ON COLUMN business.tasks.priority IS NULL;
