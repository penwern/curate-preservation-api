package cmd

import (
	"os"
	"os/signal"
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

func runServer() {
	// Load configuration from viper
	cfg := config.Config{
		DBType:       viper.GetString("db.type"),
		DBConnection: viper.GetString("db.connection"),
		Port:         viper.GetInt("server.port"),
	}

	// Create and start the server
	srv, err := server.New(cfg)
	if err != nil {
		logger.Fatal("Failed to create server: %v", err)
	}

	// Start the server in a goroutine
	go func() {
		logger.Info("Starting API server on port %d", cfg.Port)
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
