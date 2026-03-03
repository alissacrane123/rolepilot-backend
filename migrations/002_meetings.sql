-- 002_meetings.sql

CREATE TABLE IF NOT EXISTS meetings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    application_id UUID NOT NULL REFERENCES job_applications(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Scheduling
    stage VARCHAR(50) NOT NULL,
    scheduled_at TIMESTAMPTZ,
    duration_minutes INTEGER,
    timezone VARCHAR(50),

    -- Location
    location_type VARCHAR(20) DEFAULT 'video',  -- video, phone, onsite
    location_details TEXT,                        -- Zoom link, office address, phone number

    -- Meeting details
    meeting_type VARCHAR(50),                   -- behavioral, technical, system_design, hiring_manager, culture_fit, take_home, panel
    contact_name VARCHAR(255),
    contact_title VARCHAR(255),

    -- Notes
    prep_notes TEXT,
    post_notes TEXT,
    outcome VARCHAR(20),                          -- pending, passed, failed, unknown

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_meetings_application_id ON meetings(application_id);
CREATE INDEX idx_meetings_user_id ON meetings(user_id);
CREATE INDEX idx_meetings_scheduled_at ON meetings(scheduled_at);