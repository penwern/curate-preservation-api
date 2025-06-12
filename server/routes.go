package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/mitchellh/mapstructure"
	"github.com/penwern/curate-preservation-api/database"
	"github.com/penwern/curate-preservation-api/models"
	"github.com/penwern/curate-preservation-api/pkg/logger"
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

		logger.Debug("Config: %+v", config)
	}
}

// handleCreateConfig returns a handler to create a new preservation config
func (s *Server) handleCreateConfig() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse the raw JSON to detect which fields are provided
		var rawInput map[string]any
		if err := json.NewDecoder(r.Body).Decode(&rawInput); err != nil {
			logger.Warn("Invalid request payload in create config: %v", err)
			respondWithError(w, http.StatusBadRequest, "Invalid request payload")
			return
		}

		logger.Debug("Raw input: %v", rawInput)

		// Extract name (required)
		name, nameExists := rawInput["name"]
		if !nameExists {
			logger.Warn("Create config request missing required name field")
			respondWithError(w, http.StatusBadRequest, "Name is required")
			return
		}
		nameStr, ok := name.(string)
		if !ok || nameStr == "" {
			logger.Warn("Create config request has invalid name field")
			respondWithError(w, http.StatusBadRequest, "Name is required and must be a string")
			return
		}

		// Extract description (optional)
		description := ""
		if desc, exists := rawInput["description"]; exists {
			if descStr, ok := desc.(string); ok {
				description = descStr
			}
		}

		logger.Info("Creating new preservation config: %s", nameStr)

		// Start with default config
		config := models.NewPreservationConfig(nameStr, description)

		logger.Debug("Default Config: %+v", config)

		// If A3M config is provided, merge it with defaults
		if a3mConfig, exists := rawInput["a3m_config"]; exists {
			if a3mMap, ok := a3mConfig.(map[string]any); ok {
				updateA3MConfigFromMap(&config.A3MConfig, a3mMap)
			}
		}

		logger.Debug("Updated Config: %+v", config)

		if err := s.db.CreateConfig(config); err != nil {
			logger.Error("Failed to create config '%s': %v", nameStr, err)
			respondWithError(w, http.StatusInternalServerError, "Failed to create config")
			return
		}

		// Fetch the created config from the database to ensure we return the actual saved data
		createdConfig, err := s.db.GetConfig(config.ID)
		if err != nil {
			logger.Error("Failed to fetch created config %d: %v", config.ID, err)
			respondWithError(w, http.StatusInternalServerError, "Failed to fetch created config")
			return
		}

		logger.Debug("Created Config: %+v", createdConfig)

		logger.Info("Successfully created preservation config: %s (ID: %d)", createdConfig.Name, createdConfig.ID)
		respondWithJSON(w, http.StatusCreated, createdConfig)
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

		// Parse the raw JSON to detect which fields are provided
		var rawUpdate map[string]any
		if err := json.NewDecoder(r.Body).Decode(&rawUpdate); err != nil {
			logger.Warn("Invalid request payload in update config %d: %v", id, err)
			respondWithError(w, http.StatusBadRequest, "Invalid request payload")
			return
		}

		// Work with the existing config directly (avoid copying)
		updatedConfig := existingConfig

		// Update basic fields if provided
		if name, exists := rawUpdate["name"]; exists {
			if nameStr, ok := name.(string); ok {
				updatedConfig.Name = nameStr
			}
		}
		if description, exists := rawUpdate["description"]; exists {
			if descStr, ok := description.(string); ok {
				updatedConfig.Description = descStr
			}
		}

		// Handle A3M config updates if provided
		if a3mConfig, exists := rawUpdate["a3m_config"]; exists {
			if a3mMap, ok := a3mConfig.(map[string]any); ok {
				updateA3MConfigFromMap(&updatedConfig.A3MConfig, a3mMap)
			}
		}

		// Ensure the ID in the URL matches the ID in the request body (if provided)
		if idFromBody, exists := rawUpdate["id"]; exists {
			if idFloat, ok := idFromBody.(float64); ok && int64(idFloat) != id {
				logger.Warn("ID mismatch in update request: URL=%d, Body=%d", id, int64(idFloat))
				respondWithError(w, http.StatusBadRequest, "ID in URL does not match ID in request body")
				return
			}
		}

		// Set the ID (already correct, but ensure it's set)
		updatedConfig.ID = id

		if err := s.db.UpdateConfig(updatedConfig); err != nil {
			logger.Error("Failed to update config %d: %v", id, err)
			respondWithError(w, http.StatusInternalServerError, "Failed to update config")
			return
		}

		logger.Info("Successfully updated preservation config: %s (ID: %d)", updatedConfig.Name, updatedConfig.ID)
		respondWithJSON(w, http.StatusOK, updatedConfig)
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

func updateA3MConfigFromMap(target *models.A3MProcessingConfig, source map[string]any) {
	config := &mapstructure.DecoderConfig{
		Result:           target,
		WeaklyTypedInput: true, // Handles float64 -> int32 conversion
		TagName:          "json",
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		logger.Error("Failed to create decoder: %v", err)
		return
	}

	if err := decoder.Decode(source); err != nil {
		logger.Error("Failed to decode config: %v", err)
	}
}
