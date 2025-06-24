// Package config provides the Config struct for application configuration.
package config

// Config holds the server configuration
// DBType: "sqlite3" or "mysql"
// DBConnection: Connection string for the database
// Port: Port for the HTTP server
// CORSOrigins: Allowed origins for CORS requests
type Config struct {
	DBType       string   `json:"db_type"`       // "sqlite3" or "mysql"
	DBConnection string   `json:"db_connection"` // Connection string for the database
	Port         int      `json:"port"`          // Port for the HTTP server
	CORSOrigins  []string `json:"cors_origins"`  // Allowed origins for CORS requests
}
