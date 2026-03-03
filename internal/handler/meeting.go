package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/alissacrane123/rolepilot-backend/internal/database"
	"github.com/alissacrane123/rolepilot-backend/internal/middleware"
	"github.com/alissacrane123/rolepilot-backend/internal/models"
)

type MeetingHandler struct {
	DB *database.DB
}

func NewMeetingHandler(db *database.DB) *MeetingHandler {
	return &MeetingHandler{DB: db}
}

// Create handles POST /api/applications/{id}/meetings
func (h *MeetingHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	appID := chi.URLParam(r, "id")

	app, err := h.DB.GetApplication(r.Context(), appID, userID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, models.APIResponse{Error: "application not found"})
		return
	}

	var req models.CreateMeetingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Error: "invalid request body"})
		return
	}

	if req.LocationType == "" {
		req.LocationType = "video"
	}

	meeting, err := h.DB.CreateMeeting(r.Context(), appID, userID, app.CurrentStage, req)
	if err != nil {
		log.Println("CreateMeeting error:", err)
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{Error: "failed to create meeting"})
		return
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{Data: meeting})
}

// List handles GET /api/applications/{id}/meetings
func (h *MeetingHandler) List(w http.ResponseWriter, r *http.Request) {
	appID := chi.URLParam(r, "id")

	meetings, err := h.DB.GetMeetingsByApplication(r.Context(), appID)
	if err != nil {
		log.Println("ListMeetings error:", err)
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{Error: "failed to fetch meetings"})
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Data: meetings})
}

// Upcoming handles GET /api/meetings/upcoming
func (h *MeetingHandler) Upcoming(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	meetings, err := h.DB.GetUpcomingMeetings(r.Context(), userID)
	if err != nil {
		log.Println("UpcomingMeetings error:", err)
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{Error: "failed to fetch upcoming meetings"})
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Data: meetings})
}

// Update handles PATCH /api/meetings/{meetingId}
func (h *MeetingHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	meetingID := chi.URLParam(r, "meetingId")

	var req models.UpdateMeetingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Error: "invalid request body"})
		return
	}

	meeting, err := h.DB.UpdateMeeting(r.Context(), meetingID, userID, req)
	if err != nil {
		log.Println("UpdateMeeting error:", err)
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{Error: "failed to update meeting"})
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Data: meeting})
}

// Delete handles DELETE /api/meetings/{meetingId}
func (h *MeetingHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	meetingID := chi.URLParam(r, "meetingId")

	if err := h.DB.DeleteMeeting(r.Context(), meetingID, userID); err != nil {
		writeJSON(w, http.StatusNotFound, models.APIResponse{Error: "meeting not found"})
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Message: "meeting deleted"})
}