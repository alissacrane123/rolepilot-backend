package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/alissacrane123/rolepilot-backend/internal/models"
)

type DB struct {
	Pool *pgxpool.Pool
}

func New(databaseURL string) (*DB, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	config.MaxConns = 20
	config.MinConns = 2

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("ping: %w", err)
	}

	return &DB{Pool: pool}, nil
}

func (db *DB) Close() {
	db.Pool.Close()
}

// ============================================
// USER QUERIES
// ============================================

func (db *DB) CreateUser(ctx context.Context, email, passwordHash, fullName string) (*models.User, error) {
	user := &models.User{}
	err := db.Pool.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, full_name)
		VALUES ($1, $2, $3)
		RETURNING id, email, full_name, skills, preferred_locations, created_at, updated_at
	`, email, passwordHash, fullName).Scan(
		&user.ID, &user.Email, &user.FullName,
		&user.Skills, &user.PreferredLocations,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return user, nil
}

func (db *DB) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	user := &models.User{}
	err := db.Pool.QueryRow(ctx, `
		SELECT id, email, password_hash, full_name, resume_url, resume_text,
		       skills, experience_years, target_role, target_salary_min,
		       target_salary_max, preferred_locations, created_at, updated_at
		FROM users WHERE email = $1
	`, email).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.FullName,
		&user.ResumeURL, &user.ResumeText,
		&user.Skills, &user.ExperienceYears, &user.TargetRole,
		&user.TargetSalaryMin, &user.TargetSalaryMax,
		&user.PreferredLocations, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return user, nil
}

func (db *DB) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	user := &models.User{}
	err := db.Pool.QueryRow(ctx, `
		SELECT id, email, password_hash, full_name, resume_url, resume_text,
		       skills, experience_years, target_role, target_salary_min,
		       target_salary_max, preferred_locations, created_at, updated_at
		FROM users WHERE id = $1
	`, id).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.FullName,
		&user.ResumeURL, &user.ResumeText,
		&user.Skills, &user.ExperienceYears, &user.TargetRole,
		&user.TargetSalaryMin, &user.TargetSalaryMax,
		&user.PreferredLocations, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return user, nil
}


func (db *DB) UpdateUserProfile(ctx context.Context, userID string, req models.UpdateProfileRequest) (*models.User, error) {
	user := &models.User{}
	err := db.Pool.QueryRow(ctx, `
		UPDATE users SET
			full_name = COALESCE($2, full_name),
			skills = COALESCE($3, skills),
			experience_years = COALESCE($4, experience_years),
			target_role = COALESCE($5, target_role),
			target_salary_min = COALESCE($6, target_salary_min),
			target_salary_max = COALESCE($7, target_salary_max),
			preferred_locations = COALESCE($8, preferred_locations),
			updated_at = NOW()
		WHERE id = $1
		RETURNING id, email, full_name, resume_url, resume_text,
		          skills, experience_years, target_role, target_salary_min,
		          target_salary_max, preferred_locations, created_at, updated_at
	`, userID, req.FullName, req.Skills, req.ExperienceYears,
		req.TargetRole, req.TargetSalaryMin, req.TargetSalaryMax,
		req.PreferredLocations,
	).Scan(
		&user.ID, &user.Email, &user.FullName,
		&user.ResumeURL, &user.ResumeText,
		&user.Skills, &user.ExperienceYears, &user.TargetRole,
		&user.TargetSalaryMin, &user.TargetSalaryMax,
		&user.PreferredLocations, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("update profile: %w", err)
	}
	return user, nil
}

func (db *DB) UpdateUserResume(ctx context.Context, userID, resumeURL, resumeText string) error {
	_, err := db.Pool.Exec(ctx, `
		UPDATE users SET resume_url = $2, resume_text = $3, updated_at = NOW()
		WHERE id = $1
	`, userID, resumeURL, resumeText)
	return err
}