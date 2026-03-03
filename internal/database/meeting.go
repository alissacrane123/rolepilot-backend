package database

import (
	"context"
	"fmt"

	"github.com/alissacrane123/rolepilot-backend/internal/models"
)

func (db *DB) CreateMeeting(ctx context.Context, applicationID, userID, stage string, req models.CreateMeetingRequest) (*models.Meeting, error) {
	m := &models.Meeting{}
	err := db.Pool.QueryRow(ctx, `
		INSERT INTO meetings (
			application_id, user_id, stage,
			scheduled_at, duration_minutes, timezone,
			location_type, location_details,
			meeting_type, contact_name, contact_title,
			prep_notes
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, application_id, user_id, stage,
		          scheduled_at, duration_minutes, timezone,
		          location_type, location_details,
		          meeting_type, contact_name, contact_title,
		          prep_notes, post_notes, outcome,
		          created_at, updated_at
	`, applicationID, userID, stage,
		req.ScheduledAt, req.DurationMinutes, req.Timezone,
		req.LocationType, req.LocationDetails,
		req.MeetingType, req.ContactName, req.ContactTitle,
		req.PrepNotes,
	).Scan(
		&m.ID, &m.ApplicationID, &m.UserID, &m.Stage,
		&m.ScheduledAt, &m.DurationMinutes, &m.Timezone,
		&m.LocationType, &m.LocationDetails,
		&m.MeetingType, &m.ContactName, &m.ContactTitle,
		&m.PrepNotes, &m.PostNotes, &m.Outcome,
		&m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create meeting: %w", err)
	}
	return m, nil
}

func (db *DB) GetMeetingsByApplication(ctx context.Context, applicationID string) ([]models.Meeting, error) {
	rows, err := db.Pool.Query(ctx, `
		SELECT id, application_id, user_id, stage,
		       scheduled_at, duration_minutes, timezone,
		       location_type, location_details,
		       meeting_type, contact_name, contact_title,
		       prep_notes, post_notes, outcome,
		       created_at, updated_at
		FROM meetings
		WHERE application_id = $1
		ORDER BY scheduled_at ASC NULLS LAST
	`, applicationID)
	if err != nil {
		return nil, fmt.Errorf("get meetings: %w", err)
	}
	defer rows.Close()

	var meetings []models.Meeting
	for rows.Next() {
		var m models.Meeting
		err := rows.Scan(
			&m.ID, &m.ApplicationID, &m.UserID, &m.Stage,
			&m.ScheduledAt, &m.DurationMinutes, &m.Timezone,
			&m.LocationType, &m.LocationDetails,
			&m.MeetingType, &m.ContactName, &m.ContactTitle,
			&m.PrepNotes, &m.PostNotes, &m.Outcome,
			&m.CreatedAt, &m.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan meeting: %w", err)
		}
		meetings = append(meetings, m)
	}
	if meetings == nil {
		meetings = []models.Meeting{}
	}
	return meetings, nil
}

func (db *DB) GetUpcomingMeetings(ctx context.Context, userID string) ([]models.Meeting, error) {
	rows, err := db.Pool.Query(ctx, `
		SELECT m.id, m.application_id, m.user_id, m.stage,
		       m.scheduled_at, m.duration_minutes, m.timezone,
		       m.location_type, m.location_details,
		       m.meeting_type, m.contact_name, m.contact_title,
		       m.prep_notes, m.post_notes, m.outcome,
		       m.created_at, m.updated_at
		FROM meetings m
		JOIN job_applications a ON a.id = m.application_id
		WHERE m.user_id = $1
		  AND m.scheduled_at >= NOW()
		  AND a.current_stage NOT IN ('accepted', 'rejected', 'withdrawn')
		ORDER BY m.scheduled_at ASC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("get upcoming meetings: %w", err)
	}
	defer rows.Close()

	var meetings []models.Meeting
	for rows.Next() {
		var m models.Meeting
		err := rows.Scan(
			&m.ID, &m.ApplicationID, &m.UserID, &m.Stage,
			&m.ScheduledAt, &m.DurationMinutes, &m.Timezone,
			&m.LocationType, &m.LocationDetails,
			&m.MeetingType, &m.ContactName, &m.ContactTitle,
			&m.PrepNotes, &m.PostNotes, &m.Outcome,
			&m.CreatedAt, &m.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan meeting: %w", err)
		}
		meetings = append(meetings, m)
	}
	if meetings == nil {
		meetings = []models.Meeting{}
	}
	return meetings, nil
}

func (db *DB) UpdateMeeting(ctx context.Context, meetingID, userID string, req models.UpdateMeetingRequest) (*models.Meeting, error) {
	m := &models.Meeting{}
	err := db.Pool.QueryRow(ctx, `
		UPDATE meetings SET
			scheduled_at = COALESCE($3, scheduled_at),
			duration_minutes = COALESCE($4, duration_minutes),
			timezone = COALESCE($5, timezone),
			location_type = COALESCE($6, location_type),
			location_details = COALESCE($7, location_details),
			meeting_type = COALESCE($8, meeting_type),
			contact_name = COALESCE($9, contact_name),
			contact_title = COALESCE($10, contact_title),
			prep_notes = COALESCE($11, prep_notes),
			post_notes = COALESCE($12, post_notes),
			outcome = COALESCE($13, outcome),
			updated_at = NOW()
		WHERE id = $1 AND user_id = $2
		RETURNING id, application_id, user_id, stage,
		          scheduled_at, duration_minutes, timezone,
		          location_type, location_details,
		          meeting_type, contact_name, contact_title,
		          prep_notes, post_notes, outcome,
		          created_at, updated_at
	`, meetingID, userID,
		req.ScheduledAt, req.DurationMinutes, req.Timezone,
		req.LocationType, req.LocationDetails,
		req.MeetingType, req.ContactName, req.ContactTitle,
		req.PrepNotes, req.PostNotes, req.Outcome,
	).Scan(
		&m.ID, &m.ApplicationID, &m.UserID, &m.Stage,
		&m.ScheduledAt, &m.DurationMinutes, &m.Timezone,
		&m.LocationType, &m.LocationDetails,
		&m.MeetingType, &m.ContactName, &m.ContactTitle,
		&m.PrepNotes, &m.PostNotes, &m.Outcome,
		&m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("update meeting: %w", err)
	}
	return m, nil
}

func (db *DB) DeleteMeeting(ctx context.Context, meetingID, userID string) error {
	result, err := db.Pool.Exec(ctx, `
		DELETE FROM meetings WHERE id = $1 AND user_id = $2
	`, meetingID, userID)
	if err != nil {
		return fmt.Errorf("delete meeting: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("meeting not found")
	}
	return nil
}