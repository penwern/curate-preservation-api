package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/penwern/curate-preservation-core-api/config"
	"github.com/penwern/curate-preservation-core-api/pkg/logger"
	"github.com/penwern/curate-preservation-core-api/server"
)

func main() {
	// Initialize logger first
	logger.Initialize("info")

	// Define command-line flags
	configFile := flag.String("config", "", "path to config file")
	dbType := flag.String("db", "sqlite3", "database type (sqlite3 or mysql)")
	dbConn := flag.String("conn", "preservation_configs.db", "database connection string")
	port := flag.Int("port", 6910, "port to run the server on")
	logLevel := flag.String("log-level", "info", "log level (debug, info, warn, error, fatal, panic)")
	flag.Parse()

	// Re-initialize logger with specified level
	logger.Initialize(*logLevel)

	// Load configuration
	cfg := config.Config{
		DBType:       *dbType,
		DBConnection: *dbConn,
		Port:         *port,
	}

	// Override with config file if provided
	if *configFile != "" {
		c, err := config.LoadFromFile(*configFile)
		if err != nil {
			logger.Fatal("Failed to load config: %v", err)
		}
		cfg = c
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
