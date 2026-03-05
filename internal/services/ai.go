package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/alissacrane123/rolepilot-backend/internal/models"
)

type AIService struct {
	apiKey  string
	baseURL string
	model   string
}

func NewAIService() *AIService {
	model := os.Getenv("CLAUDE_MODEL")
	if model == "" {
		model = "claude-sonnet-4-5-20250514"
	}
	return &AIService{
		apiKey:  os.Getenv("ANTHROPIC_API_KEY"),
		baseURL: "https://api.anthropic.com/v1/messages",
		model:   model,
	}
}

// ============================================
// CLAUDE API TYPES
// ============================================

type claudeRequest struct {
	Model     string          `json:"model"`
	MaxTokens int             `json:"max_tokens"`
	Messages  []claudeMessage `json:"messages"`
}

type claudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type claudeResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Usage struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

// ============================================
// API CALL
// ============================================

func (s *AIService) callClaude(ctx context.Context, prompt string, maxTokens int) (string, error) {
	if s.apiKey == "" {
		return "", fmt.Errorf("ANTHROPIC_API_KEY not set")
	}

	reqBody := claudeRequest{
		Model:     s.model,
		MaxTokens: maxTokens,
		Messages: []claudeMessage{
			{Role: "user", Content: prompt},
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.baseURL, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", s.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("call claude: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("claude API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var claudeResp claudeResponse
	if err := json.Unmarshal(respBody, &claudeResp); err != nil {
		return "", fmt.Errorf("unmarshal response: %w", err)
	}

	var result strings.Builder
	for _, block := range claudeResp.Content {
		if block.Type == "text" {
			result.WriteString(block.Text)
		}
	}

	return result.String(), nil
}

// ============================================
// EXTRACT JOB DATA
// ============================================

func (s *AIService) ExtractJobData(ctx context.Context, postingText string, user *models.User) (*models.AIExtractionResult, error) {
	// Truncate posting to control costs
	if len(postingText) > 15000 {
		postingText = postingText[:15000]
	}

	// Build user profile section
	var profileParts []string
	if user.TargetRole != nil {
		profileParts = append(profileParts, fmt.Sprintf("Target role: %s", *user.TargetRole))
	}
	if user.ExperienceYears != nil {
		profileParts = append(profileParts, fmt.Sprintf("Experience: %d years", *user.ExperienceYears))
	}
	if len(user.Skills) > 0 {
		profileParts = append(profileParts, fmt.Sprintf("Key skills: %s", strings.Join(user.Skills, ", ")))
	}
	if len(user.PreferredLocations) > 0 {
		profileParts = append(profileParts, fmt.Sprintf("Preferred locations: %s", strings.Join(user.PreferredLocations, ", ")))
	}
	if user.TargetSalaryMin != nil && user.TargetSalaryMax != nil {
		profileParts = append(profileParts, fmt.Sprintf("Salary range: $%d-$%d", *user.TargetSalaryMin, *user.TargetSalaryMax))
	}

	profileSection := "No profile info provided"
	if len(profileParts) > 0 {
		profileSection = strings.Join(profileParts, "\n")
	}

	resumeSection := "No resume provided"
	if user.ResumeText != nil && *user.ResumeText != "" {
		resumeSection = *user.ResumeText
	}

	prompt := fmt.Sprintf(`You are an expert job market analyst and career advisor.

Analyze the following job posting and compare it against my resume and profile.
Extract all structured data from the posting and assess my fit.

<job_posting>
%s
</job_posting>

<my_resume>
%s
</my_resume>

<my_profile>
%s
</my_profile>

Instructions:
1. Extract ALL available information from the job posting
2. If a field is not mentioned in the posting, use null for strings and empty arrays for lists
3. For match_score, be realistic — consider years of experience, specific technologies, and domain knowledge
4. For suggested_talking_points, reference SPECIFIC items from my resume that align with this role
5. For potential_gaps, only list significant gaps, not minor ones. Also consider actions I can take to address them.

Return ONLY valid JSON (no markdown fences, no explanation) matching this exact schema:
{
  "company_name": "string",
  "company_summary": "string (1-2 sentences about what the company does)",
  "role_title": "string",
  "role_summary": "string (2-3 sentences summarizing the role)",
  "required_skills": ["string"],
  "nice_to_have_skills": ["string"],
  "experience_level": "junior|mid|senior|staff|principal",
  "salary_range": "string or null",
  "location": "string",
  "remote_policy": "remote|hybrid|onsite|not_specified",
  "key_technologies": ["string"],
  "match_score": number (0-100),
  "matching_strengths": ["string"],
  "potential_gaps": ["string"],
  "interview_focus_areas": ["string"],
  "suggested_talking_points": ["string"]
}`, postingText, resumeSection, profileSection)

	text, err := s.callClaude(ctx, prompt, 2000)
	if err != nil {
		return nil, fmt.Errorf("extraction call: %w", err)
	}

	// Clean response — strip markdown fences if present
	text = strings.TrimSpace(text)
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimPrefix(text, "```")
	text = strings.TrimSuffix(text, "```")
	text = strings.TrimSpace(text)

	var result models.AIExtractionResult
	if err := json.Unmarshal([]byte(text), &result); err != nil {
		return nil, fmt.Errorf("parse AI response: %w (response: %.200s)", err, text)
	}

	// Ensure non-nil slices
	if result.RequiredSkills == nil { result.RequiredSkills = []string{} }
	if result.NiceToHaveSkills == nil { result.NiceToHaveSkills = []string{} }
	if result.KeyTechnologies == nil { result.KeyTechnologies = []string{} }
	if result.MatchingStrengths == nil { result.MatchingStrengths = []string{} }
	if result.PotentialGaps == nil { result.PotentialGaps = []string{} }
	if result.InterviewFocusAreas == nil { result.InterviewFocusAreas = []string{} }
	if result.SuggestedTalkingPts == nil { result.SuggestedTalkingPts = []string{} }

	return &result, nil
}


func (s *AIService) GenerateCoverLetter(ctx context.Context, app *models.JobApplication, user *models.User, tone string) (string, error) {
	if tone == "" {
		tone = "professional"
	}

	resumeSection := "No resume provided"
	if user.ResumeText != nil && *user.ResumeText != "" {
		resumeSection = *user.ResumeText
	}

	var postingSection string
	if app.RawPostingText != nil && *app.RawPostingText != "" {
		postingSection = *app.RawPostingText
		if len(postingSection) > 10000 {
			postingSection = postingSection[:10000]
		}
	} else {
		// Build from extracted data
		postingSection = fmt.Sprintf("Company: %s\nRole: %s\n", safeStr(app.CompanyName), safeStr(app.RoleTitle))
		if app.RoleSummary != nil {
			postingSection += fmt.Sprintf("Role Summary: %s\n", *app.RoleSummary)
		}
	}

	prompt := fmt.Sprintf(`You are an expert cover letter writer.

Write a cover letter for the following job application. The letter should be %s in tone.

<job_posting>
%s
</job_posting>

<my_resume>
%s
</my_resume>

<applicant_name>%s</applicant_name>

Instructions:
1. Write 3-4 paragraphs maximum
2. Reference SPECIFIC experiences from the resume that are relevant to this role
3. Do NOT use generic phrases like "I am writing to express my interest" or "I believe I would be a great fit"
4. Do NOT include placeholders like [Company Name] — use the actual company and role names
5. Be specific about what the applicant brings and why this role is a good match
6. Match the tone: professional = formal but warm, conversational = friendly and direct, enthusiastic = energetic and passionate
7. Do NOT include a header, date, or address block — just the letter body
8. End with "Best regards," followed by the applicant's full name on the next line

Return ONLY the cover letter text, no commentary.`, tone, postingSection, resumeSection, user.FullName)

	return s.callClaude(ctx, prompt, 1500)
}

// helper for nil string pointers
func safeStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}