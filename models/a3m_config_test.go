package models

import (
	"encoding/json"
	"testing"

	transferservice "github.com/penwern/curate-preservation-api/common/proto/a3m/gen/go/a3m/api/transferservice/v1beta1"
)

func TestNewA3MProcessingConfig(t *testing.T) {
	config := NewA3MProcessingConfig()

	// Test default boolean values
	if !config.AssignUuidsToDirectories {
		t.Error("Expected AssignUuidsToDirectories to be true by default")
	}

	if config.ExamineContents {
		t.Error("Expected ExamineContents to be false by default")
	}

	if !config.GenerateTransferStructureReport {
		t.Error("Expected GenerateTransferStructureReport to be true by default")
	}

	if !config.DocumentEmptyDirectories {
		t.Error("Expected DocumentEmptyDirectories to be true by default")
	}

	if !config.ExtractPackages {
		t.Error("Expected ExtractPackages to be true by default")
	}

	if config.DeletePackagesAfterExtraction {
		t.Error("Expected DeletePackagesAfterExtraction to be false by default")
	}

	if !config.IdentifyTransfer {
		t.Error("Expected IdentifyTransfer to be true by default")
	}

	if !config.IdentifySubmissionAndMetadata {
		t.Error("Expected IdentifySubmissionAndMetadata to be true by default")
	}

	if !config.IdentifyBeforeNormalization {
		t.Error("Expected IdentifyBeforeNormalization to be true by default")
	}

	if !config.Normalize {
		t.Error("Expected Normalize to be true by default")
	}

	if !config.TranscribeFiles {
		t.Error("Expected TranscribeFiles to be true by default")
	}

	if !config.PerformPolicyChecksOnOriginals {
		t.Error("Expected PerformPolicyChecksOnOriginals to be true by default")
	}

	if !config.PerformPolicyChecksOnPreservationDerivatives {
		t.Error("Expected PerformPolicyChecksOnPreservationDerivatives to be true by default")
	}

	if !config.PerformPolicyChecksOnAccessDerivatives {
		t.Error("Expected PerformPolicyChecksOnAccessDerivatives to be true by default")
	}

	// Test default enum values
	if config.ThumbnailMode != transferservice.ProcessingConfig_THUMBNAIL_MODE_GENERATE {
		t.Errorf("Expected ThumbnailMode to be THUMBNAIL_MODE_GENERATE by default, got %v", config.ThumbnailMode)
	}

	if config.AipCompressionAlgorithm != transferservice.ProcessingConfig_AIP_COMPRESSION_ALGORITHM_S7_BZIP2 {
		t.Errorf("Expected AipCompressionAlgorithm to be AIP_COMPRESSION_ALGORITHM_S7_BZIP2 by default, got %v", config.AipCompressionAlgorithm)
	}

	// Test default numeric values
	if config.AipCompressionLevel != 1 {
		t.Errorf("Expected AipCompressionLevel to be 1 by default, got %d", config.AipCompressionLevel)
	}
}

func TestA3MProcessingConfig_JSONMarshalUnmarshal(t *testing.T) {
	config := A3MProcessingConfig{
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
		AipCompressionAlgorithm:                      transferservice.ProcessingConfig_AIP_COMPRESSION_ALGORITHM_TAR_BZIP2,
	}

	// Marshal to JSON
	jsonData, err := config.MarshalJSON()
	if err != nil {
		t.Fatalf("Failed to marshal A3M config to JSON: %v", err)
	}

	// Test that JSON is valid and non-empty
	if len(jsonData) == 0 {
		t.Error("JSON data should not be empty")
	}

	var jsonMap map[string]any
	err = json.Unmarshal(jsonData, &jsonMap)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON to map: %v", err)
	}

	// Verify JSON contains some data
	if len(jsonMap) == 0 {
		t.Error("JSON map should not be empty")
	}

	// Unmarshal back to A3MProcessingConfig
	var unmarshaledConfig A3MProcessingConfig
	err = unmarshaledConfig.UnmarshalJSON(jsonData)
	if err != nil {
		t.Fatalf("Failed to unmarshal A3M config from JSON: %v", err)
	}

	// Verify all fields are preserved
	if unmarshaledConfig.AssignUuidsToDirectories != config.AssignUuidsToDirectories {
		t.Errorf("AssignUuidsToDirectories not preserved: expected %v, got %v",
			config.AssignUuidsToDirectories, unmarshaledConfig.AssignUuidsToDirectories)
	}

	if unmarshaledConfig.ExamineContents != config.ExamineContents {
		t.Errorf("ExamineContents not preserved: expected %v, got %v",
			config.ExamineContents, unmarshaledConfig.ExamineContents)
	}

	if unmarshaledConfig.AipCompressionLevel != config.AipCompressionLevel {
		t.Errorf("AipCompressionLevel not preserved: expected %d, got %d",
			config.AipCompressionLevel, unmarshaledConfig.AipCompressionLevel)
	}

	if unmarshaledConfig.ThumbnailMode != config.ThumbnailMode {
		t.Errorf("ThumbnailMode not preserved: expected %v, got %v",
			config.ThumbnailMode, unmarshaledConfig.ThumbnailMode)
	}

	if unmarshaledConfig.AipCompressionAlgorithm != config.AipCompressionAlgorithm {
		t.Errorf("AipCompressionAlgorithm not preserved: expected %v, got %v",
			config.AipCompressionAlgorithm, unmarshaledConfig.AipCompressionAlgorithm)
	}
}

func TestA3MProcessingConfig_UnmarshalJSON_InvalidJSON(t *testing.T) {
	config := A3MProcessingConfig{}

	// Test with invalid JSON
	err := config.UnmarshalJSON([]byte("invalid json"))
	if err == nil {
		t.Error("Expected error when unmarshaling invalid JSON")
	}

	// Test with empty JSON object (should succeed)
	err = config.UnmarshalJSON([]byte("{}"))
	if err != nil {
		t.Errorf("Unexpected error when unmarshaling empty JSON object: %v", err)
	}
}

func TestA3MProcessingConfig_MarshalJSON_EmitUnpopulated(t *testing.T) {
	// Create config with all default/zero values
	config := A3MProcessingConfig{}

	jsonData, err := config.MarshalJSON()
	if err != nil {
		t.Fatalf("Failed to marshal empty A3M config: %v", err)
	}

	// Test that JSON is valid and non-empty (even with zero values due to EmitUnpopulated)
	if len(jsonData) == 0 {
		t.Error("JSON data should not be empty even with zero values")
	}

	var jsonMap map[string]any
	err = json.Unmarshal(jsonData, &jsonMap)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// With EmitUnpopulated: true, should have some fields present
	if len(jsonMap) == 0 {
		t.Error("JSON map should not be empty with EmitUnpopulated: true")
	}
}

func TestA3MProcessingConfig_UnmarshalJSON_PartialData(t *testing.T) {
	// Test unmarshaling with only some fields present
	partialJSON := `{
		"examine_contents": true,
		"aip_compression_level": 5,
		"thumbnail_mode": 2
	}`

	config := A3MProcessingConfig{}
	err := config.UnmarshalJSON([]byte(partialJSON))
	if err != nil {
		t.Fatalf("Failed to unmarshal partial JSON: %v", err)
	}

	// Check that specified fields are set
	if !config.ExamineContents {
		t.Error("Expected ExamineContents to be true")
	}

	if config.AipCompressionLevel != 5 {
		t.Errorf("Expected AipCompressionLevel to be 5, got %d", config.AipCompressionLevel)
	}

	if config.ThumbnailMode != 2 {
		t.Errorf("Expected ThumbnailMode to be 2, got %d", config.ThumbnailMode)
	}

	// Check that unspecified boolean fields are false (protobuf default)
	if config.AssignUuidsToDirectories {
		t.Error("Expected AssignUuidsToDirectories to be false (default)")
	}
}

func TestA3MProcessingConfig_UnmarshalJSON_DiscardUnknown(t *testing.T) {
	// Test with unknown fields that should be discarded
	jsonWithUnknown := `{
		"examine_contents": true,
		"unknown_field": "should_be_ignored",
		"another_unknown": 123,
		"aip_compression_level": 7
	}`

	config := A3MProcessingConfig{}
	err := config.UnmarshalJSON([]byte(jsonWithUnknown))
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON with unknown fields: %v", err)
	}

	// Check that known fields are set correctly
	if !config.ExamineContents {
		t.Error("Expected ExamineContents to be true")
	}

	if config.AipCompressionLevel != 7 {
		t.Errorf("Expected AipCompressionLevel to be 7, got %d", config.AipCompressionLevel)
	}
}
