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

type NoteHandler struct {
	DB *database.DB
}

func NewNoteHandler(db *database.DB) *NoteHandler {
	return &NoteHandler{DB: db}
}

// Create handles POST /api/applications/{id}/notes
func (h *NoteHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	appID := chi.URLParam(r, "id")

	var req models.CreateNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Error: "invalid request body"})
		return
	}

	note, err := h.DB.CreateNote(r.Context(), appID, userID, req)
	if err != nil {
		log.Println("CreateNote error:", err)
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{Error: "failed to create note"})
		return
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{Data: note})
}

// List handles GET /api/applications/{id}/notes
func (h *NoteHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	appID := chi.URLParam(r, "id")

	notes, err := h.DB.GetNotesByApplication(r.Context(), appID, userID)
	if err != nil {
		log.Println("ListNotes error:", err)
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{Error: "failed to fetch notes"})
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Data: notes})
}

// Update handles PATCH /api/notes/{noteId}
func (h *NoteHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	noteID := chi.URLParam(r, "noteId")

	var req models.UpdateNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Error: "invalid request body"})
		return
	}

	note, err := h.DB.UpdateNote(r.Context(), noteID, userID, req)
	if err != nil {
		log.Println("UpdateNote error:", err)
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{Error: "failed to update note"})
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Data: note})
}

// Delete handles DELETE /api/notes/{noteId}
func (h *NoteHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	noteID := chi.URLParam(r, "noteId")

	if err := h.DB.DeleteNote(r.Context(), noteID, userID); err != nil {
		writeJSON(w, http.StatusNotFound, models.APIResponse{Error: "note not found"})
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Message: "note deleted"})
}