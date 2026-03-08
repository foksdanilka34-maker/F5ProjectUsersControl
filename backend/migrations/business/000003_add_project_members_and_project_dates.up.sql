-- Migration: align business schema with current project repository expectations
-- Adds missing project date columns and members table for existing deployments.

ALTER TABLE business.projects
	ADD COLUMN IF NOT EXISTS start_date TIMESTAMP WITH TIME ZONE,
	ADD COLUMN IF NOT EXISTS end_date TIMESTAMP WITH TIME ZONE;

-- Backfill end_date from legacy deadline column if available
DO $$
BEGIN
	IF EXISTS (
		SELECT 1
		FROM information_schema.columns
		WHERE table_schema = 'business'
			AND table_name = 'projects'
			AND column_name = 'deadline'
	) THEN
		UPDATE business.projects
		SET end_date = deadline
		WHERE end_date IS NULL
			AND deadline IS NOT NULL;
	END IF;
END
$$;

CREATE TABLE IF NOT EXISTS business.project_members (
	project_id BIGINT NOT NULL REFERENCES business.projects(id) ON DELETE CASCADE,
	user_id BIGINT NOT NULL,
	role VARCHAR(50) NOT NULL DEFAULT 'member',
	joined_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
	PRIMARY KEY (project_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_project_members_user_id
	ON business.project_members(user_id);
