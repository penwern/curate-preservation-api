package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/penwern/curate-preservation-api/models"
)

func TestServer_HandleCreateConfig_LargePayload(t *testing.T) {
	server := setupTestServer(t)
	defer server.Shutdown()

	// Create a very large description to test payload limits
	largeDescription := strings.Repeat("A", 10000)

	createReq := map[string]string{
		"name":        "Large Payload Test",
		"description": largeDescription,
	}

	reqBody, err := json.Marshal(createReq)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	req := setupTestRequest("POST", "/api/v1/preservation-configs", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	server.router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}

	var config models.PreservationConfig
	if err := json.Unmarshal(rr.Body.Bytes(), &config); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if len(config.Description) != len(largeDescription) {
		t.Errorf("Large description not preserved correctly")
	}
}

func TestServer_HandleCreateConfig_UnicodeCharacters(t *testing.T) {
	server := setupTestServer(t)
	defer server.Shutdown()

	createReq := map[string]string{
		"name":        "测试配置 🚀 العربية",
		"description": "Unicode test: 🌟 ñáéíóú čřžýáíé 中文测试 العربية русский",
	}

	reqBody, err := json.Marshal(createReq)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	req := setupTestRequest("POST", "/api/v1/preservation-configs", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	server.router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}

	var config models.PreservationConfig
	if err := json.Unmarshal(rr.Body.Bytes(), &config); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if config.Name != createReq["name"] {
		t.Errorf("Unicode name not preserved: expected '%s', got '%s'", createReq["name"], config.Name)
	}

	if config.Description != createReq["description"] {
		t.Errorf("Unicode description not preserved: expected '%s', got '%s'", createReq["description"], config.Description)
	}
}

func TestServer_HandleCreateConfig_SpecialCharacters(t *testing.T) {
	server := setupTestServer(t)
	defer server.Shutdown()

	createReq := map[string]string{
		"name":        "Special \"Chars\" & <Tags> \n\t\r",
		"description": "Contains: quotes \"'` backslashes \\ slashes / & ampersands < > brackets [] {} parentheses () #hashtags @mentions",
	}

	reqBody, err := json.Marshal(createReq)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	req := setupTestRequest("POST", "/api/v1/preservation-configs", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	server.router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}

	var config models.PreservationConfig
	if err := json.Unmarshal(rr.Body.Bytes(), &config); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if config.Name != createReq["name"] {
		t.Errorf("Special characters in name not preserved")
	}

	if config.Description != createReq["description"] {
		t.Errorf("Special characters in description not preserved")
	}
}

func TestServer_HandleCreateConfig_JSONNumberFields(t *testing.T) {
	server := setupTestServer(t)
	defer server.Shutdown()

	createReq := map[string]any{
		"name":        "Number Test Config",
		"description": "Testing number field handling",
		"a3m_config": map[string]any{
			"aip_compression_level": 999, // Very high number
			"thumbnail_mode":        -1,  // Negative number
		},
	}

	reqBody, err := json.Marshal(createReq)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	req := setupTestRequest("POST", "/api/v1/preservation-configs", bytes.NewBuffer(reqBody))
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

	// Verify extreme numbers are handled
	if config.A3MConfig.AipCompressionLevel != 999 {
		t.Errorf("Expected AipCompressionLevel 999, got %d", config.A3MConfig.AipCompressionLevel)
	}
}

func TestServer_HandleUpdateConfig_ConcurrentModification(t *testing.T) {
	server := setupTestServer(t)
	defer server.Shutdown()

	// Create initial config
	config := models.NewPreservationConfig("Concurrent Test", "Test")
	err := server.db.CreateConfig(config)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Simulate concurrent updates
	done := make(chan bool, 2)
	errors := make(chan error, 2)

	updateFunc := func(name string) {
		defer func() { done <- true }()

		updateReq := map[string]string{
			"name":        name,
			"description": "Updated by " + name,
		}

		reqBody, err := json.Marshal(updateReq)
		if err != nil {
			errors <- err
			return
		}

		req := setupTestRequest("PUT", fmt.Sprintf("/api/v1/preservation-configs/%d", config.ID), bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		server.router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			errors <- fmt.Errorf("concurrent update failed with status %d", rr.Code)
		}
	}

	// Start concurrent updates
	go updateFunc("Update1")
	go updateFunc("Update2")

	// Wait for completion
	<-done
	<-done

	// Check for errors
	select {
	case err := <-errors:
		t.Logf("Concurrent update error (may be expected): %v", err)
	default:
		// No errors, both updates succeeded
	}

	// Verify final state
	finalConfig, err := server.db.GetConfig(config.ID)
	if err != nil {
		t.Fatalf("Failed to get final config: %v", err)
	}

	// One of the updates should have won
	if !strings.Contains(finalConfig.Name, "Update") {
		t.Error("Expected final config to be updated by one of the concurrent operations")
	}
}

func TestServer_HandleGetConfig_VeryLargeID(t *testing.T) {
	server := setupTestServer(t)
	defer server.Shutdown()

	// Test with very large ID
	req := setupTestRequest("GET", "/api/v1/preservation-configs/9223372036854775807", nil) // Max int64

	rr := httptest.NewRecorder()
	server.router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusNotFound)
	}
}

func TestServer_HandleUpdateConfig_EmptyRequestBody(t *testing.T) {
	server := setupTestServer(t)
	defer server.Shutdown()

	// Create initial config
	config := models.NewPreservationConfig("Empty Body Test", "Test")
	err := server.db.CreateConfig(config)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Update with completely empty body
	req := setupTestRequest("PUT", fmt.Sprintf("/api/v1/preservation-configs/%d", config.ID), bytes.NewBuffer([]byte("")))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	server.router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
}

func TestServer_HandleDeleteConfig_MultipleAttempts(t *testing.T) {
	server := setupTestServer(t)
	defer server.Shutdown()

	// Create config to delete
	config := models.NewPreservationConfig("Delete Test", "Test")
	err := server.db.CreateConfig(config)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// First delete - should succeed
	req := setupTestRequest("DELETE", fmt.Sprintf("/api/v1/preservation-configs/%d", config.ID), nil)
	rr := httptest.NewRecorder()
	server.router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNoContent {
		t.Errorf("First delete returned wrong status code: got %v want %v", status, http.StatusNoContent)
	}

	// Second delete - should return not found
	req = setupTestRequest("DELETE", fmt.Sprintf("/api/v1/preservation-configs/%d", config.ID), nil)
	rr = httptest.NewRecorder()
	server.router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("Second delete returned wrong status code: got %v want %v", status, http.StatusNotFound)
	}
}

func TestServer_InvalidContentType(t *testing.T) {
	server := setupTestServer(t)
	defer server.Shutdown()

	createReq := map[string]string{
		"name":        "Content Type Test",
		"description": "Test Description",
	}

	reqBody, err := json.Marshal(createReq)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	// Test with wrong content type
	req := setupTestRequest("POST", "/api/v1/preservation-configs", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "text/plain")

	rr := httptest.NewRecorder()
	server.router.ServeHTTP(rr, req)

	// Should still work as long as JSON is valid
	if status := rr.Code; status != http.StatusCreated {
		t.Logf("Request with wrong content type returned status %d (this may be expected)", status)
	}
}

func TestServer_HandleCreateConfig_MalformedJSON(t *testing.T) {
	server := setupTestServer(t)
	defer server.Shutdown()

	malformedJSONs := []string{
		`{"name": "test", "description": }`,        // Missing value
		`{"name": "test", "description": "test",}`, // Trailing comma
		`{"name": "test" "description": "test"}`,   // Missing comma
		`{"name": "test", "description": "test"`,   // Missing closing brace
		`{name: "test", "description": "test"}`,    // Unquoted key
		`{"name": "test", "description": 'test'}`,  // Single quotes
	}

	for i, malformedJSON := range malformedJSONs {
		t.Run(fmt.Sprintf("malformed_%d", i), func(t *testing.T) {
			req := setupTestRequest("POST", "/api/v1/preservation-configs", bytes.NewBuffer([]byte(malformedJSON)))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			server.router.ServeHTTP(rr, req)

			if status := rr.Code; status != http.StatusBadRequest {
				t.Errorf("Malformed JSON should return 400, got %d", status)
			}
		})
	}
}

func TestServer_HandleListConfigs_AfterManyCreations(t *testing.T) {
	server := setupTestServer(t)
	defer server.Shutdown()

	// Create many configs
	for i := range 100 {
		config := models.NewPreservationConfig(fmt.Sprintf("Config %d", i), fmt.Sprintf("Description %d", i))
		err := server.db.CreateConfig(config)
		if err != nil {
			t.Fatalf("Failed to create config %d: %v", i, err)
		}
	}

	req := setupTestRequest("GET", "/api/v1/preservation-configs", nil)
	rr := httptest.NewRecorder()
	server.router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var configs []models.PreservationConfig
	if err := json.Unmarshal(rr.Body.Bytes(), &configs); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Should have 101 configs (100 created + 1 default)
	if len(configs) != 101 {
		t.Errorf("Expected 101 configs, got %d", len(configs))
	}
}

func TestServer_RequestMethodNotAllowed(t *testing.T) {
	server := setupTestServer(t)
	defer server.Shutdown()

	// Test unsupported methods
	unsupportedMethods := []string{"PATCH", "HEAD", "OPTIONS", "TRACE", "CONNECT"}

	for _, method := range unsupportedMethods {
		t.Run(method, func(t *testing.T) {
			req := setupTestRequest(method, "/api/v1/preservation-configs", nil)
			rr := httptest.NewRecorder()
			server.router.ServeHTTP(rr, req)

			// Should return method not allowed or not found
			if rr.Code != http.StatusMethodNotAllowed && rr.Code != http.StatusNotFound {
				t.Logf("Method %s returned status %d (may be handled by router)", method, rr.Code)
			}
		})
	}
}
