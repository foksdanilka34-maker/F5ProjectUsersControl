CREATE SCHEMA IF NOT EXISTS employees;

CREATE TABLE IF NOT EXISTS employees.departments (
    id   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS employees.positions (
    id   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS employees.profiles (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    first_name      VARCHAR(255) NOT NULL,
    last_name       VARCHAR(255) NOT NULL,
    position_id     UUID REFERENCES employees.positions(id),
    email           VARCHAR(255) UNIQUE NOT NULL,
    department_id   UUID REFERENCES employees.departments(id) ON DELETE SET NULL,
    
    avatar_url      VARCHAR(255),
    hire_date       DATE CHECK (hire_date <= CURRENT_DATE),

    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS employees.skills (
    id   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS employees.employee_skills (
    employee_id UUID NOT NULL REFERENCES employees.profiles(id) ON DELETE CASCADE,
    skill_id    UUID NOT NULL REFERENCES employees.skills(id) ON DELETE CASCADE,
    PRIMARY KEY (employee_id, skill_id)
);

CREATE INDEX idx_profiles_department_id ON employees.profiles(department_id);
CREATE INDEX idx_employee_skills_skill_id ON employees.employee_skills(skill_id);