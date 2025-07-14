package config

import (
	"encoding/json"
	"testing"
)

func TestConfig_Defaults(t *testing.T) {
	config := Config{}

	// Test zero values
	if config.DBType != "" {
		t.Errorf("Expected empty DBType, got '%s'", config.DBType)
	}

	if config.DBConnection != "" {
		t.Errorf("Expected empty DBConnection, got '%s'", config.DBConnection)
	}

	if config.Port != 0 {
		t.Errorf("Expected Port to be 0, got %d", config.Port)
	}

	if config.SiteDomain != "" {
		t.Errorf("Expected empty SiteDomain, got '%s'", config.SiteDomain)
	}

	if config.AllowInsecureTLS != false {
		t.Errorf("Expected AllowInsecureTLS to be false, got %v", config.AllowInsecureTLS)
	}

	if config.CORSOrigins != nil {
		t.Errorf("Expected CORSOrigins to be nil, got %v", config.CORSOrigins)
	}

	if config.TrustedIPs != nil {
		t.Errorf("Expected TrustedIPs to be nil, got %v", config.TrustedIPs)
	}
}

func TestConfig_JSONMarshalUnmarshal(t *testing.T) {
	config := Config{
		DBType:           "sqlite3",
		DBConnection:     "/path/to/database.db",
		Port:             8080,
		CORSOrigins:      []string{"http://localhost:3000", "https://example.com"},
		SiteDomain:       "cells.example.com",
		TrustedIPs:       []string{"127.0.0.1", "192.168.1.0/24"},
		AllowInsecureTLS: true,
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("Failed to marshal config to JSON: %v", err)
	}

	// Unmarshal from JSON
	var unmarshaledConfig Config
	err = json.Unmarshal(jsonData, &unmarshaledConfig)
	if err != nil {
		t.Fatalf("Failed to unmarshal config from JSON: %v", err)
	}

	// Verify all fields are preserved
	if unmarshaledConfig.DBType != config.DBType {
		t.Errorf("Expected DBType '%s', got '%s'", config.DBType, unmarshaledConfig.DBType)
	}

	if unmarshaledConfig.DBConnection != config.DBConnection {
		t.Errorf("Expected DBConnection '%s', got '%s'", config.DBConnection, unmarshaledConfig.DBConnection)
	}

	if unmarshaledConfig.Port != config.Port {
		t.Errorf("Expected Port %d, got %d", config.Port, unmarshaledConfig.Port)
	}

	if unmarshaledConfig.SiteDomain != config.SiteDomain {
		t.Errorf("Expected SiteDomain '%s', got '%s'", config.SiteDomain, unmarshaledConfig.SiteDomain)
	}

	if unmarshaledConfig.AllowInsecureTLS != config.AllowInsecureTLS {
		t.Errorf("Expected AllowInsecureTLS %v, got %v", config.AllowInsecureTLS, unmarshaledConfig.AllowInsecureTLS)
	}

	// Verify slice fields
	if len(unmarshaledConfig.CORSOrigins) != len(config.CORSOrigins) {
		t.Errorf("Expected %d CORSOrigins, got %d", len(config.CORSOrigins), len(unmarshaledConfig.CORSOrigins))
	}

	for i, origin := range config.CORSOrigins {
		if unmarshaledConfig.CORSOrigins[i] != origin {
			t.Errorf("CORSOrigins[%d]: expected '%s', got '%s'", i, origin, unmarshaledConfig.CORSOrigins[i])
		}
	}

	if len(unmarshaledConfig.TrustedIPs) != len(config.TrustedIPs) {
		t.Errorf("Expected %d TrustedIPs, got %d", len(config.TrustedIPs), len(unmarshaledConfig.TrustedIPs))
	}

	for i, ip := range config.TrustedIPs {
		if unmarshaledConfig.TrustedIPs[i] != ip {
			t.Errorf("TrustedIPs[%d]: expected '%s', got '%s'", i, ip, unmarshaledConfig.TrustedIPs[i])
		}
	}
}

func TestConfig_EmptySlices(t *testing.T) {
	config := Config{
		DBType:       "mysql",
		DBConnection: "user:pass@tcp(localhost:3306)/dbname",
		Port:         9090,
		CORSOrigins:  []string{},
		TrustedIPs:   []string{},
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("Failed to marshal config with empty slices: %v", err)
	}

	// Unmarshal from JSON
	var unmarshaledConfig Config
	err = json.Unmarshal(jsonData, &unmarshaledConfig)
	if err != nil {
		t.Fatalf("Failed to unmarshal config with empty slices: %v", err)
	}

	// Empty slices should be preserved as empty slices, not nil
	if unmarshaledConfig.CORSOrigins == nil {
		t.Error("Expected CORSOrigins to be empty slice, got nil")
	}

	if len(unmarshaledConfig.CORSOrigins) != 0 {
		t.Errorf("Expected empty CORSOrigins, got %v", unmarshaledConfig.CORSOrigins)
	}

	if unmarshaledConfig.TrustedIPs == nil {
		t.Error("Expected TrustedIPs to be empty slice, got nil")
	}

	if len(unmarshaledConfig.TrustedIPs) != 0 {
		t.Errorf("Expected empty TrustedIPs, got %v", unmarshaledConfig.TrustedIPs)
	}
}

func TestConfig_DBTypes(t *testing.T) {
	validDBTypes := []string{"sqlite3", "mysql"}

	for _, dbType := range validDBTypes {
		config := Config{
			DBType:       dbType,
			DBConnection: "test-connection",
			Port:         8080,
		}

		// Test that the config can be marshaled/unmarshaled successfully
		jsonData, err := json.Marshal(config)
		if err != nil {
			t.Fatalf("Failed to marshal config with DBType '%s': %v", dbType, err)
		}

		var unmarshaledConfig Config
		err = json.Unmarshal(jsonData, &unmarshaledConfig)
		if err != nil {
			t.Fatalf("Failed to unmarshal config with DBType '%s': %v", dbType, err)
		}

		if unmarshaledConfig.DBType != dbType {
			t.Errorf("DBType not preserved: expected '%s', got '%s'", dbType, unmarshaledConfig.DBType)
		}
	}
}

func TestConfig_TrustedIPFormats(t *testing.T) {
	trustedIPs := []string{
		"127.0.0.1",
		"192.168.1.0/24",
		"10.0.0.0/8",
		"::1",
		"2001:db8::/32",
	}

	config := Config{
		DBType:       "sqlite3",
		DBConnection: ":memory:",
		Port:         8080,
		TrustedIPs:   trustedIPs,
	}

	// Test JSON round-trip
	jsonData, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("Failed to marshal config with various IP formats: %v", err)
	}

	var unmarshaledConfig Config
	err = json.Unmarshal(jsonData, &unmarshaledConfig)
	if err != nil {
		t.Fatalf("Failed to unmarshal config with various IP formats: %v", err)
	}

	// Verify all IP formats are preserved
	if len(unmarshaledConfig.TrustedIPs) != len(trustedIPs) {
		t.Errorf("Expected %d trusted IPs, got %d", len(trustedIPs), len(unmarshaledConfig.TrustedIPs))
	}

	for i, ip := range trustedIPs {
		if unmarshaledConfig.TrustedIPs[i] != ip {
			t.Errorf("TrustedIP[%d]: expected '%s', got '%s'", i, ip, unmarshaledConfig.TrustedIPs[i])
		}
	}
}

func TestConfig_CORSOriginFormats(t *testing.T) {
	corsOrigins := []string{
		"http://localhost:3000",
		"https://example.com",
		"https://subdomain.example.com:8443",
		"*",
	}

	config := Config{
		DBType:      "sqlite3",
		Port:        8080,
		CORSOrigins: corsOrigins,
	}

	// Test JSON round-trip
	jsonData, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("Failed to marshal config with various CORS origins: %v", err)
	}

	var unmarshaledConfig Config
	err = json.Unmarshal(jsonData, &unmarshaledConfig)
	if err != nil {
		t.Fatalf("Failed to unmarshal config with various CORS origins: %v", err)
	}

	// Verify all CORS origins are preserved
	if len(unmarshaledConfig.CORSOrigins) != len(corsOrigins) {
		t.Errorf("Expected %d CORS origins, got %d", len(corsOrigins), len(unmarshaledConfig.CORSOrigins))
	}

	for i, origin := range corsOrigins {
		if unmarshaledConfig.CORSOrigins[i] != origin {
			t.Errorf("CORSOrigin[%d]: expected '%s', got '%s'", i, origin, unmarshaledConfig.CORSOrigins[i])
		}
	}
}

func TestConfig_PortValues(t *testing.T) {
	testPorts := []int{0, 80, 443, 8080, 9090, 65535}

	for _, port := range testPorts {
		config := Config{
			DBType: "sqlite3",
			Port:   port,
		}

		jsonData, err := json.Marshal(config)
		if err != nil {
			t.Fatalf("Failed to marshal config with port %d: %v", port, err)
		}

		var unmarshaledConfig Config
		err = json.Unmarshal(jsonData, &unmarshaledConfig)
		if err != nil {
			t.Fatalf("Failed to unmarshal config with port %d: %v", port, err)
		}

		if unmarshaledConfig.Port != port {
			t.Errorf("Port not preserved: expected %d, got %d", port, unmarshaledConfig.Port)
		}
	}
}
