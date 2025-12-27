-- Migration: Normalize status and priority values
-- This migration ensures all status values are UPPERCASE and priority uses consistent string values

-- Step 1: Remove old constraints
ALTER TABLE business.tasks DROP CONSTRAINT IF EXISTS chk_task_status;
ALTER TABLE business.tasks DROP CONSTRAINT IF EXISTS chk_task_priority;

-- Step 2: Normalize status values to UPPERCASE
UPDATE business.tasks SET status = UPPER(status);

-- Step 3: Convert priority from integer to string if needed, or normalize string values
-- First check if priority is integer and convert to string
DO $$
BEGIN
    -- Check if priority is currently integer type
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_schema = 'business' 
        AND table_name = 'tasks' 
        AND column_name = 'priority' 
        AND data_type = 'integer'
    ) THEN
        -- Add temporary column
        ALTER TABLE business.tasks ADD COLUMN priority_new VARCHAR(50);
        
        -- Convert integer to string
        UPDATE business.tasks SET priority_new = CASE priority
            WHEN 1 THEN 'low'
            WHEN 2 THEN 'medium'
            WHEN 3 THEN 'high'
            WHEN 4 THEN 'critical'
            ELSE 'medium'
        END;
        
        -- Drop old column and rename new
        ALTER TABLE business.tasks DROP COLUMN priority;
        ALTER TABLE business.tasks RENAME COLUMN priority_new TO priority;
        ALTER TABLE business.tasks ALTER COLUMN priority SET NOT NULL;
        ALTER TABLE business.tasks ALTER COLUMN priority SET DEFAULT 'medium';
    ELSE
        -- Priority is already string, just normalize to lowercase
        UPDATE business.tasks SET priority = LOWER(priority);
    END IF;
END $$;

-- Step 4: Update default for status
ALTER TABLE business.tasks ALTER COLUMN status SET DEFAULT 'TODO';

-- Step 5: Add new constraints with correct values
ALTER TABLE business.tasks ADD CONSTRAINT chk_task_status 
    CHECK (status IN ('TODO', 'IN_PROGRESS', 'IN_REVIEW', 'DONE'));

ALTER TABLE business.tasks ADD CONSTRAINT chk_task_priority 
    CHECK (priority IN ('low', 'medium', 'high', 'critical'));

-- Step 6: Recreate indexes for better performance
DROP INDEX IF EXISTS business.idx_tasks_status;
DROP INDEX IF EXISTS business.idx_tasks_priority;
CREATE INDEX idx_tasks_status ON business.tasks(status);
CREATE INDEX idx_tasks_priority ON business.tasks(priority);

-- Add comment for documentation
COMMENT ON COLUMN business.tasks.status IS 'Task status: TODO, IN_PROGRESS, IN_REVIEW, DONE (always UPPERCASE)';
COMMENT ON COLUMN business.tasks.priority IS 'Task priority: low, medium, high, critical (always lowercase)';
