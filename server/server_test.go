package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/penwern/curate-preservation-api/database"
	"github.com/penwern/curate-preservation-api/models"
	"github.com/penwern/curate-preservation-api/pkg/config"
	"github.com/penwern/curate-preservation-api/pkg/logger"
)

const (
	testDBType       = "sqlite3"
	testOriginalName = "Original Name"
	testOriginalDesc = "Original Description"
)

func setupTestServer(t *testing.T) *Server {
	t.Helper()

	logger.Initialize("debug", "")

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	cfg := config.Config{
		DBType:       testDBType,
		DBConnection: dbPath,
		Port:         8080,
	}

	server, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}

	return server
}

func TestServer_New(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	cfg := config.Config{
		DBType:       testDBType,
		DBConnection: dbPath,
		Port:         8080,
	}

	server, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}
	defer server.Shutdown()

	if server.router == nil {
		t.Error("Expected router to be initialized")
	}

	if server.db == nil {
		t.Error("Expected database to be initialized")
	}
}

func TestServer_New_InvalidDBConfig(t *testing.T) {
	cfg := config.Config{
		DBType:       "invalid",
		DBConnection: "invalid",
		Port:         8080,
	}

	_, err := New(cfg)
	if err == nil {
		t.Error("Expected error for invalid database config, got nil")
	}
}

func TestServer_HandleHealth(t *testing.T) {
	server := setupTestServer(t)
	defer server.Shutdown()

	req, err := http.NewRequest("GET", "/api/v1/health", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	server.router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response map[string]string
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["status"] != "ok" {
		t.Errorf("Expected status 'ok', got '%s'", response["status"])
	}
}

func TestServer_HandleListConfigs_Empty(t *testing.T) {
	server := setupTestServer(t)
	defer server.Shutdown()

	req, err := http.NewRequest("GET", "/api/v1/preservation-configs", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	server.router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var configs []models.PreservationConfig
	err = json.Unmarshal(rr.Body.Bytes(), &configs)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if len(configs) != 1 {
		t.Errorf("Expected 1 config, got %d", len(configs))
	}
}

func TestServer_HandleCreateConfig(t *testing.T) {
	server := setupTestServer(t)
	defer server.Shutdown()

	createReq := map[string]string{
		"name":        "Test Config",
		"description": "Test Description",
	}

	reqBody, err := json.Marshal(createReq)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	req, err := http.NewRequest("POST", "/api/v1/preservation-configs", bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	server.router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}

	var config models.PreservationConfig
	err = json.Unmarshal(rr.Body.Bytes(), &config)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if config.Name != createReq["name"] {
		t.Errorf("Expected name '%s', got '%s'", createReq["name"], config.Name)
	}

	if config.Description != createReq["description"] {
		t.Errorf("Expected description '%s', got '%s'", createReq["description"], config.Description)
	}

	if config.ID == 0 {
		t.Error("Expected config ID to be set")
	}
}

func TestServer_HandleCreateConfig_InvalidJSON(t *testing.T) {
	server := setupTestServer(t)
	defer server.Shutdown()

	req, err := http.NewRequest("POST", "/api/v1/preservation-configs", bytes.NewBuffer([]byte("invalid json")))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	server.router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
}

func TestServer_HandleCreateConfig_MissingName(t *testing.T) {
	server := setupTestServer(t)
	defer server.Shutdown()

	createReq := map[string]string{
		"description": "Test Description",
	}

	reqBody, err := json.Marshal(createReq)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	req, err := http.NewRequest("POST", "/api/v1/preservation-configs", bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	server.router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
}

func TestServer_HandleCreateConfig_WithPartialA3MConfig(t *testing.T) {
	server := setupTestServer(t)
	defer server.Shutdown()

	createReq := map[string]any{
		"name":        "Partial A3M Config",
		"description": "Config with some custom A3M settings",
		"a3m_config": map[string]any{
			"examine_contents":      true,
			"normalize":             false,
			"aip_compression_level": 9,
			"extract_packages":      false,
		},
	}

	reqBody, err := json.Marshal(createReq)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	req, err := http.NewRequest("POST", "/api/v1/preservation-configs", bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	server.router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}

	var config models.PreservationConfig
	err = json.Unmarshal(rr.Body.Bytes(), &config)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Check that custom values were applied
	if !config.A3MConfig.ExamineContents {
		t.Error("Expected ExamineContents to be true")
	}
	if config.A3MConfig.Normalize {
		t.Error("Expected Normalize to be false")
	}
	if config.A3MConfig.AipCompressionLevel != 9 {
		t.Errorf("Expected AipCompressionLevel to be 9, got %d", config.A3MConfig.AipCompressionLevel)
	}
	if config.A3MConfig.ExtractPackages {
		t.Error("Expected ExtractPackages to be false")
	}

	// Check that defaults were preserved for unspecified fields
	if !config.A3MConfig.AssignUuidsToDirectories {
		t.Error("Expected AssignUuidsToDirectories to be true (default)")
	}
	if !config.A3MConfig.IdentifyTransfer {
		t.Error("Expected IdentifyTransfer to be true (default)")
	}
}

func TestServer_HandleCreateConfig_WithFullA3MConfig(t *testing.T) {
	server := setupTestServer(t)
	defer server.Shutdown()

	createReq := map[string]any{
		"name":        "Full A3M Config",
		"description": "Config with complete A3M settings",
		"a3m_config": map[string]any{
			"assign_uuids_to_directories":                       false,
			"examine_contents":                                  true,
			"generate_transfer_structure_report":                false,
			"document_empty_directories":                        false,
			"extract_packages":                                  false,
			"delete_packages_after_extraction":                  true,
			"identify_transfer":                                 false,
			"identify_submission_and_metadata":                  false,
			"identify_before_normalization":                     false,
			"normalize":                                         false,
			"transcribe_files":                                  false,
			"perform_policy_checks_on_originals":                false,
			"perform_policy_checks_on_preservation_derivatives": false,
			"perform_policy_checks_on_access_derivatives":       false,
			"thumbnail_mode":                                    0,
			"aip_compression_level":                             9,
			"aip_compression_algorithm":                         1,
		},
	}

	reqBody, err := json.Marshal(createReq)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	req, err := http.NewRequest("POST", "/api/v1/preservation-configs", bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	server.router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}

	var config models.PreservationConfig
	err = json.Unmarshal(rr.Body.Bytes(), &config)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Verify all custom values were applied
	if config.A3MConfig.AssignUuidsToDirectories {
		t.Error("Expected AssignUuidsToDirectories to be false")
	}
	if !config.A3MConfig.ExamineContents {
		t.Error("Expected ExamineContents to be true")
	}
	if config.A3MConfig.GenerateTransferStructureReport {
		t.Error("Expected GenerateTransferStructureReport to be false")
	}
	if config.A3MConfig.AipCompressionLevel != 9 {
		t.Errorf("Expected AipCompressionLevel to be 9, got %d", config.A3MConfig.AipCompressionLevel)
	}
}

func TestServer_HandleGetConfig(t *testing.T) {
	server := setupTestServer(t)
	defer server.Shutdown()

	// First create a config
	config := models.NewPreservationConfig("Test Config", "Test Description")
	err := server.db.CreateConfig(config)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Now get it via the API
	req, err := http.NewRequest("GET", fmt.Sprintf("/api/v1/preservation-configs/%d", config.ID), nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	server.router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var retrievedConfig models.PreservationConfig
	err = json.Unmarshal(rr.Body.Bytes(), &retrievedConfig)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if retrievedConfig.Name != config.Name {
		t.Errorf("Expected name '%s', got '%s'", config.Name, retrievedConfig.Name)
	}
}

func TestServer_HandleGetConfig_NotFound(t *testing.T) {
	server := setupTestServer(t)
	defer server.Shutdown()

	req, err := http.NewRequest("GET", "/api/v1/preservation-configs/999", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	server.router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusNotFound)
	}
}

func TestServer_HandleGetConfig_InvalidID(t *testing.T) {
	server := setupTestServer(t)
	defer server.Shutdown()

	req, err := http.NewRequest("GET", "/api/v1/preservation-configs/invalid", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	server.router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
}

func TestServer_HandleUpdateConfig(t *testing.T) {
	server := setupTestServer(t)
	defer server.Shutdown()

	// First create a config
	config := models.NewPreservationConfig(testOriginalName, testOriginalDesc)
	err := server.db.CreateConfig(config)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Update request
	updateReq := map[string]any{
		"name":        "Updated Name",
		"description": "Updated Description",
		"a3m_config": map[string]any{
			"examine_contents": true,
		},
	}

	reqBody, err := json.Marshal(updateReq)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	req, err := http.NewRequest("PUT", fmt.Sprintf("/api/v1/preservation-configs/%d", config.ID), bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	server.router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var updatedConfig models.PreservationConfig
	err = json.Unmarshal(rr.Body.Bytes(), &updatedConfig)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if updatedConfig.Name != updateReq["name"] {
		t.Errorf("Expected name '%s', got '%s'", updateReq["name"], updatedConfig.Name)
	}
}

func TestServer_HandleUpdateConfig_PartialUpdate(t *testing.T) {
	server := setupTestServer(t)
	defer server.Shutdown()

	// First create a config with specific values
	config := models.NewPreservationConfig(testOriginalName, testOriginalDesc)
	config.A3MConfig.ExamineContents = false
	config.A3MConfig.Normalize = true
	config.A3MConfig.AipCompressionLevel = 1
	err := server.db.CreateConfig(config)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Partial update - only update description and one A3M field
	updateReq := map[string]any{
		"description":  "Updated Description Only",
		"compress_aip": false,
		"a3m_config": map[string]any{
			"examine_contents": true, // Change this
			// Don't specify other fields - they should remain unchanged
		},
	}

	reqBody, err := json.Marshal(updateReq)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	req, err := http.NewRequest("PUT", fmt.Sprintf("/api/v1/preservation-configs/%d", config.ID), bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	server.router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var updatedConfig models.PreservationConfig
	err = json.Unmarshal(rr.Body.Bytes(), &updatedConfig)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Check that specified fields were updated
	if updatedConfig.Description != "Updated Description Only" {
		t.Errorf("Expected description 'Updated Description Only', got '%s'", updatedConfig.Description)
	}
	if updatedConfig.CompressAIP {
		t.Error("Expected CompressAIP to be false")
	}
	if !updatedConfig.A3MConfig.ExamineContents {
		t.Error("Expected ExamineContents to be updated to true")
	}

	// Check that unspecified fields remained unchanged
	if updatedConfig.Name != testOriginalName {
		t.Errorf("Expected name to remain '%s', got '%s'", testOriginalName, updatedConfig.Name)
	}
	if !updatedConfig.A3MConfig.Normalize {
		t.Error("Expected Normalize to remain true (unchanged)")
	}
	if updatedConfig.A3MConfig.AipCompressionLevel != 1 {
		t.Errorf("Expected AipCompressionLevel to remain 1, got %d", updatedConfig.A3MConfig.AipCompressionLevel)
	}
}

func TestServer_HandleUpdateConfig_OnlyA3MConfig(t *testing.T) {
	server := setupTestServer(t)
	defer server.Shutdown()

	// First create a config
	config := models.NewPreservationConfig(testOriginalName, testOriginalDesc)
	err := server.db.CreateConfig(config)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Update only A3M config fields
	updateReq := map[string]any{
		"a3m_config": map[string]any{
			"examine_contents":      true,
			"normalize":             false,
			"aip_compression_level": 9,
			"extract_packages":      false,
		},
	}

	reqBody, err := json.Marshal(updateReq)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	req, err := http.NewRequest("PUT", fmt.Sprintf("/api/v1/preservation-configs/%d", config.ID), bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	server.router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var updatedConfig models.PreservationConfig
	err = json.Unmarshal(rr.Body.Bytes(), &updatedConfig)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Check that basic fields remained unchanged
	if updatedConfig.Name != testOriginalName {
		t.Errorf("Expected name to remain '%s', got '%s'", testOriginalName, updatedConfig.Name)
	}
	if updatedConfig.Description != testOriginalDesc {
		t.Errorf("Expected description to remain '%s', got '%s'", testOriginalDesc, updatedConfig.Description)
	}

	// Check that A3M config fields were updated
	if !updatedConfig.A3MConfig.ExamineContents {
		t.Error("Expected ExamineContents to be true")
	}
	if updatedConfig.A3MConfig.Normalize {
		t.Error("Expected Normalize to be false")
	}
	if updatedConfig.A3MConfig.AipCompressionLevel != 9 {
		t.Errorf("Expected AipCompressionLevel to be 9, got %d", updatedConfig.A3MConfig.AipCompressionLevel)
	}
	if updatedConfig.A3MConfig.ExtractPackages {
		t.Error("Expected ExtractPackages to be false")
	}

	// Check that unspecified A3M fields remained at defaults
	if !updatedConfig.A3MConfig.AssignUuidsToDirectories {
		t.Error("Expected AssignUuidsToDirectories to remain true (default)")
	}
}

func TestServer_HandleUpdateConfig_EmptyDescription(t *testing.T) {
	server := setupTestServer(t)
	defer server.Shutdown()

	// First create a config
	config := models.NewPreservationConfig(testOriginalName, testOriginalDesc)
	err := server.db.CreateConfig(config)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Update with empty description (should clear it)
	updateReq := map[string]any{
		"description": "",
	}

	reqBody, err := json.Marshal(updateReq)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	req, err := http.NewRequest("PUT", fmt.Sprintf("/api/v1/preservation-configs/%d", config.ID), bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	server.router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var updatedConfig models.PreservationConfig
	err = json.Unmarshal(rr.Body.Bytes(), &updatedConfig)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Check that description was cleared
	if updatedConfig.Description != "" {
		t.Errorf("Expected description to be empty, got '%s'", updatedConfig.Description)
	}

	// Check that name remained unchanged
	if updatedConfig.Name != testOriginalName {
		t.Errorf("Expected name to remain '%s', got '%s'", testOriginalName, updatedConfig.Name)
	}
}

func TestServer_HandleDeleteConfig(t *testing.T) {
	server := setupTestServer(t)
	defer server.Shutdown()

	// First create a config
	config := models.NewPreservationConfig("To Delete", "Will be deleted")
	err := server.db.CreateConfig(config)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Delete it
	req, err := http.NewRequest("DELETE", fmt.Sprintf("/api/v1/preservation-configs/%d", config.ID), nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	server.router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNoContent {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusNoContent)
	}

	// Verify it's gone
	_, err = server.db.GetConfig(config.ID)
	if err != database.ErrNotFound {
		t.Errorf("Expected config to be deleted, but it still exists")
	}
}

func TestServer_Shutdown(t *testing.T) {
	server := setupTestServer(t)

	// Test that shutdown doesn't return an error
	err := server.Shutdown()
	if err != nil {
		t.Errorf("Shutdown failed: %v", err)
	}
}

func TestServer_Integration_FullWorkflow(t *testing.T) {
	server := setupTestServer(t)
	defer server.Shutdown()

	// 1. Check health
	req, _ := http.NewRequest("GET", "/api/v1/health", nil)
	rr := httptest.NewRecorder()
	server.router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("Health check failed")
	}

	// 2. List configs (should be empty)
	req, _ = http.NewRequest("GET", "/api/v1/preservation-configs", nil)
	rr = httptest.NewRecorder()
	server.router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("List configs failed")
	}

	// 3. Create a config
	createReq := map[string]string{
		"name":        "Integration Test Config",
		"description": "Created during integration test",
	}
	reqBody, _ := json.Marshal(createReq)
	req, _ = http.NewRequest("POST", "/api/v1/preservation-configs", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	server.router.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("Create config failed")
	}

	var createdConfig models.PreservationConfig
	json.Unmarshal(rr.Body.Bytes(), &createdConfig)

	// 4. Get the created config
	req, _ = http.NewRequest("GET", fmt.Sprintf("/api/v1/preservation-configs/%d", createdConfig.ID), nil)
	rr = httptest.NewRecorder()
	server.router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("Get config failed")
	}

	// 5. Update the config
	updateReq := map[string]string{
		"name":        "Updated Integration Test Config",
		"description": "Updated during integration test",
	}
	reqBody, _ = json.Marshal(updateReq)
	req, _ = http.NewRequest("PUT", fmt.Sprintf("/api/v1/preservation-configs/%d", createdConfig.ID), bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	server.router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("Update config failed")
	}

	// 6. Delete the config
	req, _ = http.NewRequest("DELETE", fmt.Sprintf("/api/v1/preservation-configs/%d", createdConfig.ID), nil)
	rr = httptest.NewRecorder()
	server.router.ServeHTTP(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("Delete config failed")
	}
}

func TestServer_CORS_Methods(t *testing.T) {
	server := setupTestServer(t)
	defer server.Shutdown()

	methods := []string{"GET", "POST", "PUT", "DELETE"}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			var url string
			var body []byte

			switch method {
			case "GET":
				url = "/api/v1/preservation-configs"
			case "POST":
				url = "/api/v1/preservation-configs"
				createReq := map[string]string{"name": "Test", "description": "Test"}
				body, _ = json.Marshal(createReq)
			case "PUT", "DELETE":
				// First create a config to update/delete
				config := models.NewPreservationConfig("Test", "Test")
				server.db.CreateConfig(config)
				url = fmt.Sprintf("/api/v1/preservation-configs/%d", config.ID)
				if method == "PUT" {
					updateReq := map[string]string{"name": "Updated", "description": "Updated"}
					body, _ = json.Marshal(updateReq)
				}
			}

			req, _ := http.NewRequest(method, url, bytes.NewBuffer(body))
			if len(body) > 0 {
				req.Header.Set("Content-Type", "application/json")
			}

			rr := httptest.NewRecorder()
			server.router.ServeHTTP(rr, req)

			// Should not return method not allowed
			if rr.Code == http.StatusMethodNotAllowed {
				t.Errorf("Method %s not allowed for %s", method, url)
			}
		})
	}
}

func TestServer_HandleCreateConfig_EmptyName(t *testing.T) {
	server := setupTestServer(t)
	defer server.Shutdown()

	createReq := map[string]any{
		"name":        "", // Empty name should fail
		"description": "Test Description",
	}

	reqBody, err := json.Marshal(createReq)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	req, err := http.NewRequest("POST", "/api/v1/preservation-configs", bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	server.router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
}

func TestServer_HandleUpdateConfig_IDMismatch(t *testing.T) {
	server := setupTestServer(t)
	defer server.Shutdown()

	// First create a config
	config := models.NewPreservationConfig(testOriginalName, testOriginalDesc)
	err := server.db.CreateConfig(config)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Update with mismatched ID in body
	updateReq := map[string]any{
		"id":          999, // Different from URL
		"name":        "Updated Name",
		"description": "Updated Description",
	}

	reqBody, err := json.Marshal(updateReq)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	req, err := http.NewRequest("PUT", fmt.Sprintf("/api/v1/preservation-configs/%d", config.ID), bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	server.router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
}

func TestServer_HandleUpdateConfig_NoFieldsProvided(t *testing.T) {
	server := setupTestServer(t)
	defer server.Shutdown()

	// First create a config
	config := models.NewPreservationConfig(testOriginalName, testOriginalDesc)
	err := server.db.CreateConfig(config)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Update with empty body (should succeed but not change anything)
	updateReq := map[string]any{}

	reqBody, err := json.Marshal(updateReq)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	req, err := http.NewRequest("PUT", fmt.Sprintf("/api/v1/preservation-configs/%d", config.ID), bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	server.router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var updatedConfig models.PreservationConfig
	err = json.Unmarshal(rr.Body.Bytes(), &updatedConfig)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Check that nothing changed
	if updatedConfig.Name != testOriginalName {
		t.Errorf("Expected name to remain '%s', got '%s'", testOriginalName, updatedConfig.Name)
	}
	if updatedConfig.Description != testOriginalDesc {
		t.Errorf("Expected description to remain '%s', got '%s'", testOriginalDesc, updatedConfig.Description)
	}
}
