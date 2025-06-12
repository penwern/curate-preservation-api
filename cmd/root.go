package cmd

import (
	"fmt"
	"os"

	"github.com/penwern/curate-preservation-core-api/pkg/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile  string
	dbType   string
	dbConn   string
	port     int
	logLevel string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "preservation-api",
	Short: "Curate Preservation Core API Server",
	Long: `A REST API server for managing preservation configurations and workflows.
	
This application provides endpoints for creating, reading, updating, and deleting
preservation configurations, as well as managing preservation workflows.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.preservation-api.yaml)")
	rootCmd.PersistentFlags().StringVar(&dbType, "db-type", "sqlite3", "database type (sqlite3 or mysql)")
	rootCmd.PersistentFlags().StringVar(&dbConn, "db-connection", "preservation_configs.db", "database connection string")
	rootCmd.PersistentFlags().IntVar(&port, "port", 6910, "port to run the server on")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "log level (debug, info, warn, error, fatal, panic)")

	// Bind flags to viper
	viper.BindPFlag("db.type", rootCmd.PersistentFlags().Lookup("db-type"))
	viper.BindPFlag("db.connection", rootCmd.PersistentFlags().Lookup("db-connection"))
	viper.BindPFlag("server.port", rootCmd.PersistentFlags().Lookup("port"))
	viper.BindPFlag("log.level", rootCmd.PersistentFlags().Lookup("log-level"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".preservation-api" (without extension).
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName(".preservation-api")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}

	// Initialize logger with the configured level
	logLevel := viper.GetString("log.level")
	if logLevel == "" {
		logLevel = "info"
	}
	logger.Initialize(logLevel)
}
