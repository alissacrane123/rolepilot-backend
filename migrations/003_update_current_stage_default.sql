-- 003_update_current_stage_default.sql
-- Update default stage for new job applications

ALTER TABLE job_applications
    ALTER COLUMN current_stage SET DEFAULT 'saved';
