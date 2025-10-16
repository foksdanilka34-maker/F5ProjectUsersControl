DROP INDEX IF EXISTS auth.exp_at_idx;
DROP TABLE IF EXISTS auth.sessions;
DROP TABLE IF EXISTS auth.credentials;
DROP TYPE IF EXISTS auth.user_role;
DROP SCHEMA IF EXISTS auth;