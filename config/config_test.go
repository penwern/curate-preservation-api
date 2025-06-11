package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestConfig_LoadFromFile(t *testing.T) {
	tests := []struct {
		name        string
		fileContent string
		want        Config
		wantErr     bool
	}{
		{
			name: "valid config file",
			fileContent: `{
				"db_type": "mysql",
				"db_connection": "user:pass@tcp(localhost:3306)/testdb",
				"port": 8080
			}`,
			want: Config{
				DBType:       "mysql",
				DBConnection: "user:pass@tcp(localhost:3306)/testdb",
				Port:         8080,
			},
			wantErr: false,
		},
		{
			name: "config with defaults",
			fileContent: `{
				"db_type": "sqlite3",
				"db_connection": "test.db"
			}`,
			want: Config{
				DBType:       "sqlite3",
				DBConnection: "test.db",
				Port:         0,
			},
			wantErr: false,
		},
		{
			name:        "invalid json",
			fileContent: `{"invalid": json}`,
			want:        Config{},
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "config.json")

			err := os.WriteFile(tmpFile, []byte(tt.fileContent), 0644)
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}

			got, err := LoadFromFile(tmpFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadFromFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("LoadFromFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_LoadFromFile_NonExistentFile(t *testing.T) {
	_, err := LoadFromFile("non-existent-file.json")
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}

func TestConfig_SaveToFile(t *testing.T) {
	config := Config{
		DBType:       "sqlite3",
		DBConnection: "test.db",
		Port:         9000,
	}

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test-config.json")

	err := config.SaveToFile(tmpFile)
	if err != nil {
		t.Fatalf("SaveToFile() error = %v", err)
	}

	// Verify file was created and contains correct content
	data, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to read saved file: %v", err)
	}

	var savedConfig Config
	err = json.Unmarshal(data, &savedConfig)
	if err != nil {
		t.Fatalf("Failed to unmarshal saved config: %v", err)
	}

	if savedConfig != config {
		t.Errorf("Saved config = %v, want %v", savedConfig, config)
	}
}

func TestConfig_SaveToFile_InvalidPath(t *testing.T) {
	config := Config{
		DBType:       "sqlite3",
		DBConnection: "test.db",
		Port:         9000,
	}

	// Try to save to a directory that doesn't exist and can't be created
	err := config.SaveToFile("/invalid/path/config.json")
	if err == nil {
		t.Error("Expected error for invalid path, got nil")
	}
}
