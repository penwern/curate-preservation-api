// Package config provides the Config struct for application configuration.
package config

// Config holds the server configuration
// DBType: "sqlite3" or "mysql"
// DBConnection: Connection string for the database
// Port: Port for the HTTP server
// CORSOrigins: Allowed origins for CORS requests
// SiteDomain: Domain for Pydio Cells OIDC and user endpoints
// TrustedIPs: List of IP addresses/CIDR ranges that bypass authentication
// AllowInsecureTLS: Whether to allow insecure TLS connections when making OIDC/Pydio requests
type Config struct {
	DBType           string   `json:"db_type"`            // "sqlite3" or "mysql"
	DBConnection     string   `json:"db_connection"`      // Connection string for the database
	Port             int      `json:"port"`               // Port for the HTTP server
	CORSOrigins      []string `json:"cors_origins"`       // Allowed origins for CORS requests
	SiteDomain       string   `json:"site_domain"`        // Domain for Pydio Cells OIDC and user endpoints
	TrustedIPs       []string `json:"trusted_ips"`        // IP addresses/CIDR ranges that bypass authentication
	AllowInsecureTLS bool     `json:"allow_insecure_tls"` // Whether to allow insecure TLS connections
}
