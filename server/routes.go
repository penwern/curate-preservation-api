package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/penwern/curate-preservation-core-api/database"
	"github.com/penwern/curate-preservation-core-api/models"
	"github.com/penwern/curate-preservation-core-api/pkg/logger"
)

// routes registers the API routes
func (s *Server) routes() {
	// API version prefix
	s.router.Route("/api/v1", func(r chi.Router) {
		// Health check
		r.Get("/health", s.handleHealth())

		// Preservation configurations
		r.Route("/preservation-configs", func(r chi.Router) {
			r.Get("/", s.handleListConfigs())
			r.Post("/", s.handleCreateConfig())

			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", s.handleGetConfig())
				r.Put("/", s.handleUpdateConfig())
				r.Delete("/", s.handleDeleteConfig())
			})
		})
	})
}

// handleHealth returns a health check handler
func (s *Server) handleHealth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		respondWithJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}

// handleListConfigs returns a handler to list all preservation configs
func (s *Server) handleListConfigs() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Fetching all preservation configs")
		configs, err := s.db.ListConfigs()
		if err != nil {
			logger.Error("Failed to fetch configs: %v", err)
			respondWithError(w, http.StatusInternalServerError, "Failed to fetch configs")
			return
		}

		logger.Debug("Successfully fetched %d configs", len(configs))
		respondWithJSON(w, http.StatusOK, configs)
	}
}

// handleGetConfig returns a handler to get a specific preservation config
func (s *Server) handleGetConfig() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		if idStr == "" {
			logger.Warn("Get config request missing ID parameter")
			respondWithError(w, http.StatusBadRequest, "ID is required")
			return
		}

		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			logger.Warn("Invalid ID format in get config request: %s", idStr)
			respondWithError(w, http.StatusBadRequest, "Invalid ID format")
			return
		}

		logger.Info("Fetching preservation config with ID: %d", id)
		config, err := s.db.GetConfig(id)
		if err != nil {
			if errors.Is(err, database.ErrNotFound) {
				logger.Warn("Preservation config not found: %d", id)
				respondWithError(w, http.StatusNotFound, "Preservation config not found")
				return
			}
			logger.Error("Failed to fetch config %d: %v", id, err)
			respondWithError(w, http.StatusInternalServerError, "Failed to fetch config")
			return
		}

		logger.Debug("Successfully fetched config: %s (ID: %d)", config.Name, config.ID)
		respondWithJSON(w, http.StatusOK, config)
	}
}

// handleCreateConfig returns a handler to create a new preservation config
func (s *Server) handleCreateConfig() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			Name        string `json:"name"`
			Description string `json:"description"`
		}

		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			logger.Warn("Invalid request payload in create config: %v", err)
			respondWithError(w, http.StatusBadRequest, "Invalid request payload")
			return
		}

		if input.Name == "" {
			logger.Warn("Create config request missing required name field")
			respondWithError(w, http.StatusBadRequest, "Name is required")
			return
		}

		logger.Info("Creating new preservation config: %s", input.Name)
		config := models.NewPreservationConfig(input.Name, input.Description)

		if err := s.db.CreateConfig(config); err != nil {
			logger.Error("Failed to create config '%s': %v", input.Name, err)
			respondWithError(w, http.StatusInternalServerError, "Failed to create config")
			return
		}

		logger.Info("Successfully created preservation config: %s (ID: %d)", config.Name, config.ID)
		respondWithJSON(w, http.StatusCreated, config)
	}
}

// handleUpdateConfig returns a handler to update an existing preservation config
func (s *Server) handleUpdateConfig() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		if idStr == "" {
			logger.Warn("Update config request missing ID parameter")
			respondWithError(w, http.StatusBadRequest, "ID is required")
			return
		}

		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			logger.Warn("Invalid ID format in update config request: %s", idStr)
			respondWithError(w, http.StatusBadRequest, "Invalid ID format")
			return
		}

		logger.Info("Updating preservation config with ID: %d", id)

		// Get the existing config to verify it exists
		existingConfig, err := s.db.GetConfig(id)
		if err != nil {
			if errors.Is(err, database.ErrNotFound) {
				logger.Warn("Attempted to update non-existent config: %d", id)
				respondWithError(w, http.StatusNotFound, "Preservation config not found")
				return
			}
			logger.Error("Failed to fetch existing config %d for update: %v", id, err)
			respondWithError(w, http.StatusInternalServerError, "Failed to fetch config")
			return
		}

		// Decode the updated config from request body
		var updatedConfig models.PreservationConfig
		if err := json.NewDecoder(r.Body).Decode(&updatedConfig); err != nil {
			logger.Warn("Invalid request payload in update config %d: %v", id, err)
			respondWithError(w, http.StatusBadRequest, "Invalid request payload")
			return
		}

		// Ensure the ID in the URL matches the ID in the request body (if provided)
		if updatedConfig.ID != 0 && updatedConfig.ID != id {
			logger.Warn("ID mismatch in update request: URL=%d, Body=%d", id, updatedConfig.ID)
			respondWithError(w, http.StatusBadRequest, "ID in URL does not match ID in request body")
			return
		}

		// Set the ID and preserve created time
		updatedConfig.ID = id
		updatedConfig.CreatedAt = existingConfig.CreatedAt

		if err := s.db.UpdateConfig(&updatedConfig); err != nil {
			logger.Error("Failed to update config %d: %v", id, err)
			respondWithError(w, http.StatusInternalServerError, "Failed to update config")
			return
		}

		logger.Info("Successfully updated preservation config: %s (ID: %d)", updatedConfig.Name, updatedConfig.ID)
		respondWithJSON(w, http.StatusOK, &updatedConfig)
	}
}

// handleDeleteConfig returns a handler to delete a preservation config
func (s *Server) handleDeleteConfig() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		if idStr == "" {
			logger.Warn("Delete config request missing ID parameter")
			respondWithError(w, http.StatusBadRequest, "ID is required")
			return
		}

		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			logger.Warn("Invalid ID format in delete config request: %s", idStr)
			respondWithError(w, http.StatusBadRequest, "Invalid ID format")
			return
		}

		logger.Info("Deleting preservation config with ID: %d", id)

		if err := s.db.DeleteConfig(id); err != nil {
			if errors.Is(err, database.ErrNotFound) {
				logger.Warn("Attempted to delete non-existent config: %d", id)
				respondWithError(w, http.StatusNotFound, "Preservation config not found")
				return
			}
			logger.Error("Failed to delete config %d: %v", id, err)
			respondWithError(w, http.StatusInternalServerError, "Failed to delete config")
			return
		}

		logger.Info("Successfully deleted preservation config with ID: %d", id)
		w.WriteHeader(http.StatusNoContent)
	}
}
