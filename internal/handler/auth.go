package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/alissacrane123/rolepilot-backend/internal/database"
	"github.com/alissacrane123/rolepilot-backend/internal/middleware"
	"github.com/alissacrane123/rolepilot-backend/internal/models"

	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	DB *database.DB
}

func NewAuthHandler(db *database.DB) *AuthHandler {
	return &AuthHandler{DB: db}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Error: "invalid request body"})
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.FullName = strings.TrimSpace(req.FullName)

	if req.Email == "" || req.Password == "" || req.FullName == "" {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Error: "email, password, and full_name are required"})
		return
	}

	if len(req.Password) < 8 {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Error: "password must be at least 8 characters"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{Error: "failed to hash password"})
		return
	}

	user, err := h.DB.CreateUser(r.Context(), req.Email, string(hash), req.FullName)
	if err != nil {
		log.Println("CreateUser error:", err)
		if strings.Contains(err.Error(), "duplicate key") {
			writeJSON(w, http.StatusConflict, models.APIResponse{Error: "email already registered"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{Error: "failed to create user"})
		return
	}

	token, err := middleware.GenerateToken(user.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{Error: "failed to generate token"})
		return
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{
		Data: models.AuthResponse{Token: token, User: *user},
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Error: "invalid request body"})
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	user, err := h.DB.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, models.APIResponse{Error: "invalid email or password"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		writeJSON(w, http.StatusUnauthorized, models.APIResponse{Error: "invalid email or password"})
		return
	}

	token, err := middleware.GenerateToken(user.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{Error: "failed to generate token"})
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Data: models.AuthResponse{Token: token, User: *user},
	})
}

func (h *AuthHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	user, err := h.DB.GetUserByID(r.Context(), userID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, models.APIResponse{Error: "user not found"})
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Data: user})
}

func (h *AuthHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	var req models.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Error: "invalid request body"})
		return
	}

	user, err := h.DB.UpdateUserProfile(r.Context(), userID, req)
	if err != nil {
		log.Println("UpdateProfile error:", err)
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{Error: "failed to update profile"})
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Data: user})
}

// UploadResume handles POST /api/profile/resume
// Accepts multipart form with a "resume" file field
func (h *AuthHandler) UploadResume(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	// Limit upload to 10MB
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Error: "file too large (max 10MB)"})
		return
	}

	file, header, err := r.FormFile("resume")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Error: "resume file is required"})
		return
	}
	defer file.Close()

	// Validate file type
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".pdf" && ext != ".txt" && ext != ".md" {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Error: "only .pdf, .txt, and .md files are supported"})
		return
	}

	// Create uploads directory
	uploadDir := "./uploads/resumes"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{Error: "failed to create upload directory"})
		return
	}

	// Save file with unique name
	filename := fmt.Sprintf("%s_%d%s", userID, time.Now().Unix(), ext)
	filePath := filepath.Join(uploadDir, filename)

	dst, err := os.Create(filePath)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{Error: "failed to save file"})
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{Error: "failed to save file"})
		return
	}

	// Extract text based on file type
	var resumeText string
	switch ext {
	case ".txt", ".md":
		textBytes, err := os.ReadFile(filePath)
		if err != nil {
			log.Println("Read text file error:", err)
			writeJSON(w, http.StatusInternalServerError, models.APIResponse{Error: "failed to read file"})
			return
		}
		resumeText = string(textBytes)
	case ".pdf":
		text, err := extractPDFText(filePath)
		if err != nil {
			log.Println("PDF extraction error:", err)
			// Still save the file even if text extraction fails
			resumeText = ""
		} else {
			resumeText = text
		}
	}

	// Save to database
	if err := h.DB.UpdateUserResume(r.Context(), userID, filePath, resumeText); err != nil {
		log.Println("UpdateUserResume error:", err)
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{Error: "failed to update resume"})
		return
	}

	user, err := h.DB.GetUserByID(r.Context(), userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{Error: "failed to fetch user"})
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Data:    user,
		Message: "resume uploaded successfully",
	})
}

// UploadResumeText handles POST /api/profile/resume/text
// Accepts JSON with raw resume text (for copy-paste)
func (h *AuthHandler) UploadResumeText(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	var req struct {
		ResumeText string `json:"resume_text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Error: "invalid request body"})
		return
	}

	if strings.TrimSpace(req.ResumeText) == "" {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Error: "resume_text is required"})
		return
	}

	if err := h.DB.UpdateUserResume(r.Context(), userID, "", req.ResumeText); err != nil {
		log.Println("UpdateUserResume error:", err)
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{Error: "failed to update resume"})
		return
	}

	user, err := h.DB.GetUserByID(r.Context(), userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{Error: "failed to fetch user"})
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Data:    user,
		Message: "resume text saved successfully",
	})
}