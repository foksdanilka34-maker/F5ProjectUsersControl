-- Identity Service Schema

CREATE SCHEMA IF NOT EXISTS identity;

-- Role enum
CREATE TYPE identity.user_role AS ENUM ('admin', 'director', 'manager', 'employee', 'developer');

-- Credentials table for authentication
CREATE TABLE identity.credentials (
    user_id BIGSERIAL PRIMARY KEY,
    login VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    role identity.user_role NOT NULL DEFAULT 'employee',
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Sessions table for refresh tokens
CREATE TABLE identity.sessions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES identity.credentials(user_id) ON DELETE CASCADE,
    refresh_token VARCHAR(500) NOT NULL,
    user_agent VARCHAR(500),
    ip_address VARCHAR(50),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sessions_user_id ON identity.sessions(user_id);
CREATE INDEX idx_sessions_refresh_token ON identity.sessions(refresh_token);
CREATE INDEX idx_sessions_expires_at ON identity.sessions(expires_at);

-- Departments table
CREATE TABLE identity.departments (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Positions table
CREATE TABLE identity.positions (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Skills table
CREATE TABLE identity.skills (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Profiles table
CREATE TABLE identity.profiles (
    id BIGINT PRIMARY KEY REFERENCES identity.credentials(user_id) ON DELETE CASCADE,
    first_name VARCHAR(255) NOT NULL,
    last_name VARCHAR(255) NOT NULL,
    position_id BIGINT NOT NULL REFERENCES identity.positions(id),
    department_id BIGINT REFERENCES identity.departments(id),
    email VARCHAR(255) NOT NULL UNIQUE,
    avatar_url VARCHAR(500),
    hire_date DATE NOT NULL DEFAULT CURRENT_DATE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_profiles_department_id ON identity.profiles(department_id);
CREATE INDEX idx_profiles_position_id ON identity.profiles(position_id);

-- Profile skills junction table
CREATE TABLE identity.profile_skills (
    profile_id BIGINT NOT NULL REFERENCES identity.profiles(id) ON DELETE CASCADE,
    skill_id BIGINT NOT NULL REFERENCES identity.skills(id) ON DELETE CASCADE,
    PRIMARY KEY (profile_id, skill_id)
);
