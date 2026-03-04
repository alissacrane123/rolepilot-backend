package models

import (
	"encoding/json"
	"time"
)

// ============================================
// USER
// ============================================

type User struct {
	ID                 string    `json:"id"`
	Email              string    `json:"email"`
	PasswordHash       string    `json:"-"`
	FullName           string    `json:"full_name"`
	ResumeURL          *string   `json:"resume_url,omitempty"`
	ResumeText         *string   `json:"resume_text,omitempty"`
	Skills             []string  `json:"skills"`
	ExperienceYears    *int      `json:"experience_years,omitempty"`
	TargetRole         *string   `json:"target_role,omitempty"`
	TargetSalaryMin    *int      `json:"target_salary_min,omitempty"`
	TargetSalaryMax    *int      `json:"target_salary_max,omitempty"`
	PreferredLocations []string  `json:"preferred_locations"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type CreateUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UpdateProfileRequest struct {
	FullName           *string  `json:"full_name,omitempty"`
	Skills             []string `json:"skills,omitempty"`
	ExperienceYears    *int     `json:"experience_years,omitempty"`
	TargetRole         *string  `json:"target_role,omitempty"`
	TargetSalaryMin    *int     `json:"target_salary_min,omitempty"`
	TargetSalaryMax    *int     `json:"target_salary_max,omitempty"`
	PreferredLocations []string `json:"preferred_locations,omitempty"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// API response wrapper
type APIResponse struct {
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

// ============================================
// VALID STAGES
// ============================================

var ValidStages = map[string]bool{
	"saved":               true,
	"applied":             true,
	"recruiter_response":  true,
	"phone_screen":        true,
	"technical_interview": true,
	"onsite_final":        true,
	"offer":               true,
	"accepted":            true,
	"rejected":            true,
	"withdrawn":           true,
}

var TerminalStages = map[string]bool{
	"accepted":  true,
	"rejected":  true,
	"withdrawn": true,
}

// ============================================
// JOB APPLICATION
// ============================================

type JobApplication struct {
	ID                  string         `json:"id"`
	UserID              string         `json:"user_id"`
	JobURL              *string        `json:"job_url,omitempty"`
	RawPostingText      *string        `json:"raw_posting_text,omitempty"`
	CompanyName         *string        `json:"company_name,omitempty"`
	CompanySummary      *string        `json:"company_summary,omitempty"`
	RoleTitle           *string        `json:"role_title,omitempty"`
	RoleSummary         *string        `json:"role_summary,omitempty"`
	RequiredSkills      JSONArray      `json:"required_skills"`
	NiceToHaveSkills    JSONArray      `json:"nice_to_have_skills"`
	KeyTechnologies     JSONArray      `json:"key_technologies"`
	ExperienceLevel     *string        `json:"experience_level,omitempty"`
	SalaryRange         *string        `json:"salary_range,omitempty"`
	Location            *string        `json:"location,omitempty"`
	RemotePolicy        *string        `json:"remote_policy,omitempty"`
	MatchScore          *int           `json:"match_score,omitempty"`
	CurrentStage        string         `json:"current_stage"`
	ProcessingStatus    string         `json:"processing_status"`
	AppliedAt           time.Time      `json:"applied_at"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	StageHistory        []StageHistory `json:"stage_history,omitempty"`
	MatchingStrengths   JSONArray      `json:"matching_strengths"`
	PotentialGaps       JSONArray      `json:"potential_gaps"`
	InterviewFocusAreas JSONArray      `json:"interview_focus_areas"`
	SuggestedTalkingPts JSONArray      `json:"suggested_talking_points"`
}

type CreateApplicationRequest struct {
	JobURL         *string `json:"job_url,omitempty"`
	RawPostingText *string `json:"raw_posting_text,omitempty"`
	CompanyName    *string `json:"company_name,omitempty"`
	RoleTitle      *string `json:"role_title,omitempty"`
}

type UpdateStageRequest struct {
	ToStage string `json:"to_stage"`
	Notes   string `json:"notes"`
}

// ============================================
// STAGE HISTORY
// ============================================

type StageHistory struct {
	ID            string    `json:"id"`
	ApplicationID string    `json:"application_id"`
	FromStage     *string   `json:"from_stage,omitempty"`
	ToStage       string    `json:"to_stage"`
	Notes         *string   `json:"notes,omitempty"`
	MovedAt       time.Time `json:"moved_at"`
}

// ============================================
// BOARD VIEW
// ============================================

type BoardView struct {
	Saved              []JobApplication `json:"saved"`
	Applied            []JobApplication `json:"applied"`
	RecruiterResponse  []JobApplication `json:"recruiter_response"`
	PhoneScreen        []JobApplication `json:"phone_screen"`
	TechnicalInterview []JobApplication `json:"technical_interview"`
	OnsiteFinal        []JobApplication `json:"onsite_final"`
	Offer              []JobApplication `json:"offer"`
	Accepted           []JobApplication `json:"accepted"`
	Rejected           []JobApplication `json:"rejected"`
	Withdrawn          []JobApplication `json:"withdrawn"`
}

// ============================================
// HELPERS
// ============================================

// JSONArray handles JSONB string arrays from PostgreSQL
type JSONArray []string

func (j *JSONArray) Scan(src interface{}) error {
	if src == nil {
		*j = []string{}
		return nil
	}
	source, ok := src.([]byte)
	if !ok {
		*j = []string{}
		return nil
	}
	var arr []string
	if err := json.Unmarshal(source, &arr); err != nil {
		*j = []string{}
		return nil
	}
	*j = arr
	return nil
}

// ============================================
// AI EXTRACTION RESPONSE
// ============================================

type AIExtractionResult struct {
	CompanyName         string   `json:"company_name"`
	CompanySummary      string   `json:"company_summary"`
	RoleTitle           string   `json:"role_title"`
	RoleSummary         string   `json:"role_summary"`
	RequiredSkills      []string `json:"required_skills"`
	NiceToHaveSkills    []string `json:"nice_to_have_skills"`
	ExperienceLevel     string   `json:"experience_level"`
	SalaryRange         *string  `json:"salary_range"`
	Location            string   `json:"location"`
	RemotePolicy        string   `json:"remote_policy"`
	KeyTechnologies     []string `json:"key_technologies"`
	MatchScore          int      `json:"match_score"`
	MatchingStrengths   []string `json:"matching_strengths"`
	PotentialGaps       []string `json:"potential_gaps"`
	InterviewFocusAreas []string `json:"interview_focus_areas"`
	SuggestedTalkingPts []string `json:"suggested_talking_points"`
}

// ============================================
// MEETINGS
// ============================================

type Meeting struct {
	ID              string     `json:"id"`
	ApplicationID   string     `json:"application_id"`
	UserID          string     `json:"user_id"`
	Stage           string     `json:"stage"`
	ScheduledAt     *time.Time `json:"scheduled_at,omitempty"`
	DurationMinutes *int       `json:"duration_minutes,omitempty"`
	Timezone        *string    `json:"timezone,omitempty"`
	LocationType    string     `json:"location_type"`
	LocationDetails *string    `json:"location_details,omitempty"`
	MeetingType     *string    `json:"meeting_type,omitempty"`
	ContactName     *string    `json:"contact_name,omitempty"`
	ContactTitle    *string    `json:"contact_title,omitempty"`
	PrepNotes       *string    `json:"prep_notes,omitempty"`
	PostNotes       *string    `json:"post_notes,omitempty"`
	Outcome         *string    `json:"outcome,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type CreateMeetingRequest struct {
	ScheduledAt     *time.Time `json:"scheduled_at,omitempty"`
	DurationMinutes *int       `json:"duration_minutes,omitempty"`
	Timezone        *string    `json:"timezone,omitempty"`
	LocationType    string     `json:"location_type"`
	LocationDetails *string    `json:"location_details,omitempty"`
	MeetingType     *string    `json:"meeting_type,omitempty"`
	ContactName     *string    `json:"contact_name,omitempty"`
	ContactTitle    *string    `json:"contact_title,omitempty"`
	PrepNotes       *string    `json:"prep_notes,omitempty"`
}

type UpdateMeetingRequest struct {
	ScheduledAt     *time.Time `json:"scheduled_at,omitempty"`
	DurationMinutes *int       `json:"duration_minutes,omitempty"`
	Timezone        *string    `json:"timezone,omitempty"`
	LocationType    *string    `json:"location_type,omitempty"`
	LocationDetails *string    `json:"location_details,omitempty"`
	MeetingType     *string    `json:"meeting_type,omitempty"`
	ContactName     *string    `json:"contact_name,omitempty"`
	ContactTitle    *string    `json:"contact_title,omitempty"`
	PrepNotes       *string    `json:"prep_notes,omitempty"`
	PostNotes       *string    `json:"post_notes,omitempty"`
	Outcome         *string    `json:"outcome,omitempty"`
}

type UpdateStageWithMeetingRequest struct {
	ToStage string                `json:"to_stage"`
	Notes   string                `json:"notes"`
	Meeting *CreateMeetingRequest `json:"meeting,omitempty"`
}
