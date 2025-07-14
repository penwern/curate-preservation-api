package models

import (
	"encoding/json"
	"testing"
	"time"

	transferservice "github.com/penwern/curate-preservation-api/common/proto/a3m/gen/go/a3m/api/transferservice/v1beta1"
)

func TestNewPreservationConfig(t *testing.T) {
	name := "Test Config"
	description := "Test Description"

	config := NewPreservationConfig(name, description)

	if config.Name != name {
		t.Errorf("Expected name '%s', got '%s'", name, config.Name)
	}

	if config.Description != description {
		t.Errorf("Expected description '%s', got '%s'", description, config.Description)
	}

	if config.CompressAIP != false {
		t.Errorf("Expected CompressAIP to be false by default, got %v", config.CompressAIP)
	}

	if config.ID != 0 {
		t.Errorf("Expected ID to be 0 for new config, got %d", config.ID)
	}

	// Test that A3MConfig has default values
	if !config.A3MConfig.AssignUuidsToDirectories {
		t.Error("Expected AssignUuidsToDirectories to be true by default")
	}

	if config.A3MConfig.ExamineContents {
		t.Error("Expected ExamineContents to be false by default")
	}

	if config.A3MConfig.AipCompressionLevel != 1 {
		t.Errorf("Expected AipCompressionLevel to be 1 by default, got %d", config.A3MConfig.AipCompressionLevel)
	}
}

func TestPreservationConfig_JSONMarshalUnmarshal(t *testing.T) {
	config := &PreservationConfig{
		ID:          123,
		Name:        "Test Config",
		Description: "Test Description",
		CompressAIP: true,
		A3MConfig: A3MProcessingConfig{
			AssignUuidsToDirectories: false,
			ExamineContents:          true,
			AipCompressionLevel:      9,
			ThumbnailMode:            transferservice.ProcessingConfig_THUMBNAIL_MODE_DO_NOT_GENERATE,
		},
		CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC),
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("Failed to marshal config to JSON: %v", err)
	}

	// Unmarshal from JSON
	var unmarshaledConfig PreservationConfig
	err = json.Unmarshal(jsonData, &unmarshaledConfig)
	if err != nil {
		t.Fatalf("Failed to unmarshal config from JSON: %v", err)
	}

	// Verify basic fields
	if unmarshaledConfig.ID != config.ID {
		t.Errorf("Expected ID %d, got %d", config.ID, unmarshaledConfig.ID)
	}

	if unmarshaledConfig.Name != config.Name {
		t.Errorf("Expected name '%s', got '%s'", config.Name, unmarshaledConfig.Name)
	}

	if unmarshaledConfig.Description != config.Description {
		t.Errorf("Expected description '%s', got '%s'", config.Description, unmarshaledConfig.Description)
	}

	if unmarshaledConfig.CompressAIP != config.CompressAIP {
		t.Errorf("Expected CompressAIP %v, got %v", config.CompressAIP, unmarshaledConfig.CompressAIP)
	}

	// Verify A3M config fields
	if unmarshaledConfig.A3MConfig.AssignUuidsToDirectories != config.A3MConfig.AssignUuidsToDirectories {
		t.Errorf("Expected AssignUuidsToDirectories %v, got %v", 
			config.A3MConfig.AssignUuidsToDirectories, unmarshaledConfig.A3MConfig.AssignUuidsToDirectories)
	}

	if unmarshaledConfig.A3MConfig.ExamineContents != config.A3MConfig.ExamineContents {
		t.Errorf("Expected ExamineContents %v, got %v", 
			config.A3MConfig.ExamineContents, unmarshaledConfig.A3MConfig.ExamineContents)
	}

	if unmarshaledConfig.A3MConfig.AipCompressionLevel != config.A3MConfig.AipCompressionLevel {
		t.Errorf("Expected AipCompressionLevel %d, got %d", 
			config.A3MConfig.AipCompressionLevel, unmarshaledConfig.A3MConfig.AipCompressionLevel)
	}

	if unmarshaledConfig.A3MConfig.ThumbnailMode != config.A3MConfig.ThumbnailMode {
		t.Errorf("Expected ThumbnailMode %v, got %v", 
			config.A3MConfig.ThumbnailMode, unmarshaledConfig.A3MConfig.ThumbnailMode)
	}
}

func TestPreservationConfig_EmptyFields(t *testing.T) {
	config := NewPreservationConfig("", "")

	if config.Name != "" {
		t.Errorf("Expected empty name, got '%s'", config.Name)
	}

	if config.Description != "" {
		t.Errorf("Expected empty description, got '%s'", config.Description)
	}

	// Should still have valid A3MConfig defaults
	if !config.A3MConfig.AssignUuidsToDirectories {
		t.Error("Expected AssignUuidsToDirectories to be true even with empty fields")
	}
}

func TestPreservationConfig_TimeFields(t *testing.T) {
	config := NewPreservationConfig("Test", "Test")

	// New config should have zero time values
	if !config.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be zero time for new config")
	}

	if !config.UpdatedAt.IsZero() {
		t.Error("Expected UpdatedAt to be zero time for new config")
	}

	// Set time values and test
	now := time.Now()
	config.CreatedAt = now
	config.UpdatedAt = now

	if config.CreatedAt != now {
		t.Error("CreatedAt was not set correctly")
	}

	if config.UpdatedAt != now {
		t.Error("UpdatedAt was not set correctly")
	}
}

func TestPreservationConfig_LongStrings(t *testing.T) {
	longName := "This is a very long name that might exceed normal length expectations for testing purposes"
	longDescription := "This is an extremely long description that contains many words and sentences to test how the system handles longer text fields. It should handle this gracefully without any issues, but we want to make sure that the marshaling and unmarshaling still works correctly with longer strings."

	config := NewPreservationConfig(longName, longDescription)

	if config.Name != longName {
		t.Error("Long name not preserved correctly")
	}

	if config.Description != longDescription {
		t.Error("Long description not preserved correctly")
	}

	// Test JSON round-trip with long strings
	jsonData, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("Failed to marshal config with long strings: %v", err)
	}

	var unmarshaledConfig PreservationConfig
	err = json.Unmarshal(jsonData, &unmarshaledConfig)
	if err != nil {
		t.Fatalf("Failed to unmarshal config with long strings: %v", err)
	}

	if unmarshaledConfig.Name != longName {
		t.Error("Long name not preserved after JSON round-trip")
	}

	if unmarshaledConfig.Description != longDescription {
		t.Error("Long description not preserved after JSON round-trip")
	}
}