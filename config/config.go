package config

import (
	"encoding/json"
	"os"
)

// Config holds the server configuration
type Config struct {
	DBType       string `json:"db_type"`       // "sqlite3" or "mysql"
	DBConnection string `json:"db_connection"` // Connection string for the database
	Port         int    `json:"port"`          // Port for the HTTP server
}

// LoadFromFile loads configuration from a JSON file
func LoadFromFile(path string) (Config, error) {
	var cfg Config

	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}

	err = json.Unmarshal(data, &cfg)
	return cfg, err
}

// SaveToFile saves the configuration to a JSON file
func (c Config) SaveToFile(path string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
