CREATE SCHEMA IF NOT EXISTS auth;

CREATE TYPE auth.user_role AS ENUM ('employee', 'manager', 'admin', 'director');

CREATE TABLE IF NOT EXISTS auth.credentials (
    user_id UUID UNIQUE PRIMARY KEY,
    role user_role NOT NULL, 
    login VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(72) NOT NULL, 
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    is_active BOOLEAN NOT NULL DEFAULT true
);

CREATE TABLE auth.sessions (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID NOT NULL REFERENCES auth(user_id) ON DELETE CASCADE,
    refresh_token VARCHAR(255) UNIQUE NOT NULL,                    
    user_agent    VARCHAR(255),                           
    ip_address    VARCHAR(45),                            
    expires_at    TIMESTAMPTZ NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX exp_at ON auth.sessions(expires_at);