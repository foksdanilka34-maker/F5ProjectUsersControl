-- Rollback: remove objects introduced in 000003

DROP INDEX IF EXISTS business.idx_project_members_user_id;
DROP TABLE IF EXISTS business.project_members;

ALTER TABLE business.projects
	DROP COLUMN IF EXISTS start_date,
	DROP COLUMN IF EXISTS end_date;
