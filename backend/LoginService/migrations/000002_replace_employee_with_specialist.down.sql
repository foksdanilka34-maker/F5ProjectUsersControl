CREATE TYPE auth.user_role_old AS ENUM ('employee', 'manager', 'admin', 'director');
ALTER TABLE auth.credentials ALTER COLUMN role TYPE auth.user_role_old USING (CASE role::text WHEN 'specialist' THEN 'employee'::text ELSE role::text END)::auth.user_role_old;
DROP TYPE auth.user_role;
ALTER TYPE auth.user_role_old RENAME TO user_role;
