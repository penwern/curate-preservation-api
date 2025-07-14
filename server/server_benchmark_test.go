package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/penwern/curate-preservation-api/models"
	"github.com/penwern/curate-preservation-api/pkg/config"
)

func BenchmarkServer_HandleHealth(b *testing.B) {
	server := setupTestServerForBenchmark(b)
	defer server.Shutdown()

	for b.Loop() {
		req, _ := http.NewRequest("GET", "/api/v1/health", nil)
		rr := httptest.NewRecorder()
		server.router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			b.Fatalf("Health check failed")
		}
	}
}

func BenchmarkServer_HandleListConfigs(b *testing.B) {
	server := setupTestServerForBenchmark(b)
	defer server.Shutdown()

	// Pre-populate with some configs
	for i := range 10 {
		config := models.NewPreservationConfig(fmt.Sprintf("Benchmark Config %d", i), "Description")
		server.db.CreateConfig(config)
	}

	for b.Loop() {
		req := setupTestRequest("GET", "/api/v1/preservation-configs", nil)
		rr := httptest.NewRecorder()
		server.router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			b.Fatalf("List configs failed")
		}
	}
}

func BenchmarkServer_HandleCreateConfig(b *testing.B) {
	server := setupTestServerForBenchmark(b)
	defer server.Shutdown()

	createReq := map[string]string{
		"name":        "Benchmark Config",
		"description": "Benchmark Description",
	}

	reqBody, _ := json.Marshal(createReq)

	for b.Loop() {
		req := setupTestRequest("POST", "/api/v1/preservation-configs", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		server.router.ServeHTTP(rr, req)

		if rr.Code != http.StatusCreated {
			b.Fatalf("Create config failed")
		}
	}
}

func BenchmarkServer_HandleGetConfig(b *testing.B) {
	server := setupTestServerForBenchmark(b)
	defer server.Shutdown()

	// Create a config to get
	config := models.NewPreservationConfig("Benchmark Get Config", "Description")
	server.db.CreateConfig(config)

	for b.Loop() {
		req := setupTestRequest("GET", fmt.Sprintf("/api/v1/preservation-configs/%d", config.ID), nil)
		rr := httptest.NewRecorder()
		server.router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			b.Fatalf("Get config failed")
		}
	}
}

func BenchmarkServer_HandleUpdateConfig(b *testing.B) {
	server := setupTestServerForBenchmark(b)
	defer server.Shutdown()

	// Create a config to update
	config := models.NewPreservationConfig("Benchmark Update Config", "Description")
	server.db.CreateConfig(config)

	updateReq := map[string]string{
		"name":        "Updated Benchmark Config",
		"description": "Updated Description",
	}

	reqBody, _ := json.Marshal(updateReq)

	for b.Loop() {
		req := setupTestRequest("PUT", fmt.Sprintf("/api/v1/preservation-configs/%d", config.ID), bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		server.router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			b.Fatalf("Update config failed")
		}
	}
}

func BenchmarkServer_HandleCreateConfigWithA3M(b *testing.B) {
	server := setupTestServerForBenchmark(b)
	defer server.Shutdown()

	createReq := map[string]any{
		"name":        "Benchmark A3M Config",
		"description": "Benchmark with A3M settings",
		"a3m_config": map[string]any{
			"assign_uuids_to_directories":                       true,
			"examine_contents":                                  true,
			"generate_transfer_structure_report":                true,
			"document_empty_directories":                        true,
			"extract_packages":                                  true,
			"delete_packages_after_extraction":                  false,
			"identify_transfer":                                 true,
			"identify_submission_and_metadata":                  true,
			"identify_before_normalization":                     true,
			"normalize":                                         true,
			"transcribe_files":                                  true,
			"perform_policy_checks_on_originals":                true,
			"perform_policy_checks_on_preservation_derivatives": true,
			"perform_policy_checks_on_access_derivatives":       true,
			"thumbnail_mode":                                    1,
			"aip_compression_level":                             9,
			"aip_compression_algorithm":                         2,
		},
	}

	reqBody, _ := json.Marshal(createReq)

	for b.Loop() {
		req := setupTestRequest("POST", "/api/v1/preservation-configs", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		server.router.ServeHTTP(rr, req)

		if rr.Code != http.StatusCreated {
			b.Fatalf("Create config with A3M failed")
		}
	}
}

func BenchmarkServer_Authentication(b *testing.B) {
	server := setupTestServerForBenchmark(b)
	defer server.Shutdown()

	for b.Loop() {
		req := setupTestRequest("GET", "/api/v1/preservation-configs", nil)
		rr := httptest.NewRecorder()
		server.router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			b.Fatalf("Authentication failed")
		}
	}
}

func BenchmarkServer_JSONMarshaling(b *testing.B) {
	// Benchmark JSON marshaling of configurations
	config := &models.PreservationConfig{
		ID:          123,
		Name:        "Benchmark Config",
		Description: "This is a benchmark configuration for testing JSON marshaling performance",
		CompressAIP: true,
		A3MConfig:   models.NewA3MProcessingConfig(),
	}

	for b.Loop() {
		_, err := json.Marshal(config)
		if err != nil {
			b.Fatalf("JSON marshaling failed: %v", err)
		}
	}
}

func BenchmarkServer_JSONUnmarshaling(b *testing.B) {
	// Benchmark JSON unmarshaling of configurations
	jsonData := `{
		"id": 123,
		"name": "Benchmark Config",
		"description": "This is a benchmark configuration for testing JSON unmarshaling performance",
		"compress_aip": true,
		"a3m_config": {
			"assign_uuids_to_directories": true,
			"examine_contents": false,
			"generate_transfer_structure_report": true,
			"document_empty_directories": true,
			"extract_packages": true,
			"delete_packages_after_extraction": false,
			"identify_transfer": true,
			"identify_submission_and_metadata": true,
			"identify_before_normalization": true,
			"normalize": true,
			"transcribe_files": true,
			"perform_policy_checks_on_originals": true,
			"perform_policy_checks_on_preservation_derivatives": true,
			"perform_policy_checks_on_access_derivatives": true,
			"thumbnail_mode": 1,
			"aip_compression_level": 1,
			"aip_compression_algorithm": 2
		}
	}`

	for b.Loop() {
		var config models.PreservationConfig
		err := json.Unmarshal([]byte(jsonData), &config)
		if err != nil {
			b.Fatalf("JSON unmarshaling failed: %v", err)
		}
	}
}

// setupTestServerForBenchmark creates a test server optimized for benchmarks
func setupTestServerForBenchmark(b *testing.B) *Server {
	b.Helper()

	// Use in-memory database for faster benchmarks
	tmpDir := b.TempDir()
	dbPath := tmpDir + "/benchmark.db"

	cfg := config.Config{
		DBType:       "sqlite3",
		DBConnection: dbPath,
		Port:         8080,
		TrustedIPs:   []string{"127.0.0.1"},
	}

	server, err := New(cfg)
	if err != nil {
		b.Fatalf("Failed to create benchmark server: %v", err)
	}

	return server
}

func BenchmarkServer_ConcurrentRequests(b *testing.B) {
	server := setupTestServerForBenchmark(b)
	defer server.Shutdown()

	// Pre-populate with some configs
	for i := range 5 {
		config := models.NewPreservationConfig(fmt.Sprintf("Concurrent Config %d", i), "Description")
		server.db.CreateConfig(config)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := setupTestRequest("GET", "/api/v1/preservation-configs", nil)
			rr := httptest.NewRecorder()
			server.router.ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				b.Fatalf("Concurrent request failed")
			}
		}
	})
}
