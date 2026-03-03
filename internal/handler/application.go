package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/alissacrane123/rolepilot-backend/internal/database"
	"github.com/alissacrane123/rolepilot-backend/internal/middleware"
	"github.com/alissacrane123/rolepilot-backend/internal/models"
	"github.com/alissacrane123/rolepilot-backend/internal/services"
)

type ApplicationHandler struct {
	DB        *database.DB
	AIService *services.AIService
}

func NewApplicationHandler(db *database.DB, ai *services.AIService) *ApplicationHandler {
	return &ApplicationHandler{DB: db, AIService: ai}
}

// Create handles POST /api/applications
func (h *ApplicationHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	var req models.CreateApplicationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Error: "invalid request body"})
		return
	}

	if req.JobURL == nil && req.CompanyName == nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Error: "job_url or company_name is required"})
		return
	}

	app, err := h.DB.CreateApplication(r.Context(), userID, req)
	if err != nil {
		log.Println("CreateApplication error:", err)
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{Error: "failed to create application"})
		return
	}

	// If raw posting text was provided, kick off AI processing in background
	if req.RawPostingText != nil && *req.RawPostingText != "" {
		go h.processApplication(app.ID, userID)
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{Data: app})
}

// processApplication runs AI extraction in the background
func (h *ApplicationHandler) processApplication(appID, userID string) {
	ctx := context.Background()

	h.DB.SetApplicationProcessingStatus(ctx, appID, "processing")

	app, err := h.DB.GetApplication(ctx, appID, userID)
	if err != nil {
		log.Println("AI processing - get app error:", err)
		h.DB.SetApplicationProcessingStatus(ctx, appID, "failed")
		return
	}

	user, err := h.DB.GetUserByID(ctx, userID)
	if err != nil {
		log.Println("AI processing - get user error:", err)
		h.DB.SetApplicationProcessingStatus(ctx, appID, "failed")
		return
	}

	var postingText string
	if app.RawPostingText != nil {
		postingText = *app.RawPostingText
	}

	if postingText == "" {
		h.DB.SetApplicationProcessingStatus(ctx, appID, "needs_text")
		return
	}

	log.Printf("AI processing started for application %s", appID)

	result, err := h.AIService.ExtractJobData(ctx, postingText, user)
	if err != nil {
		log.Println("AI extraction error:", err)
		h.DB.SetApplicationProcessingStatus(ctx, appID, "failed")
		return
	}

	if err := h.DB.UpdateApplicationAIData(ctx, appID, *result); err != nil {
		log.Println("AI save error:", err)
		h.DB.SetApplicationProcessingStatus(ctx, appID, "failed")
		return
	}

	log.Printf("AI processing complete for application %s (match score: %d)", appID, result.MatchScore)
}

// List handles GET /api/applications
func (h *ApplicationHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	apps, err := h.DB.GetApplicationsByUser(r.Context(), userID)
	if err != nil {
		log.Println("ListApplications error:", err)
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{Error: "failed to fetch applications"})
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Data: apps})
}

// Board handles GET /api/applications/board
func (h *ApplicationHandler) Board(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	board, err := h.DB.GetBoardView(r.Context(), userID)
	if err != nil {
		log.Println("BoardView error:", err)
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{Error: "failed to fetch board"})
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Data: board})
}

// Get handles GET /api/applications/{id}
func (h *ApplicationHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	appID := chi.URLParam(r, "id")

	app, err := h.DB.GetApplication(r.Context(), appID, userID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, models.APIResponse{Error: "application not found"})
		return
	}

	history, err := h.DB.GetStageHistory(r.Context(), appID)
	if err == nil {
		app.StageHistory = history
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Data: app})
}

// UpdateStage handles PATCH /api/applications/{id}/stage

func (h *ApplicationHandler) UpdateStage(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	appID := chi.URLParam(r, "id")

	var req models.UpdateStageWithMeetingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Error: "invalid request body"})
		return
	}

	if !models.ValidStages[req.ToStage] {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Error: "invalid stage"})
		return
	}

	app, err := h.DB.GetApplication(r.Context(), appID, userID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, models.APIResponse{Error: "application not found"})
		return
	}

	if models.TerminalStages[app.CurrentStage] {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Error: "cannot transition from terminal stage: " + app.CurrentStage})
		return
	}

	if app.CurrentStage == req.ToStage {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Error: "application is already in stage: " + req.ToStage})
		return
	}

	if err := h.DB.UpdateApplicationStage(r.Context(), appID, userID, app.CurrentStage, req.ToStage, req.Notes); err != nil {
		log.Println("UpdateStage error:", err)
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{Error: "failed to update stage"})
		return
	}

	// If meeting details were included, create the meeting
	if req.Meeting != nil {
		if req.Meeting.LocationType == "" {
			req.Meeting.LocationType = "video"
		}
		_, err := h.DB.CreateMeeting(r.Context(), appID, userID, req.ToStage, *req.Meeting)
		if err != nil {
			log.Println("CreateMeeting on stage change error:", err)
		}
	}

	updated, err := h.DB.GetApplication(r.Context(), appID, userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{Error: "failed to fetch updated application"})
		return
	}

	history, _ := h.DB.GetStageHistory(r.Context(), appID)
	updated.StageHistory = history

	writeJSON(w, http.StatusOK, models.APIResponse{
		Data:    updated,
		Message: "stage updated to " + req.ToStage,
	})
}

// GetStageHistory handles GET /api/applications/{id}/history
func (h *ApplicationHandler) GetStageHistory(w http.ResponseWriter, r *http.Request) {
	appID := chi.URLParam(r, "id")

	history, err := h.DB.GetStageHistory(r.Context(), appID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{Error: "failed to fetch history"})
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Data: history})
}