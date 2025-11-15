CREATE TYPE auth.user_role_new AS ENUM ('specialist', 'manager', 'admin', 'director');
ALTER TABLE auth.credentials ALTER COLUMN role TYPE auth.user_role_new USING (CASE role::text WHEN 'employee' THEN 'specialist'::text ELSE role::text END)::auth.user_role_new;
DROP TYPE auth.user_role;
ALTER TYPE auth.user_role_new RENAME TO user_role;
