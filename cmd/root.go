package cmd

import (
	"fmt"
	"os"

	"github.com/penwern/curate-preservation-api/pkg/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile          string
	dbType           string
	dbConn           string
	port             int
	siteDomain       string
	logLevel         string
	logFilePath      string
	allowInsecureTLS bool
	trustedIPs       []string
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
	rootCmd.PersistentFlags().StringVar(&siteDomain, "site-domain", "https://localhost:8080", "site domain for Pydio Cells OIDC and user endpoints")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "log level (debug, info, warn, error, fatal, panic)")
	rootCmd.PersistentFlags().StringVar(&logFilePath, "log-file", "", "log file path (default is /var/log/curate/curate-preservation-api.log)")
	rootCmd.PersistentFlags().BoolVar(&allowInsecureTLS, "allow-insecure-tls", false, "allow insecure TLS connections when making OIDC/Pydio requests")
	rootCmd.PersistentFlags().StringSliceVar(&trustedIPs, "trusted-ips", []string{}, "comma-separated list of trusted IP addresses/CIDR ranges that bypass authentication")

	// Bind flags to viper
	if err := viper.BindPFlag("db.type", rootCmd.PersistentFlags().Lookup("db-type")); err != nil {
		logger.Error("Failed to bind db.type flag: %v", err)
	}
	if err := viper.BindPFlag("db.connection", rootCmd.PersistentFlags().Lookup("db-connection")); err != nil {
		logger.Error("Failed to bind db.connection flag: %v", err)
	}
	if err := viper.BindPFlag("server.port", rootCmd.PersistentFlags().Lookup("port")); err != nil {
		logger.Error("Failed to bind server.port flag: %v", err)
	}
	if err := viper.BindPFlag("server.site_domain", rootCmd.PersistentFlags().Lookup("site-domain")); err != nil {
		logger.Error("Failed to bind server.site_domain flag: %v", err)
	}
	if err := viper.BindPFlag("log.level", rootCmd.PersistentFlags().Lookup("log-level")); err != nil {
		logger.Error("Failed to bind log.level flag: %v", err)
	}
	if err := viper.BindPFlag("log.file", rootCmd.PersistentFlags().Lookup("log-file")); err != nil {
		logger.Error("Failed to bind log.file flag: %v", err)
	}
	if err := viper.BindPFlag("server.allow_insecure_tls", rootCmd.PersistentFlags().Lookup("allow-insecure-tls")); err != nil {
		logger.Error("Failed to bind server.allow_insecure_tls flag: %v", err)
	}
	if err := viper.BindPFlag("server.trusted_ips", rootCmd.PersistentFlags().Lookup("trusted-ips")); err != nil {
		logger.Error("Failed to bind server.trusted_ips flag: %v", err)
	}
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

	// Initialize logger with the configured level and file path
	logLevel := viper.GetString("log.level")
	if logLevel == "" {
		logLevel = "info"
	}
	logFilePath := viper.GetString("log.file")
	logger.Initialize(logLevel, logFilePath)
}
