-- 005_todos.sql

-- Todo groups (e.g. "Home Reno", "Interview Prep", "Follow-ups")
CREATE TABLE IF NOT EXISTS todo_groups (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name        VARCHAR(255) NOT NULL,
    color       VARCHAR(7) DEFAULT '#6366f1',  -- hex color for UI
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX todo_groups_user_id_idx ON todo_groups(user_id);

-- Todos
CREATE TABLE IF NOT EXISTS todos (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id             UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    application_id      UUID REFERENCES job_applications(id) ON DELETE SET NULL,  -- optional
    group_id            UUID REFERENCES todo_groups(id) ON DELETE SET NULL,       -- optional

    -- Content
    title               TEXT NOT NULL,
    description         TEXT,
    completed           BOOLEAN NOT NULL DEFAULT false,
    completed_at        TIMESTAMPTZ,

    -- Priority: 1=urgent, 2=high, 3=medium, 4=low
    priority            INTEGER NOT NULL DEFAULT 3,

    -- Scheduling
    due_date            DATE,                   -- when it's due
    due_time            TIME,                   -- optional time of day
    reminder_at         TIMESTAMPTZ,            -- don't show until this datetime
    should_carry_over   BOOLEAN NOT NULL DEFAULT false,  -- carry to next day if incomplete

    -- Recurrence
    is_recurring        BOOLEAN NOT NULL DEFAULT false,
    recurrence_rule     JSONB,
    -- Example recurrence_rule:
    -- { "frequency": "weekly", "days": ["monday", "wednesday", "friday"], "time": "09:00" }
    -- { "frequency": "daily", "time": "08:00" }
    -- { "frequency": "monthly", "day_of_month": 1, "time": "10:00" }
    -- { "frequency": "custom", "dates": ["2026-03-10", "2026-03-15"] }

    -- Notifications (future)
    notify              BOOLEAN NOT NULL DEFAULT false,
    notify_minutes_before INTEGER DEFAULT 15,

    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX todos_user_id_idx ON todos(user_id);
CREATE INDEX todos_application_id_idx ON todos(application_id);
CREATE INDEX todos_group_id_idx ON todos(group_id);
CREATE INDEX todos_due_date_idx ON todos(due_date);
CREATE INDEX todos_reminder_at_idx ON todos(reminder_at);
CREATE INDEX todos_completed_idx ON todos(user_id, completed);

-- Triggers for updated_at
CREATE TRIGGER todo_groups_updated_at
    BEFORE UPDATE ON todo_groups
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER todos_updated_at
    BEFORE UPDATE ON todos
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();