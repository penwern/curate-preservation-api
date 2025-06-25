package database

import (
	"path/filepath"
	"testing"
	"time"

	transferservice "github.com/penwern/curate-preservation-api/common/proto/a3m/gen/go/a3m/api/transferservice/v1beta1"
	"github.com/penwern/curate-preservation-api/models"
	"github.com/penwern/curate-preservation-api/pkg/logger"
)

const (
	testDBType       = "sqlite3"
	testOriginalName = "Original Name"
	testOriginalDesc = "Original Description"
)

func setupTestDB(t *testing.T) *Database {
	t.Helper()

	logger.Initialize("debug", "/tmp/curate-preservation-api.log")

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := New(testDBType, dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	return db
}

func TestNew_SQLite(t *testing.T) {
	logger.Initialize("debug", "/tmp/curate-preservation-api.log")

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := New(testDBType, dbPath)
	if err != nil {
		t.Fatalf("Failed to create SQLite database: %v", err)
	}
	defer db.Close()

	if db.dbType != testDBType {
		t.Errorf("Expected dbType '%s', got '%s'", testDBType, db.dbType)
	}
}

func TestNew_UnsupportedDBType(t *testing.T) {
	_, err := New("postgres", "connection-string")
	if err == nil {
		t.Error("Expected error for unsupported database type, got nil")
	}

	expectedError := "unsupported database type, must be 'sqlite3' or 'mysql'"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestNew_InvalidConnectionString(t *testing.T) {
	_, err := New(testDBType, "/invalid/path/that/does/not/exist/test.db")
	if err == nil {
		t.Error("Expected error for invalid connection string, got nil")
	}
}

func TestDatabase_CreateAndGetConfig(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create a test config
	config := models.NewPreservationConfig("Test Config", "Test Description")

	// Test CreateConfig
	err := db.CreateConfig(config)
	if err != nil {
		t.Fatalf("CreateConfig failed: %v", err)
	}

	if config.ID == 0 {
		t.Error("Expected config ID to be set after creation")
	}

	// Test GetConfig
	retrievedConfig, err := db.GetConfig(config.ID)
	if err != nil {
		t.Fatalf("GetConfig failed: %v", err)
	}

	if retrievedConfig.Name != config.Name {
		t.Errorf("Expected name '%s', got '%s'", config.Name, retrievedConfig.Name)
	}

	if retrievedConfig.Description != config.Description {
		t.Errorf("Expected description '%s', got '%s'", config.Description, retrievedConfig.Description)
	}

	// Check some A3M config values
	if retrievedConfig.A3MConfig.AssignUuidsToDirectories != config.A3MConfig.AssignUuidsToDirectories {
		t.Errorf("AssignUuidsToDirectories mismatch: expected %v, got %v",
			config.A3MConfig.AssignUuidsToDirectories, retrievedConfig.A3MConfig.AssignUuidsToDirectories)
	}
}

func TestDatabase_GetConfig_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	_, err := db.GetConfig(999)
	if err == nil {
		t.Error("Expected error for non-existent config, got nil")
	}

	if err != ErrNotFound {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}

func TestDatabase_ListConfigs(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Initially should be empty
	configs, err := db.ListConfigs()
	if err != nil {
		t.Fatalf("ListConfigs failed: %v", err)
	}

	if len(configs) != 1 {
		t.Errorf("Expected 1 config, got %d", len(configs))
	}

	// Create a few test configs
	config1 := models.NewPreservationConfig("Config 1", "Description 1")
	config2 := models.NewPreservationConfig("Config 2", "Description 2")

	err = db.CreateConfig(config1)
	if err != nil {
		t.Fatalf("Failed to create config1: %v", err)
	}

	err = db.CreateConfig(config2)
	if err != nil {
		t.Fatalf("Failed to create config2: %v", err)
	}

	// List configs again
	configs, err = db.ListConfigs()
	if err != nil {
		t.Fatalf("ListConfigs failed: %v", err)
	}

	if len(configs) != 3 {
		t.Errorf("Expected 3 configs, got %d", len(configs))
	}
}

func TestDatabase_UpdateConfig(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create initial config
	config := models.NewPreservationConfig(testOriginalName, testOriginalDesc)
	err := db.CreateConfig(config)
	if err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}

	// Modify the config
	config.Name = "Updated Name"
	config.Description = "Updated Description"
	config.A3MConfig.ExamineContents = true // Change a boolean value

	// Update the config
	err = db.UpdateConfig(config)
	if err != nil {
		t.Fatalf("UpdateConfig failed: %v", err)
	}

	// Retrieve and verify the update
	updatedConfig, err := db.GetConfig(config.ID)
	if err != nil {
		t.Fatalf("Failed to get updated config: %v", err)
	}

	if updatedConfig.Name != "Updated Name" {
		t.Errorf("Expected name 'Updated Name', got '%s'", updatedConfig.Name)
	}

	if updatedConfig.Description != "Updated Description" {
		t.Errorf("Expected description 'Updated Description', got '%s'", updatedConfig.Description)
	}

	if updatedConfig.A3MConfig.ExamineContents != true {
		t.Error("Expected ExamineContents to be true")
	}
}

func TestDatabase_UpdateConfig_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	config := &models.PreservationConfig{
		ID:          999,
		Name:        "Non-existent",
		Description: "This config doesn't exist",
	}

	err := db.UpdateConfig(config)
	if err == nil {
		t.Error("Expected error for updating non-existent config, got nil")
	}
}

func TestDatabase_DeleteConfig(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create a config to delete
	config := models.NewPreservationConfig("To Delete", "Will be deleted")
	err := db.CreateConfig(config)
	if err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}

	// Delete the config
	err = db.DeleteConfig(config.ID)
	if err != nil {
		t.Fatalf("DeleteConfig failed: %v", err)
	}

	// Verify it's gone
	_, err = db.GetConfig(config.ID)
	if err != ErrNotFound {
		t.Errorf("Expected ErrNotFound after deletion, got %v", err)
	}
}

func TestDatabase_DeleteConfig_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	err := db.DeleteConfig(999)
	if err == nil {
		t.Error("Expected error for deleting non-existent config, got nil")
	}
}

func TestDatabase_Close(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := New(testDBType, dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	err = db.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}

	// After closing, operations should fail
	_, err = db.ListConfigs()
	if err == nil {
		t.Error("Expected error after closing database, got nil")
	}
}

func TestDatabase_ConfigWithCustomA3MValues(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create config with custom A3M values
	config := &models.PreservationConfig{
		Name:        "Custom Config",
		Description: "Config with custom A3M values",
		A3MConfig: models.A3MProcessingConfig{
			AssignUuidsToDirectories:                     false,
			ExamineContents:                              true,
			GenerateTransferStructureReport:              false,
			DocumentEmptyDirectories:                     false,
			ExtractPackages:                              false,
			DeletePackagesAfterExtraction:                true,
			IdentifyTransfer:                             false,
			IdentifySubmissionAndMetadata:                false,
			IdentifyBeforeNormalization:                  false,
			Normalize:                                    false,
			TranscribeFiles:                              false,
			PerformPolicyChecksOnOriginals:               false,
			PerformPolicyChecksOnPreservationDerivatives: false,
			PerformPolicyChecksOnAccessDerivatives:       false,
			ThumbnailMode:                                transferservice.ProcessingConfig_THUMBNAIL_MODE_DO_NOT_GENERATE,
			AipCompressionLevel:                          9,
			AipCompressionAlgorithm:                      transferservice.ProcessingConfig_AIP_COMPRESSION_ALGORITHM_S7_BZIP2,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := db.CreateConfig(config)
	if err != nil {
		t.Fatalf("Failed to create config with custom values: %v", err)
	}

	retrievedConfig, err := db.GetConfig(config.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve config: %v", err)
	}

	// Verify custom values are preserved
	if retrievedConfig.A3MConfig.AssignUuidsToDirectories != false {
		t.Error("Expected AssignUuidsToDirectories to be false")
	}

	if retrievedConfig.A3MConfig.ExamineContents != true {
		t.Error("Expected ExamineContents to be true")
	}

	if retrievedConfig.A3MConfig.AipCompressionLevel != 9 {
		t.Errorf("Expected AipCompressionLevel 9, got %d", retrievedConfig.A3MConfig.AipCompressionLevel)
	}

	if retrievedConfig.A3MConfig.ThumbnailMode != transferservice.ProcessingConfig_THUMBNAIL_MODE_DO_NOT_GENERATE {
		t.Errorf("Expected ThumbnailMode DO_NOT_GENERATE, got %v", retrievedConfig.A3MConfig.ThumbnailMode)
	}
}
