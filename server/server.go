// Package server provides HTTP server functionality for the preservation API.
package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/render"
	"github.com/penwern/curate-preservation-api/database"
	"github.com/penwern/curate-preservation-api/pkg/config"
	"github.com/penwern/curate-preservation-api/pkg/logger"
)

// Server represents the API server
type Server struct {
	router *chi.Mux
	db     *database.Database
	srv    *http.Server
	config config.Config
}

// New creates a new server
func New(cfg config.Config) (*Server, error) {
	db, err := database.New(cfg.DBType, cfg.DBConnection)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	router := chi.NewRouter()

	// CORS middleware - configure to allow requests from Pydio Cells
	corsOrigins := cfg.CORSOrigins
	if len(corsOrigins) == 0 {
		// Default origins if none specified in config
		corsOrigins = []string{
			"https://localhost:8080",
			"http://localhost:8080",
		}
	}

	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   corsOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	// Middleware
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Timeout(5 * time.Second))
	router.Use(render.SetContentType(render.ContentTypeJSON))

	server := &Server{
		router: router,
		db:     db,
		srv: &http.Server{
			Addr:              fmt.Sprintf(":%d", cfg.Port),
			Handler:           router,
			ReadHeaderTimeout: 15 * time.Second,
		},
		config: cfg,
	}

	// Register routes
	server.routes()

	return server, nil
}

// Start starts the HTTP server
func (s *Server) Start() error {
	return s.srv.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown() error {
	// Close the database connection
	if err := s.db.Close(); err != nil {
		logger.Error("Error closing database: %v", err)
	}

	// Create a deadline to wait for current connections to complete
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Shutdown the server
	return s.srv.Shutdown(ctx)
}

// respondWithJSON writes a JSON response
func respondWithJSON(w http.ResponseWriter, code int, payload any) {
	b, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if _, err := w.Write(b); err != nil {
		logger.Error("Failed to write response: %v", err)
	}
}

// respondWithError writes an error response
func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}
