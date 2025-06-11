package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/penwern/curate-preservation-core-api/config"
	"github.com/penwern/curate-preservation-core-api/database"
	"github.com/penwern/curate-preservation-core-api/models"
)

func setupTestServer(t *testing.T) *Server {
	t.Helper()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	cfg := config.Config{
		DBType:       "sqlite3",
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
		DBType:       "sqlite3",
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
	config := models.NewPreservationConfig("Original Name", "Original Description")
	err := server.db.CreateConfig(config)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Update request
	updateReq := map[string]interface{}{
		"name":        "Updated Name",
		"description": "Updated Description",
		"a3m_config": map[string]interface{}{
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
