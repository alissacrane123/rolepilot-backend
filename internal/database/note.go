package database

import (
	"context"
	"fmt"

	"github.com/alissacrane123/rolepilot-backend/internal/models"
)

func (db *DB) CreateNote(ctx context.Context, applicationID, userID string, req models.CreateNoteRequest) (*models.Note, error) {
	n := &models.Note{}
	title := req.Title
	if title == "" {
		title = "Untitled Note"
	}
	err := db.Pool.QueryRow(ctx, `
		INSERT INTO notes (application_id, user_id, title, content)
		VALUES ($1, $2, $3, $4)
		RETURNING id, application_id, user_id, title, content, created_at, updated_at
	`, applicationID, userID, title, req.Content).Scan(
		&n.ID, &n.ApplicationID, &n.UserID, &n.Title, &n.Content,
		&n.CreatedAt, &n.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create note: %w", err)
	}
	return n, nil
}

func (db *DB) GetNotesByApplication(ctx context.Context, applicationID, userID string) ([]models.Note, error) {
	rows, err := db.Pool.Query(ctx, `
		SELECT id, application_id, user_id, title, content, created_at, updated_at
		FROM notes
		WHERE application_id = $1 AND user_id = $2
		ORDER BY updated_at DESC
	`, applicationID, userID)
	if err != nil {
		return nil, fmt.Errorf("get notes: %w", err)
	}
	defer rows.Close()

	var notes []models.Note
	for rows.Next() {
		var n models.Note
		err := rows.Scan(&n.ID, &n.ApplicationID, &n.UserID, &n.Title, &n.Content, &n.CreatedAt, &n.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan note: %w", err)
		}
		notes = append(notes, n)
	}
	if notes == nil {
		notes = []models.Note{}
	}
	return notes, nil
}

func (db *DB) UpdateNote(ctx context.Context, noteID, userID string, req models.UpdateNoteRequest) (*models.Note, error) {
	n := &models.Note{}
	err := db.Pool.QueryRow(ctx, `
		UPDATE notes SET
			title = COALESCE($3, title),
			content = COALESCE($4, content)
		WHERE id = $1 AND user_id = $2
		RETURNING id, application_id, user_id, title, content, created_at, updated_at
	`, noteID, userID, req.Title, req.Content).Scan(
		&n.ID, &n.ApplicationID, &n.UserID, &n.Title, &n.Content,
		&n.CreatedAt, &n.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("update note: %w", err)
	}
	return n, nil
}

func (db *DB) DeleteNote(ctx context.Context, noteID, userID string) error {
	result, err := db.Pool.Exec(ctx, `
		DELETE FROM notes WHERE id = $1 AND user_id = $2
	`, noteID, userID)
	if err != nil {
		return fmt.Errorf("delete note: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("note not found")
	}
	return nil
}