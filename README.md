# RolePilot

An AI-powered job application tracker that helps you manage your job search, analyze job postings, and prepare for interviews.

## Features

- **Job Application Tracking** — Kanban-style board with stages: Applied → Recruiter Response → Phone Screen → Technical Interview → Onsite → Offer → Accepted/Rejected/Withdrawn
- **AI Job Analysis** — Paste a job posting and AI extracts company info, required skills, salary range, and more
- **Resume Matching** — Upload your resume and get a match score, strengths, gaps, and suggested talking points for each application
- **Stage History** — Full audit trail with notes at every stage transition
- **Cover Letter Generation** — AI-generated cover letters tailored to each job (coming soon)

## Tech Stack

- **Backend:** Go, Chi router
- **Database:** PostgreSQL
- **AI:** Anthropic Claude API
- **Auth:** JWT

## Getting Started

### Prerequisites

- Go 1.22+
- PostgreSQL
- [Anthropic API key](https://console.anthropic.com/)
- poppler (for PDF text extraction): `brew install poppler`

### Setup

1. Clone the repo:
   ```bash
   git clone https://github.com/alissacrane123/rolepilot-backend.git
   cd rolepilot-backend
   ```

2. Create the database:
   ```bash
   createdb rolepilot
   ```

3. Run migrations:
   ```bash
   psql postgres://localhost:5432/rolepilot -f migrations/001_init.sql
   ```

4. Create your `.env` file:
   ```bash
   cp .env.example .env
   ```
   Then fill in your `DATABASE_URL`, `JWT_SECRET`, and `ANTHROPIC_API_KEY`.

5. Install dependencies and run:
   ```bash
   go mod tidy
   go run cmd/server/main.go
   ```

## API Endpoints

### Auth
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/auth/register` | Create account |
| POST | `/api/auth/login` | Login, returns JWT |

### Profile
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/profile` | Get your profile |
| PATCH | `/api/profile` | Update skills, experience, preferences |
| POST | `/api/profile/resume` | Upload resume file (PDF, TXT, MD) |
| POST | `/api/profile/resume/text` | Paste resume text directly |

### Applications
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/applications` | Create application (triggers AI analysis) |
| GET | `/api/applications` | List all applications |
| GET | `/api/applications/board` | Board view grouped by stage |
| GET | `/api/applications/{id}` | Application detail with AI data |
| PATCH | `/api/applications/{id}/stage` | Move to new stage with notes |
| GET | `/api/applications/{id}/history` | Full stage transition history |

### Meetings
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/applications/{id}/meetings` | Add meeting to an application |
| GET | `/api/applications/{id}/meetings` | List meetings for an application |
| GET | `/api/meetings/upcoming` | All upcoming meetings (for calendar) |
| PATCH | `/api/meetings/{meetingId}` | Update meeting details/notes/outcome |
| DELETE | `/api/meetings/{meetingId}` | Delete a meeting |

## Project Structure

```
rolepilot-backend/
├── cmd/server/main.go           # Entry point, routes
├── internal/
│   ├── database/
│   │   ├── db.go                # Connection pool, user queries
│   │   ├── application.go       # Application + stage queries
│   │   └── meeting.go           # Meeting CRUD queries
│   ├── handler/
│   │   ├── auth.go              # Register, login, profile, resume upload
│   │   ├── application.go       # CRUD, stage transitions, AI trigger
│   │   ├── meeting.go           # Meeting CRUD + upcoming
│   │   ├── helpers.go           # Response helpers
│   │   └── pdf.go               # PDF text extraction
│   ├── middleware/
│   │   └── auth.go              # JWT middleware
│   ├── models/
│   │   └── models.go            # All types and request/response structs
│   └── services/
│       └── ai.go                # Claude API integration
├── migrations/
│   ├── 001_init.sql             # Database schema
│   └── 002_meetings.sql         # Meetings table
├── .env.example
└── go.mod
```


## Database Schema
```mermaid
erDiagram
    users {
        UUID id PK
        VARCHAR email UK
        VARCHAR password_hash
        VARCHAR full_name
        TEXT resume_url
        TEXT resume_text
        TEXT[] skills
        INTEGER experience_years
        VARCHAR target_role
        INTEGER target_salary_min
        INTEGER target_salary_max
        TEXT[] preferred_locations
        TIMESTAMPTZ created_at
        TIMESTAMPTZ updated_at
    }

    job_applications {
        UUID id PK
        UUID user_id FK
        TEXT job_url
        TEXT raw_posting_text
        VARCHAR company_name
        TEXT company_summary
        VARCHAR role_title
        TEXT role_summary
        JSONB required_skills
        JSONB nice_to_have_skills
        JSONB key_technologies
        VARCHAR experience_level
        VARCHAR salary_range
        VARCHAR location
        VARCHAR remote_policy
        INTEGER match_score
        JSONB matching_strengths
        JSONB potential_gaps
        JSONB interview_focus_areas
        JSONB suggested_talking_points
        VARCHAR current_stage
        VARCHAR processing_status
        TIMESTAMPTZ applied_at
        TIMESTAMPTZ created_at
        TIMESTAMPTZ updated_at
    }

    stage_history {
        UUID id PK
        UUID application_id FK
        VARCHAR from_stage
        VARCHAR to_stage
        TEXT notes
        TIMESTAMPTZ moved_at
    }

    cover_letters {
        UUID id PK
        UUID application_id FK
        TEXT content
        INTEGER version
        VARCHAR tone
        TIMESTAMPTZ created_at
    }

    saved_job_listings {
        UUID id PK
        UUID user_id FK
        TEXT source_url
        VARCHAR title
        VARCHAR company
        INTEGER match_score
        TEXT match_reasoning
        TIMESTAMPTZ discovered_at
    }

    meetings {
        UUID id PK
        UUID application_id FK
        UUID user_id FK
        VARCHAR stage
        TIMESTAMPTZ scheduled_at
        INTEGER duration_minutes
        VARCHAR timezone
        VARCHAR location_type
        TEXT location_details
        VARCHAR meeting_type
        VARCHAR contact_name
        VARCHAR contact_title
        TEXT prep_notes
        TEXT post_notes
        VARCHAR outcome
        TIMESTAMPTZ created_at
        TIMESTAMPTZ updated_at
    }

    job_applications ||--o{ meetings : "has many"

    users ||--o{ job_applications : "has many"
    users ||--o{ saved_job_listings : "has many"
    job_applications ||--o{ stage_history : "has many"
    job_applications ||--o{ cover_letters : "has many"
```