package cmd

import (
	"fmt"
	"os"

	"slices"

	"github.com/penwern/curate-preservation-core-api/config"
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
	Run: func(cmd *cobra.Command, args []string) {
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
			fmt.Printf("Error generating config file: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Configuration file generated: %s\n", filename)
	},
}

// configValidateCmd validates a configuration file
var configValidateCmd = &cobra.Command{
	Use:   "validate [filename]",
	Short: "Validate a configuration file",
	Long:  `Validate the syntax and values in a configuration file.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
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
			fmt.Printf("Error reading config file: %v\n", err)
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
			fmt.Printf("Error: Invalid database type '%s'. Must be 'sqlite3' or 'mysql'\n", cfg.DBType)
			os.Exit(1)
		}

		if cfg.Port < 1 || cfg.Port > 65535 {
			fmt.Printf("Error: Invalid port %d. Must be between 1 and 65535\n", cfg.Port)
			os.Exit(1)
		}

		if cfg.DBConnection == "" {
			fmt.Printf("Error: Database connection string cannot be empty\n")
			os.Exit(1)
		}

		logLevel := viper.GetString("log.level")
		validLogLevels := []string{"debug", "info", "warn", "error", "fatal", "panic"}
		validLevel := slices.Contains(validLogLevels, logLevel)
		if !validLevel {
			fmt.Printf("Error: Invalid log level '%s'. Must be one of: %v\n", logLevel, validLogLevels)
			os.Exit(1)
		}

		fmt.Printf("Configuration file is valid\n")
		fmt.Printf("Database Type: %s\n", cfg.DBType)
		fmt.Printf("Database Connection: %s\n", cfg.DBConnection)
		fmt.Printf("Server Port: %d\n", cfg.Port)
		fmt.Printf("Log Level: %s\n", logLevel)
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configGenerateCmd)
	configCmd.AddCommand(configValidateCmd)
}
