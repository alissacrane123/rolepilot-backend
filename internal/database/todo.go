package database

import (
	"context"
	"fmt"
	"time"

	"github.com/alissacrane123/rolepilot-backend/internal/models"
)

// ============================================
// TODO GROUP QUERIES
// ============================================

func (db *DB) CreateTodoGroup(ctx context.Context, userID string, req models.CreateTodoGroupRequest) (*models.TodoGroup, error) {
	g := &models.TodoGroup{}
	color := req.Color
	if color == "" {
		color = "#6366f1"
	}
	err := db.Pool.QueryRow(ctx, `
		INSERT INTO todo_groups (user_id, name, color)
		VALUES ($1, $2, $3)
		RETURNING id, user_id, name, color, created_at, updated_at
	`, userID, req.Name, color).Scan(
		&g.ID, &g.UserID, &g.Name, &g.Color, &g.CreatedAt, &g.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create todo group: %w", err)
	}
	return g, nil
}

func (db *DB) GetTodoGroups(ctx context.Context, userID string) ([]models.TodoGroup, error) {
	rows, err := db.Pool.Query(ctx, `
		SELECT id, user_id, name, color, created_at, updated_at
		FROM todo_groups WHERE user_id = $1
		ORDER BY name ASC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("get todo groups: %w", err)
	}
	defer rows.Close()

	var groups []models.TodoGroup
	for rows.Next() {
		var g models.TodoGroup
		err := rows.Scan(&g.ID, &g.UserID, &g.Name, &g.Color, &g.CreatedAt, &g.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan todo group: %w", err)
		}
		groups = append(groups, g)
	}
	if groups == nil {
		groups = []models.TodoGroup{}
	}
	return groups, nil
}

func (db *DB) UpdateTodoGroup(ctx context.Context, groupID, userID string, req models.UpdateTodoGroupRequest) (*models.TodoGroup, error) {
	g := &models.TodoGroup{}
	err := db.Pool.QueryRow(ctx, `
		UPDATE todo_groups SET
			name = COALESCE($3, name),
			color = COALESCE($4, color)
		WHERE id = $1 AND user_id = $2
		RETURNING id, user_id, name, color, created_at, updated_at
	`, groupID, userID, req.Name, req.Color).Scan(
		&g.ID, &g.UserID, &g.Name, &g.Color, &g.CreatedAt, &g.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("update todo group: %w", err)
	}
	return g, nil
}

func (db *DB) DeleteTodoGroup(ctx context.Context, groupID, userID string) error {
	result, err := db.Pool.Exec(ctx, `
		DELETE FROM todo_groups WHERE id = $1 AND user_id = $2
	`, groupID, userID)
	if err != nil {
		return fmt.Errorf("delete todo group: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("group not found")
	}
	return nil
}

// ============================================
// TODO QUERIES
// ============================================

const todoSelectFields = `
	t.id, t.user_id, t.application_id, t.group_id,
	t.title, t.description, t.completed, t.completed_at,
	t.priority, t.due_date::text, t.due_time::text,
	t.reminder_at, t.should_carry_over,
	t.is_recurring, t.recurrence_rule,
	t.notify, t.notify_minutes_before,
	t.created_at, t.updated_at,
	g.name, g.color,
	a.company_name, a.role_title
`

const todoJoins = `
	FROM todos t
	LEFT JOIN todo_groups g ON g.id = t.group_id
	LEFT JOIN job_applications a ON a.id = t.application_id
`

func scanTodo(scan func(dest ...interface{}) error) (*models.Todo, error) {
	var t models.Todo
	err := scan(
		&t.ID, &t.UserID, &t.ApplicationID, &t.GroupID,
		&t.Title, &t.Description, &t.Completed, &t.CompletedAt,
		&t.Priority, &t.DueDate, &t.DueTime,
		&t.ReminderAt, &t.ShouldCarryOver,
		&t.IsRecurring, &t.RecurrenceRule,
		&t.Notify, &t.NotifyMinutesBefore,
		&t.CreatedAt, &t.UpdatedAt,
		&t.GroupName, &t.GroupColor,
		&t.CompanyName, &t.RoleTitle,
	)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (db *DB) CreateTodo(ctx context.Context, userID string, req models.CreateTodoRequest) (*models.Todo, error) {
	priority := 3
	if req.Priority != nil {
		priority = *req.Priority
	}
	shouldCarry := true
	if req.ShouldCarryOver != nil {
		shouldCarry = *req.ShouldCarryOver
	}
	isRecurring := false
	if req.IsRecurring != nil {
		isRecurring = *req.IsRecurring
	}
	notify := false
	if req.Notify != nil {
		notify = *req.Notify
	}

	var id string
	err := db.Pool.QueryRow(ctx, `
		INSERT INTO todos (
			user_id, application_id, group_id,
			title, description, priority,
			due_date, due_time, reminder_at,
			should_carry_over, is_recurring, recurrence_rule,
			notify, notify_minutes_before
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id
	`, userID, req.ApplicationID, req.GroupID,
		req.Title, req.Description, priority,
		req.DueDate, req.DueTime, req.ReminderAt,
		shouldCarry, isRecurring, req.RecurrenceRule,
		notify, req.NotifyMinutesBefore,
	).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("create todo: %w", err)
	}

	return db.GetTodo(ctx, id, userID)
}

func (db *DB) GetTodo(ctx context.Context, todoID, userID string) (*models.Todo, error) {
	row := db.Pool.QueryRow(ctx, `
		SELECT `+todoSelectFields+todoJoins+`
		WHERE t.id = $1 AND t.user_id = $2
	`, todoID, userID)

	t, err := scanTodo(row.Scan)
	if err != nil {
		return nil, fmt.Errorf("get todo: %w", err)
	}
	return t, nil
}

func (db *DB) GetTodos(ctx context.Context, userID string, filter TodoFilter) ([]models.Todo, error) {
	query := `SELECT ` + todoSelectFields + todoJoins + ` WHERE t.user_id = $1`
	args := []interface{}{userID}
	argIdx := 2

	if filter.Completed != nil {
		query += fmt.Sprintf(" AND t.completed = $%d", argIdx)
		args = append(args, *filter.Completed)
		argIdx++
	}
	if filter.ApplicationID != nil {
		query += fmt.Sprintf(" AND t.application_id = $%d", argIdx)
		args = append(args, *filter.ApplicationID)
		argIdx++
	}
	if filter.GroupID != nil {
		query += fmt.Sprintf(" AND t.group_id = $%d", argIdx)
		args = append(args, *filter.GroupID)
		argIdx++
	}
	if filter.DueDate != nil {
		query += fmt.Sprintf(" AND t.due_date = $%d", argIdx)
		args = append(args, *filter.DueDate)
		argIdx++
	}
	if filter.DueBefore != nil {
		query += fmt.Sprintf(" AND t.due_date <= $%d", argIdx)
		args = append(args, *filter.DueBefore)
		argIdx++
	}
	if filter.DueAfter != nil {
		query += fmt.Sprintf(" AND t.due_date >= $%d", argIdx)
		args = append(args, *filter.DueAfter)
		argIdx++
	}
	if filter.Search != nil && *filter.Search != "" {
		query += fmt.Sprintf(" AND (t.title ILIKE $%d OR t.description ILIKE $%d)", argIdx, argIdx)
		args = append(args, "%"+*filter.Search+"%")
		argIdx++
	}
	// Hide future reminders unless explicitly asking
	if !filter.IncludeReminders {
		query += " AND (t.reminder_at IS NULL OR t.reminder_at <= $" + fmt.Sprintf("%d", argIdx) + ")"
		args = append(args, time.Now())
		argIdx++
	}

	query += " ORDER BY t.completed ASC, t.priority ASC, t.due_date ASC NULLS LAST, t.created_at DESC"

	rows, err := db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("get todos: %w", err)
	}
	defer rows.Close()

	var todos []models.Todo
	for rows.Next() {
		t, err := scanTodo(rows.Scan)
		if err != nil {
			return nil, fmt.Errorf("scan todo: %w", err)
		}
		todos = append(todos, *t)
	}
	if todos == nil {
		todos = []models.Todo{}
	}
	return todos, nil
}

// GetTodosForDate returns todos due on a specific date + carried-over incomplete todos from previous days
func (db *DB) GetTodosForDate(ctx context.Context, userID, date string) ([]models.Todo, error) {
	rows, err := db.Pool.Query(ctx, `
		SELECT `+todoSelectFields+todoJoins+`
		WHERE t.user_id = $1
		  AND (
			t.due_date = $2
			OR (t.should_carry_over = true AND t.completed = false AND t.due_date < $2)
		  )
		  AND (t.reminder_at IS NULL OR t.reminder_at <= NOW())
		ORDER BY t.completed ASC, t.priority ASC, t.due_time ASC NULLS LAST
	`, userID, date)
	if err != nil {
		return nil, fmt.Errorf("get todos for date: %w", err)
	}
	defer rows.Close()

	var todos []models.Todo
	for rows.Next() {
		t, err := scanTodo(rows.Scan)
		if err != nil {
			return nil, fmt.Errorf("scan todo: %w", err)
		}
		todos = append(todos, *t)
	}
	if todos == nil {
		todos = []models.Todo{}
	}
	return todos, nil
}

func (db *DB) UpdateTodo(ctx context.Context, todoID, userID string, req models.UpdateTodoRequest) (*models.Todo, error) {
	// Handle completion timestamp
	var completedAt *time.Time
	if req.Completed != nil && *req.Completed {
		now := time.Now()
		completedAt = &now
	}

	_, err := db.Pool.Exec(ctx, `
		UPDATE todos SET
			application_id = COALESCE($3, application_id),
			group_id = COALESCE($4, group_id),
			title = COALESCE($5, title),
			description = COALESCE($6, description),
			completed = COALESCE($7, completed),
			completed_at = CASE WHEN $7 = true THEN COALESCE($8, completed_at) WHEN $7 = false THEN NULL ELSE completed_at END,
			priority = COALESCE($9, priority),
			due_date = COALESCE($10, due_date),
			due_time = COALESCE($11, due_time),
			reminder_at = COALESCE($12, reminder_at),
			should_carry_over = COALESCE($13, should_carry_over),
			is_recurring = COALESCE($14, is_recurring),
			recurrence_rule = COALESCE($15, recurrence_rule),
			notify = COALESCE($16, notify),
			notify_minutes_before = COALESCE($17, notify_minutes_before)
		WHERE id = $1 AND user_id = $2
	`, todoID, userID,
		req.ApplicationID, req.GroupID,
		req.Title, req.Description,
		req.Completed, completedAt,
		req.Priority,
		req.DueDate, req.DueTime,
		req.ReminderAt, req.ShouldCarryOver,
		req.IsRecurring, req.RecurrenceRule,
		req.Notify, req.NotifyMinutesBefore,
	)
	if err != nil {
		return nil, fmt.Errorf("update todo: %w", err)
	}

	return db.GetTodo(ctx, todoID, userID)
}

func (db *DB) DeleteTodo(ctx context.Context, todoID, userID string) error {
	result, err := db.Pool.Exec(ctx, `
		DELETE FROM todos WHERE id = $1 AND user_id = $2
	`, todoID, userID)
	if err != nil {
		return fmt.Errorf("delete todo: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("todo not found")
	}
	return nil
}

// TodoFilter for flexible querying
type TodoFilter struct {
	Completed        *bool
	ApplicationID    *string
	GroupID          *string
	DueDate          *string
	DueBefore        *string
	DueAfter         *string
	Search           *string
	IncludeReminders bool
}