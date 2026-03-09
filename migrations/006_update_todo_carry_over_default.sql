-- 006_update_todo_carry_over_default.sql
-- Set todos.should_carry_over to default true

ALTER TABLE todos
    ALTER COLUMN should_carry_over SET DEFAULT true;
