package database

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/alissacrane123/rolepilot-backend/internal/models"
)

// ============================================
// JOB APPLICATION QUERIES
// ============================================

func (db *DB) CreateApplication(ctx context.Context, userID string, req models.CreateApplicationRequest) (*models.JobApplication, error) {
	app := &models.JobApplication{}
	err := db.Pool.QueryRow(ctx, `
		INSERT INTO job_applications (user_id, job_url, raw_posting_text, company_name, role_title, processing_status)
		VALUES ($1, $2, $3, $4, $5, 'pending')
		RETURNING id, user_id, job_url, raw_posting_text, company_name, role_title,
		          current_stage, processing_status, applied_at, created_at, updated_at
	`, userID, req.JobURL, req.RawPostingText, req.CompanyName, req.RoleTitle).Scan(
		&app.ID, &app.UserID, &app.JobURL, &app.RawPostingText,
		&app.CompanyName, &app.RoleTitle,
		&app.CurrentStage, &app.ProcessingStatus,
		&app.AppliedAt, &app.CreatedAt, &app.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create application: %w", err)
	}

	// Create initial stage history entry
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO stage_history (application_id, to_stage, notes)
		VALUES ($1, 'saved', 'Application created')
	`, app.ID)
	if err != nil {
		return nil, fmt.Errorf("create initial stage history: %w", err)
	}

	return app, nil
}

// REPLACE your existing GetApplication function in database/application.go with this.
// Make sure "encoding/json" is in your imports.

func (db *DB) GetApplication(ctx context.Context, appID, userID string) (*models.JobApplication, error) {
	app := &models.JobApplication{}
	var (
		reqSkills  []byte
		niceSkills []byte
		keyTech    []byte
		strengths  []byte
		gaps       []byte
		focusAreas []byte
		talkingPts []byte
	)

	err := db.Pool.QueryRow(ctx, `
		SELECT id, user_id, job_url, raw_posting_text,
		       company_name, company_summary, role_title, role_summary,
		       required_skills, nice_to_have_skills, key_technologies,
		       experience_level, salary_range, location, remote_policy,
		       match_score, matching_strengths, potential_gaps,
		       interview_focus_areas, suggested_talking_points,
		       current_stage, processing_status, applied_at, created_at, updated_at
		FROM job_applications
		WHERE id = $1 AND user_id = $2
	`, appID, userID).Scan(
		&app.ID, &app.UserID, &app.JobURL, &app.RawPostingText,
		&app.CompanyName, &app.CompanySummary, &app.RoleTitle, &app.RoleSummary,
		&reqSkills, &niceSkills, &keyTech,
		&app.ExperienceLevel, &app.SalaryRange, &app.Location, &app.RemotePolicy,
		&app.MatchScore, &strengths, &gaps,
		&focusAreas, &talkingPts,
		&app.CurrentStage, &app.ProcessingStatus, &app.AppliedAt, &app.CreatedAt, &app.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get application: %w", err)
	}

	// Parse JSONB arrays
	json.Unmarshal(reqSkills, &app.RequiredSkills)
	json.Unmarshal(niceSkills, &app.NiceToHaveSkills)
	json.Unmarshal(keyTech, &app.KeyTechnologies)
	json.Unmarshal(strengths, &app.MatchingStrengths)
	json.Unmarshal(gaps, &app.PotentialGaps)
	json.Unmarshal(focusAreas, &app.InterviewFocusAreas)
	json.Unmarshal(talkingPts, &app.SuggestedTalkingPts)

	// Ensure non-nil slices for clean JSON output
	if app.RequiredSkills == nil {
		app.RequiredSkills = []string{}
	}
	if app.NiceToHaveSkills == nil {
		app.NiceToHaveSkills = []string{}
	}
	if app.KeyTechnologies == nil {
		app.KeyTechnologies = []string{}
	}
	if app.MatchingStrengths == nil {
		app.MatchingStrengths = []string{}
	}
	if app.PotentialGaps == nil {
		app.PotentialGaps = []string{}
	}
	if app.InterviewFocusAreas == nil {
		app.InterviewFocusAreas = []string{}
	}
	if app.SuggestedTalkingPts == nil {
		app.SuggestedTalkingPts = []string{}
	}

	return app, nil
}

func (db *DB) GetApplicationsByUser(ctx context.Context, userID string) ([]models.JobApplication, error) {
	rows, err := db.Pool.Query(ctx, `
		SELECT id, user_id, job_url, company_name, role_title,
		       experience_level, salary_range, location, remote_policy,
		       match_score, current_stage, processing_status,
		       applied_at, created_at, updated_at
		FROM job_applications
		WHERE user_id = $1
		ORDER BY updated_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("get applications: %w", err)
	}
	defer rows.Close()

	var apps []models.JobApplication
	for rows.Next() {
		var app models.JobApplication
		err := rows.Scan(
			&app.ID, &app.UserID, &app.JobURL, &app.CompanyName, &app.RoleTitle,
			&app.ExperienceLevel, &app.SalaryRange, &app.Location, &app.RemotePolicy,
			&app.MatchScore, &app.CurrentStage, &app.ProcessingStatus,
			&app.AppliedAt, &app.CreatedAt, &app.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan application: %w", err)
		}
		apps = append(apps, app)
	}

	if apps == nil {
		apps = []models.JobApplication{}
	}
	return apps, nil
}

func (db *DB) GetBoardView(ctx context.Context, userID string) (*models.BoardView, error) {
	apps, err := db.GetApplicationsByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	board := &models.BoardView{
		Saved:              []models.JobApplication{},
		Applied:            []models.JobApplication{},
		RecruiterResponse:  []models.JobApplication{},
		PhoneScreen:        []models.JobApplication{},
		TechnicalInterview: []models.JobApplication{},
		OnsiteFinal:        []models.JobApplication{},
		Offer:              []models.JobApplication{},
		Accepted:           []models.JobApplication{},
		Rejected:           []models.JobApplication{},
		Withdrawn:          []models.JobApplication{},
	}

	for _, app := range apps {
		switch app.CurrentStage {
		case "saved":
			board.Saved = append(board.Saved, app)
		case "applied":
			board.Applied = append(board.Applied, app)
		case "recruiter_response":
			board.RecruiterResponse = append(board.RecruiterResponse, app)
		case "phone_screen":
			board.PhoneScreen = append(board.PhoneScreen, app)
		case "technical_interview":
			board.TechnicalInterview = append(board.TechnicalInterview, app)
		case "onsite_final":
			board.OnsiteFinal = append(board.OnsiteFinal, app)
		case "offer":
			board.Offer = append(board.Offer, app)
		case "accepted":
			board.Accepted = append(board.Accepted, app)
		case "rejected":
			board.Rejected = append(board.Rejected, app)
		case "withdrawn":
			board.Withdrawn = append(board.Withdrawn, app)
		}
	}

	return board, nil
}

func (db *DB) UpdateApplicationStage(ctx context.Context, appID, userID, fromStage, toStage, notes string) error {
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// Update the application
	result, err := tx.Exec(ctx, `
		UPDATE job_applications
		SET current_stage = $3, updated_at = NOW()
		WHERE id = $1 AND user_id = $2
	`, appID, userID, toStage)
	if err != nil {
		return fmt.Errorf("update stage: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("application not found")
	}

	// Insert stage history
	_, err = tx.Exec(ctx, `
		INSERT INTO stage_history (application_id, from_stage, to_stage, notes)
		VALUES ($1, $2, $3, $4)
	`, appID, fromStage, toStage, notes)
	if err != nil {
		return fmt.Errorf("insert stage history: %w", err)
	}

	return tx.Commit(ctx)
}

// ============================================
// STAGE HISTORY QUERIES
// ============================================

func (db *DB) GetStageHistory(ctx context.Context, appID string) ([]models.StageHistory, error) {
	rows, err := db.Pool.Query(ctx, `
		SELECT id, application_id, from_stage, to_stage, notes, moved_at
		FROM stage_history
		WHERE application_id = $1
		ORDER BY moved_at ASC
	`, appID)
	if err != nil {
		return nil, fmt.Errorf("get stage history: %w", err)
	}
	defer rows.Close()

	var history []models.StageHistory
	for rows.Next() {
		var h models.StageHistory
		err := rows.Scan(&h.ID, &h.ApplicationID, &h.FromStage, &h.ToStage, &h.Notes, &h.MovedAt)
		if err != nil {
			return nil, fmt.Errorf("scan history: %w", err)
		}
		history = append(history, h)
	}

	if history == nil {
		history = []models.StageHistory{}
	}
	return history, nil
}

// ============================================
// AI EXTRACTION QUERIES
// ============================================

func (db *DB) UpdateApplicationAIData(ctx context.Context, appID string, data models.AIExtractionResult) error {
	reqSkills, _ := json.Marshal(data.RequiredSkills)
	niceSkills, _ := json.Marshal(data.NiceToHaveSkills)
	keyTech, _ := json.Marshal(data.KeyTechnologies)
	strengths, _ := json.Marshal(data.MatchingStrengths)
	gaps, _ := json.Marshal(data.PotentialGaps)
	focusAreas, _ := json.Marshal(data.InterviewFocusAreas)
	talkingPts, _ := json.Marshal(data.SuggestedTalkingPts)

	_, err := db.Pool.Exec(ctx, `
		UPDATE job_applications SET
			company_name = $2, company_summary = $3,
			role_title = $4, role_summary = $5,
			required_skills = $6, nice_to_have_skills = $7,
			key_technologies = $8, experience_level = $9,
			salary_range = $10, location = $11, remote_policy = $12,
			match_score = $13, matching_strengths = $14,
			potential_gaps = $15, interview_focus_areas = $16,
			suggested_talking_points = $17,
			processing_status = 'complete',
			updated_at = NOW()
		WHERE id = $1
	`, appID,
		data.CompanyName, data.CompanySummary,
		data.RoleTitle, data.RoleSummary,
		reqSkills, niceSkills,
		keyTech, data.ExperienceLevel,
		data.SalaryRange, data.Location, data.RemotePolicy,
		data.MatchScore, strengths,
		gaps, focusAreas,
		talkingPts,
	)
	return err
}

func (db *DB) SetApplicationProcessingStatus(ctx context.Context, appID, status string) error {
	_, err := db.Pool.Exec(ctx, `
		UPDATE job_applications SET processing_status = $2, updated_at = NOW()
		WHERE id = $1
	`, appID, status)
	return err
}
