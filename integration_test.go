package main

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/penwern/curate-preservation-api/database"
	"github.com/penwern/curate-preservation-api/models"
	"github.com/penwern/curate-preservation-api/pkg/config"
	"github.com/penwern/curate-preservation-api/pkg/logger"
	"github.com/penwern/curate-preservation-api/server"
)

// TestIntegrationDatabaseOperations tests database operations independently
func TestIntegrationDatabaseOperations(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "integration_test.db")
	logPath := filepath.Join(tmpDir, "integration_test.log")

	// Initialize logger
	logger.Initialize("debug", logPath)

	// Test database initialization
	db, err := database.New("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Test basic CRUD operations
	t.Run("Create Config", func(t *testing.T) {
		config := models.NewPreservationConfig("Integration Test Config", "Test Description")

		err := db.CreateConfig(config)
		if err != nil {
			t.Fatalf("Failed to create config: %v", err)
		}

		if config.ID == 0 {
			t.Error("Config should have non-zero ID after creation")
		}
	})

	t.Run("List Configs", func(t *testing.T) {
		configs, err := db.ListConfigs()
		if err != nil {
			t.Fatalf("Failed to list configs: %v", err)
		}

		// Should have default config + the one we created
		if len(configs) < 2 {
			t.Errorf("Expected at least 2 configs, got %d", len(configs))
		}
	})

	t.Run("Complex A3M Config", func(t *testing.T) {
		config := &models.PreservationConfig{
			Name:        "Complex A3M Config",
			Description: "Testing complex A3M settings",
			CompressAIP: true,
			A3MConfig: models.A3MProcessingConfig{
				AssignUuidsToDirectories: false,
				ExamineContents:          true,
				AipCompressionLevel:      9,
			},
		}

		err := db.CreateConfig(config)
		if err != nil {
			t.Fatalf("Failed to create complex config: %v", err)
		}

		// Retrieve and verify
		retrieved, err := db.GetConfig(config.ID)
		if err != nil {
			t.Fatalf("Failed to retrieve config: %v", err)
		}

		if retrieved.CompressAIP != true {
			t.Error("CompressAIP not preserved")
		}

		if retrieved.A3MConfig.ExamineContents != true {
			t.Error("A3M ExamineContents not preserved")
		}

		if retrieved.A3MConfig.AipCompressionLevel != 9 {
			t.Errorf("A3M AipCompressionLevel not preserved: expected 9, got %d", retrieved.A3MConfig.AipCompressionLevel)
		}
	})

	t.Run("Update and Delete", func(t *testing.T) {
		// Create
		config := models.NewPreservationConfig("To Update and Delete", "Original description")
		err := db.CreateConfig(config)
		if err != nil {
			t.Fatalf("Failed to create config: %v", err)
		}

		originalID := config.ID

		// Update
		config.Description = "Updated description"
		config.A3MConfig.ExamineContents = true
		err = db.UpdateConfig(config)
		if err != nil {
			t.Fatalf("Failed to update config: %v", err)
		}

		// Verify update
		updated, err := db.GetConfig(originalID)
		if err != nil {
			t.Fatalf("Failed to get updated config: %v", err)
		}

		if updated.Description != "Updated description" {
			t.Error("Description was not updated")
		}

		if !updated.A3MConfig.ExamineContents {
			t.Error("A3M setting was not updated")
		}

		// Delete
		err = db.DeleteConfig(originalID)
		if err != nil {
			t.Fatalf("Failed to delete config: %v", err)
		}

		// Verify deletion
		_, err = db.GetConfig(originalID)
		if err != database.ErrNotFound {
			t.Error("Config should be deleted")
		}
	})
}

// TestIntegrationServerInitialization tests server setup and basic functionality
func TestIntegrationServerInitialization(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "server_test.db")

	cfg := config.Config{
		DBType:       "sqlite3",
		DBConnection: dbPath,
		Port:         8080,
		TrustedIPs:   []string{"127.0.0.1"},
		CORSOrigins:  []string{"http://localhost:3000"},
	}

	// Test server creation
	srv, err := server.New(cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}
	defer srv.Shutdown()

	// Test that we can shut down and restart
	err = srv.Shutdown()
	if err != nil {
		t.Fatalf("Failed to shutdown server: %v", err)
	}

	// Create another server with same config
	srv2, err := server.New(cfg)
	if err != nil {
		t.Fatalf("Failed to create second server: %v", err)
	}
	defer srv2.Shutdown()
}

// TestIntegrationDataPersistence tests that data persists across database connections
func TestIntegrationDataPersistence(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "persistence_test.db")

	// First connection
	db1, err := database.New("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Failed to create first database connection: %v", err)
	}

	// Create a config
	config := models.NewPreservationConfig("Persistence Test", "Should persist across connections")
	config.A3MConfig.ExamineContents = true
	config.A3MConfig.AipCompressionLevel = 7

	err = db1.CreateConfig(config)
	if err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}

	configID := config.ID
	db1.Close()

	// Second connection
	db2, err := database.New("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Failed to create second database connection: %v", err)
	}
	defer db2.Close()

	// Retrieve the config
	retrieved, err := db2.GetConfig(configID)
	if err != nil {
		t.Fatalf("Failed to retrieve persisted config: %v", err)
	}

	// Verify all fields are preserved
	if retrieved.Name != "Persistence Test" {
		t.Errorf("Name not preserved: expected 'Persistence Test', got '%s'", retrieved.Name)
	}

	if !retrieved.A3MConfig.ExamineContents {
		t.Error("A3M ExamineContents not preserved")
	}

	if retrieved.A3MConfig.AipCompressionLevel != 7 {
		t.Errorf("A3M compression level not preserved: expected 7, got %d", retrieved.A3MConfig.AipCompressionLevel)
	}
}

// TestIntegrationJSONSerialization tests JSON marshaling/unmarshaling across the system
func TestIntegrationJSONSerialization(t *testing.T) {
	// Test complex configuration JSON serialization
	config := &models.PreservationConfig{
		ID:          42,
		Name:        "JSON Test Config",
		Description: "Testing JSON serialization with unicode: 🚀 العربية русский 中文",
		CompressAIP: true,
		A3MConfig:   models.NewA3MProcessingConfig(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Customize A3M config
	config.A3MConfig.ExamineContents = true
	config.A3MConfig.AipCompressionLevel = 9

	// Marshal to JSON
	jsonData, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("Failed to marshal config to JSON: %v", err)
	}

	// Unmarshal from JSON
	var unmarshaled models.PreservationConfig
	err = json.Unmarshal(jsonData, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal config from JSON: %v", err)
	}

	// Verify preservation
	if unmarshaled.Name != config.Name {
		t.Error("Name not preserved in JSON round-trip")
	}

	if unmarshaled.Description != config.Description {
		t.Error("Unicode description not preserved in JSON round-trip")
	}

	if unmarshaled.CompressAIP != config.CompressAIP {
		t.Error("CompressAIP not preserved in JSON round-trip")
	}

	if unmarshaled.A3MConfig.ExamineContents != config.A3MConfig.ExamineContents {
		t.Error("A3M ExamineContents not preserved in JSON round-trip")
	}

	if unmarshaled.A3MConfig.AipCompressionLevel != config.A3MConfig.AipCompressionLevel {
		t.Error("A3M AipCompressionLevel not preserved in JSON round-trip")
	}
}

// TestIntegrationErrorHandling tests error conditions across the system
func TestIntegrationErrorHandling(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "error_test.db")

	db, err := database.New("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	t.Run("Database Errors", func(t *testing.T) {
		// Test getting non-existent config
		_, err := db.GetConfig(99999)
		if err != database.ErrNotFound {
			t.Errorf("Expected ErrNotFound for non-existent config, got %v", err)
		}

		// Test updating non-existent config
		config := &models.PreservationConfig{
			ID:   99999,
			Name: "Non-existent",
		}
		err = db.UpdateConfig(config)
		if err == nil {
			t.Error("Expected error when updating non-existent config")
		}

		// Test deleting non-existent config
		err = db.DeleteConfig(99999)
		if err == nil {
			t.Error("Expected error when deleting non-existent config")
		}
	})

	t.Run("Invalid Database Configuration", func(t *testing.T) {
		// Test with invalid database type
		_, err := database.New("invalid_type", "connection")
		if err == nil {
			t.Error("Expected error for invalid database type")
		}

		// Test with invalid connection string
		_, err = database.New("sqlite3", "/invalid/path/that/does/not/exist/test.db")
		if err == nil {
			t.Error("Expected error for invalid connection string")
		}
	})

	t.Run("Server Configuration Errors", func(t *testing.T) {
		// Test server with invalid database config
		cfg := config.Config{
			DBType:       "invalid",
			DBConnection: "invalid",
			Port:         8080,
		}

		_, err := server.New(cfg)
		if err == nil {
			t.Error("Expected error when creating server with invalid database config")
		}
	})
}

// TestIntegrationConcurrency tests concurrent operations
func TestIntegrationConcurrency(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "concurrency_test.db")

	db, err := database.New("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Test concurrent config creation
	t.Run("Concurrent Creation", func(t *testing.T) {
		numGoroutines := 10
		done := make(chan bool, numGoroutines)
		errors := make(chan error, numGoroutines)

		for i := range numGoroutines {
			go func(id int) {
				defer func() { done <- true }()

				config := models.NewPreservationConfig(
					fmt.Sprintf("Concurrent Config %d", id),
					fmt.Sprintf("Created by goroutine %d", id),
				)

				err := db.CreateConfig(config)
				if err != nil {
					errors <- err
				}
			}(i)
		}

		// Wait for all goroutines
		for range numGoroutines {
			<-done
		}

		// Check for errors
		select {
		case err := <-errors:
			t.Errorf("Concurrent creation failed: %v", err)
		default:
			// No errors
		}

		// Verify all configs were created
		configs, err := db.ListConfigs()
		if err != nil {
			t.Fatalf("Failed to list configs after concurrent creation: %v", err)
		}

		// Should have default config + numGoroutines configs
		expectedCount := 1 + numGoroutines
		if len(configs) < expectedCount {
			t.Errorf("Expected at least %d configs after concurrent creation, got %d", expectedCount, len(configs))
		}
	})
}

// BenchmarkIntegrationDatabaseOperations benchmarks database operations
func BenchmarkIntegrationDatabaseOperations(b *testing.B) {
	tmpDir := b.TempDir()
	dbPath := tmpDir + "/benchmark.db"

	db, err := database.New("sqlite3", dbPath)
	if err != nil {
		b.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	b.Run("CreateConfig", func(b *testing.B) {
		for i := 0; b.Loop(); i++ {
			config := models.NewPreservationConfig(
				fmt.Sprintf("Benchmark Config %d", i),
				"Benchmark description",
			)

			err := db.CreateConfig(config)
			if err != nil {
				b.Fatalf("Failed to create config: %v", err)
			}
		}
	})

	// Pre-create some configs for other benchmarks
	for i := range 100 {
		config := models.NewPreservationConfig(
			fmt.Sprintf("Pre-created Config %d", i),
			"Pre-created for benchmarks",
		)
		db.CreateConfig(config)
	}

	b.Run("ListConfigs", func(b *testing.B) {
		for b.Loop() {
			_, err := db.ListConfigs()
			if err != nil {
				b.Fatalf("Failed to list configs: %v", err)
			}
		}
	})

	b.Run("GetConfig", func(b *testing.B) {
		configID := int64(2) // Assuming this config exists
		for b.Loop() {
			_, err := db.GetConfig(configID)
			if err != nil {
				b.Fatalf("Failed to get config: %v", err)
			}
		}
	})
}
