package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"

	"github.com/alissacrane123/rolepilot-backend/internal/database"
	"github.com/alissacrane123/rolepilot-backend/internal/handler"
	"github.com/alissacrane123/rolepilot-backend/internal/middleware"
	"github.com/alissacrane123/rolepilot-backend/internal/services"
)

func main() {
	godotenv.Load()

	db, err := database.New(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()
	fmt.Println("Connected to database!")

	aiService := services.NewAIService()
	authHandler := handler.NewAuthHandler(db)
	appHandler := handler.NewApplicationHandler(db, aiService)
	meetingHandler := handler.NewMeetingHandler(db)
	noteHandler := handler.NewNoteHandler(db)

	r := chi.NewRouter()

	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
	}))

	r.Get("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Public routes
	r.Post("/api/auth/register", authHandler.Register)
	r.Post("/api/auth/login", authHandler.Login)

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth)

		// Profile
		r.Get("/api/profile", authHandler.GetProfile)
		r.Patch("/api/profile", authHandler.UpdateProfile)
		r.Post("/api/profile/resume", authHandler.UploadResume)
		r.Post("/api/profile/resume/text", authHandler.UploadResumeText)

		// Applications
		r.Post("/api/applications", appHandler.Create)
		r.Get("/api/applications", appHandler.List)
		r.Get("/api/applications/board", appHandler.Board)
		r.Get("/api/applications/{id}", appHandler.Get)
		r.Patch("/api/applications/{id}/stage", appHandler.UpdateStage)
		r.Get("/api/applications/{id}/history", appHandler.GetStageHistory)
		r.Post("/api/applications/{id}/reanalyze", appHandler.Reanalyze)
		r.Post("/api/applications/{id}/cover-letter", appHandler.GenerateCoverLetter)
		r.Get("/api/applications/{id}/cover-letters", appHandler.GetCoverLetters)

		// Meetings (per application)
		r.Post("/api/applications/{id}/meetings", meetingHandler.Create)
		r.Get("/api/applications/{id}/meetings", meetingHandler.List)

		// Meetings (global)
		r.Get("/api/meetings/upcoming", meetingHandler.Upcoming)
		r.Patch("/api/meetings/{meetingId}", meetingHandler.Update)
		r.Delete("/api/meetings/{meetingId}", meetingHandler.Delete)

		// Notes (per application)
		r.Post("/api/applications/{id}/notes", noteHandler.Create)
		r.Get("/api/applications/{id}/notes", noteHandler.List)

		// Notes (global)
		r.Patch("/api/notes/{noteId}", noteHandler.Update)
		r.Delete("/api/notes/{noteId}", noteHandler.Delete)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Printf("Server running on :%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
