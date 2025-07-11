package cmd

import (
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/penwern/curate-preservation-api/pkg/config"
	"github.com/penwern/curate-preservation-api/pkg/logger"
	"github.com/penwern/curate-preservation-api/server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the API server",
	Long: `Start the preservation API server with the specified configuration.
	
The server will listen on the configured port and handle REST API requests
for managing preservation configurations and workflows.`,
	Run: func(_ *cobra.Command, _ []string) {
		runServer()
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}

// getStringSlice handles viper's limitation with comma-separated environment variables
func getStringSlice(key string) []string {
	slice := viper.GetStringSlice(key)

	// If we got a slice with one element that contains commas, split it
	if len(slice) == 1 && strings.Contains(slice[0], ",") {
		parts := strings.Split(slice[0], ",")
		result := make([]string, len(parts))
		for i, part := range parts {
			result[i] = strings.TrimSpace(part)
		}
		return result
	}

	return slice
}

func runServer() {
	// Load configuration from viper
	cfg := config.Config{
		DBType:           viper.GetString("db.type"),
		DBConnection:     viper.GetString("db.connection"),
		Port:             viper.GetInt("server.port"),
		SiteDomain:       viper.GetString("server.site_domain"),
		AllowInsecureTLS: viper.GetBool("server.allow_insecure_tls"),
		TrustedIPs:       getStringSlice("server.trusted_ips"),
	}

	// Create and start the server
	srv, err := server.New(cfg)
	if err != nil {
		logger.Fatal("Failed to create server: %v", err)
	}

	// Start the server in a goroutine
	go func() {
		logger.Info("===========================================")
		logger.Info("Starting API server on port %d", cfg.Port)
		logger.Info("Cells Site Domain: %s", cfg.SiteDomain)
		logger.Info("Allow Insecure TLS: %v", cfg.AllowInsecureTLS)
		if len(cfg.TrustedIPs) > 0 {
			logger.Info("Trusted IPs configured: %v", cfg.TrustedIPs)
		} else {
			logger.Info("No trusted IPs configured - all requests require authentication")
		}
		if err := srv.Start(); err != nil {
			logger.Fatal("Server failed: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")
	if err := srv.Shutdown(); err != nil {
		logger.Fatal("Server shutdown failed: %v", err)
	}
	logger.Info("Server stopped")
}
