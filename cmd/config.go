// Package cmd provides command-line interface commands for the preservation API.
package cmd

import (
	"os"
	"slices"

	"github.com/penwern/curate-preservation-api/config"
	"github.com/penwern/curate-preservation-api/pkg/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration management commands",
	Long:  `Commands for managing configuration files and settings.`,
}

// configGenerateCmd generates a sample configuration file
var configGenerateCmd = &cobra.Command{
	Use:   "generate [filename]",
	Short: "Generate a sample configuration file",
	Long: `Generate a sample configuration file with default values.
	
If no filename is provided, it will create .preservation-api.yaml in the current directory.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		filename := ".preservation-api.yaml"
		if len(args) > 0 {
			filename = args[0]
		}

		// Set default values
		viper.SetDefault("db.type", "sqlite3")
		viper.SetDefault("db.connection", "preservation_configs.db")
		viper.SetDefault("server.port", 6910)
		viper.SetDefault("log.level", "info")

		// Write config file
		err := viper.WriteConfigAs(filename)
		if err != nil {
			logger.Error("Error generating config file: %v", err)
			os.Exit(1)
		}

		logger.Info("Configuration file generated: %s", filename)
	},
}

// configValidateCmd validates a configuration file
var configValidateCmd = &cobra.Command{
	Use:   "validate [filename]",
	Short: "Validate a configuration file",
	Long:  `Validate the syntax and values in a configuration file.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		filename := ""
		if len(args) > 0 {
			filename = args[0]
		}

		// Try to load the config
		if filename != "" {
			viper.SetConfigFile(filename)
		}

		err := viper.ReadInConfig()
		if err != nil {
			logger.Error("Error reading config file: %v", err)
			os.Exit(1)
		}

		// Validate the configuration
		cfg := config.Config{
			DBType:       viper.GetString("db.type"),
			DBConnection: viper.GetString("db.connection"),
			Port:         viper.GetInt("server.port"),
		}

		// Basic validation
		if cfg.DBType != "sqlite3" && cfg.DBType != "mysql" {
			logger.Error("Error: Invalid database type '%s'. Must be 'sqlite3' or 'mysql'", cfg.DBType)
			os.Exit(1)
		}

		if cfg.Port < 1 || cfg.Port > 65535 {
			logger.Error("Error: Invalid port %d. Must be between 1 and 65535", cfg.Port)
			os.Exit(1)
		}

		if cfg.DBConnection == "" {
			logger.Error("Error: Database connection string cannot be empty")
			os.Exit(1)
		}

		logLevel := viper.GetString("log.level")
		validLogLevels := []string{"debug", "info", "warn", "error", "fatal", "panic"}
		validLevel := slices.Contains(validLogLevels, logLevel)
		if !validLevel {
			logger.Error("Error: Invalid log level '%s'. Must be one of: %v", logLevel, validLogLevels)
			os.Exit(1)
		}

		logger.Info("Configuration file is valid")
		logger.Info("Database Type: %s", cfg.DBType)
		logger.Info("Database Connection: %s", cfg.DBConnection)
		logger.Info("Server Port: %d", cfg.Port)
		logger.Info("Log Level: %s", logLevel)
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configGenerateCmd)
	configCmd.AddCommand(configValidateCmd)
}
