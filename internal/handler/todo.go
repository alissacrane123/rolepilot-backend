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

type TodoHandler struct {
	DB *database.DB
}

func NewTodoHandler(db *database.DB) *TodoHandler {
	return &TodoHandler{DB: db}
}

// ============================================
// GROUPS
// ============================================

// CreateGroup handles POST /api/todo-groups
func (h *TodoHandler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	var req models.CreateTodoGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Error: "invalid request body"})
		return
	}
	if req.Name == "" {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Error: "name is required"})
		return
	}
	group, err := h.DB.CreateTodoGroup(r.Context(), userID, req)
	if err != nil {
		log.Println("CreateTodoGroup error:", err)
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{Error: "failed to create group"})
		return
	}
	writeJSON(w, http.StatusCreated, models.APIResponse{Data: group})
}

// ListGroups handles GET /api/todo-groups
func (h *TodoHandler) ListGroups(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	groups, err := h.DB.GetTodoGroups(r.Context(), userID)
	if err != nil {
		log.Println("GetTodoGroups error:", err)
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{Error: "failed to fetch groups"})
		return
	}
	writeJSON(w, http.StatusOK, models.APIResponse{Data: groups})
}

// UpdateGroup handles PATCH /api/todo-groups/{groupId}
func (h *TodoHandler) UpdateGroup(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	groupID := chi.URLParam(r, "groupId")
	var req models.UpdateTodoGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Error: "invalid request body"})
		return
	}
	group, err := h.DB.UpdateTodoGroup(r.Context(), groupID, userID, req)
	if err != nil {
		log.Println("UpdateTodoGroup error:", err)
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{Error: "failed to update group"})
		return
	}
	writeJSON(w, http.StatusOK, models.APIResponse{Data: group})
}

// DeleteGroup handles DELETE /api/todo-groups/{groupId}
func (h *TodoHandler) DeleteGroup(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	groupID := chi.URLParam(r, "groupId")
	if err := h.DB.DeleteTodoGroup(r.Context(), groupID, userID); err != nil {
		writeJSON(w, http.StatusNotFound, models.APIResponse{Error: "group not found"})
		return
	}
	writeJSON(w, http.StatusOK, models.APIResponse{Message: "group deleted"})
}

// ============================================
// TODOS
// ============================================

// Create handles POST /api/todos
func (h *TodoHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	var req models.CreateTodoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Error: "invalid request body"})
		return
	}
	if req.Title == "" {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Error: "title is required"})
		return
	}
	todo, err := h.DB.CreateTodo(r.Context(), userID, req)
	if err != nil {
		log.Println("CreateTodo error:", err)
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{Error: "failed to create todo"})
		return
	}
	writeJSON(w, http.StatusCreated, models.APIResponse{Data: todo})
}

// List handles GET /api/todos
// Query params: completed, application_id, group_id, due_date, due_before, due_after, search
func (h *TodoHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	filter := database.TodoFilter{}

	if v := r.URL.Query().Get("completed"); v != "" {
		b := v == "true"
		filter.Completed = &b
	}
	if v := r.URL.Query().Get("application_id"); v != "" {
		filter.ApplicationID = &v
	}
	if v := r.URL.Query().Get("group_id"); v != "" {
		filter.GroupID = &v
	}
	if v := r.URL.Query().Get("due_date"); v != "" {
		filter.DueDate = &v
	}
	if v := r.URL.Query().Get("due_before"); v != "" {
		filter.DueBefore = &v
	}
	if v := r.URL.Query().Get("due_after"); v != "" {
		filter.DueAfter = &v
	}
	if v := r.URL.Query().Get("search"); v != "" {
		filter.Search = &v
	}
	if r.URL.Query().Get("include_reminders") == "true" {
		filter.IncludeReminders = true
	}

	todos, err := h.DB.GetTodos(r.Context(), userID, filter)
	if err != nil {
		log.Println("GetTodos error:", err)
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{Error: "failed to fetch todos"})
		return
	}
	writeJSON(w, http.StatusOK, models.APIResponse{Data: todos})
}

// GetForDate handles GET /api/todos/date/{date}
func (h *TodoHandler) GetForDate(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	date := chi.URLParam(r, "date")

	todos, err := h.DB.GetTodosForDate(r.Context(), userID, date)
	if err != nil {
		log.Println("GetTodosForDate error:", err)
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{Error: "failed to fetch todos"})
		return
	}
	writeJSON(w, http.StatusOK, models.APIResponse{Data: todos})
}

// Update handles PATCH /api/todos/{todoId}
func (h *TodoHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	todoID := chi.URLParam(r, "todoId")

	var req models.UpdateTodoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Error: "invalid request body"})
		return
	}

	todo, err := h.DB.UpdateTodo(r.Context(), todoID, userID, req)
	if err != nil {
		log.Println("UpdateTodo error:", err)
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{Error: "failed to update todo"})
		return
	}
	writeJSON(w, http.StatusOK, models.APIResponse{Data: todo})
}

// Toggle handles PATCH /api/todos/{todoId}/toggle
func (h *TodoHandler) Toggle(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	todoID := chi.URLParam(r, "todoId")

	// Get current state
	todo, err := h.DB.GetTodo(r.Context(), todoID, userID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, models.APIResponse{Error: "todo not found"})
		return
	}

	newCompleted := !todo.Completed
	updated, err := h.DB.UpdateTodo(r.Context(), todoID, userID, models.UpdateTodoRequest{
		Completed: &newCompleted,
	})
	if err != nil {
		log.Println("ToggleTodo error:", err)
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{Error: "failed to toggle todo"})
		return
	}
	writeJSON(w, http.StatusOK, models.APIResponse{Data: updated})
}

// Delete handles DELETE /api/todos/{todoId}
func (h *TodoHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	todoID := chi.URLParam(r, "todoId")
	if err := h.DB.DeleteTodo(r.Context(), todoID, userID); err != nil {
		writeJSON(w, http.StatusNotFound, models.APIResponse{Error: "todo not found"})
		return
	}
	writeJSON(w, http.StatusOK, models.APIResponse{Message: "todo deleted"})
}