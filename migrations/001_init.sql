-- 001_init.sql
-- RolePilot Database Schema

-- Enable UUID generation
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================
-- USERS
-- ============================================
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(100) NOT NULL,
    resume_url TEXT,
    resume_text TEXT,
    skills TEXT[] DEFAULT '{}',
    experience_years INTEGER,
    target_role VARCHAR(100),
    target_salary_min INTEGER,
    target_salary_max INTEGER,
    preferred_locations TEXT[] DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================
-- JOB APPLICATIONS
-- ============================================
CREATE TABLE IF NOT EXISTS job_applications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Source data
    job_url TEXT,
    raw_posting_text TEXT,

    -- AI-extracted fields
    company_name VARCHAR(255),
    company_summary TEXT,
    role_title VARCHAR(255),
    role_summary TEXT,
    required_skills JSONB DEFAULT '[]',
    nice_to_have_skills JSONB DEFAULT '[]',
    key_technologies JSONB DEFAULT '[]',
    experience_level VARCHAR(50),
    salary_range VARCHAR(100),
    location VARCHAR(255),
    remote_policy VARCHAR(50) DEFAULT 'not_specified',

    -- AI analysis
    match_score INTEGER,
    matching_strengths JSONB DEFAULT '[]',
    potential_gaps JSONB DEFAULT '[]',
    interview_focus_areas JSONB DEFAULT '[]',
    suggested_talking_points JSONB DEFAULT '[]',

    -- Status tracking
    current_stage VARCHAR(50) NOT NULL DEFAULT 'applied',
    processing_status VARCHAR(20) NOT NULL DEFAULT 'pending',

    -- Timestamps
    applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_job_applications_user_id ON job_applications(user_id);
CREATE INDEX idx_job_applications_current_stage ON job_applications(current_stage);

-- ============================================
-- STAGE HISTORY
-- ============================================
CREATE TABLE IF NOT EXISTS stage_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    application_id UUID NOT NULL REFERENCES job_applications(id) ON DELETE CASCADE,
    from_stage VARCHAR(50),
    to_stage VARCHAR(50) NOT NULL,
    notes TEXT,
    moved_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_stage_history_application_id ON stage_history(application_id);

-- ============================================
-- COVER LETTERS
-- ============================================
CREATE TABLE IF NOT EXISTS cover_letters (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    application_id UUID NOT NULL REFERENCES job_applications(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    version INTEGER NOT NULL DEFAULT 1,
    tone VARCHAR(50) DEFAULT 'professional',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_cover_letters_application_id ON cover_letters(application_id);

-- ============================================
-- SAVED JOB LISTINGS (future feature)
-- ============================================
CREATE TABLE IF NOT EXISTS saved_job_listings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    source_url TEXT NOT NULL,
    title VARCHAR(255),
    company VARCHAR(255),
    match_score INTEGER,
    match_reasoning TEXT,
    discovered_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_saved_job_listings_user_id ON saved_job_listings(user_id);
