package handler

import (
	"context"
	"encoding/json"
	"net/http"
)

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// r_context returns a background context for goroutines
// (since the HTTP request context gets cancelled after response)
func r_context() context.Context {
	return context.Background()
}
