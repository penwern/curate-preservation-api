package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/penwern/curate-preservation-core-api/config"
	"github.com/penwern/curate-preservation-core-api/server"
)

func main() {
	// Define command-line flags
	configFile := flag.String("config", "", "path to config file")
	dbType := flag.String("db", "sqlite3", "database type (sqlite3 or mysql)")
	dbConn := flag.String("conn", "preservation_configs.db", "database connection string")
	port := flag.Int("port", 6910, "port to run the server on")
	flag.Parse()

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
			log.Fatalf("Failed to load config: %v", err)
		}
		cfg = c
	}

	// Create and start the server
	srv, err := server.New(cfg)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Start the server in a goroutine
	go func() {
		log.Printf("Starting API server on port %d", cfg.Port)
		if err := srv.Start(); err != nil {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	if err := srv.Shutdown(); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}
	log.Println("Server stopped")
}
