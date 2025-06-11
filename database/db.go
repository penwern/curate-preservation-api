package database

import (
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
)

// Database represents a database connection
type Database struct {
	db     *sql.DB
	dbType string
}

// New creates a new database connection
func New(dbType, connString string) (*Database, error) {
	if dbType != "sqlite3" && dbType != "mysql" {
		return nil, errors.New("unsupported database type, must be 'sqlite3' or 'mysql'")
	}

	db, err := sql.Open(dbType, connString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	database := &Database{
		db:     db,
		dbType: dbType,
	}

	// Initialize database tables
	if err := database.initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	return database, nil
}

// Close closes the database connection
func (d *Database) Close() error {
	return d.db.Close()
}

// initialize creates necessary tables if they don't exist
func (d *Database) initialize() error {
	var createTableSQL string

	if d.dbType == "sqlite3" {
		createTableSQL = `
		CREATE TABLE IF NOT EXISTS preservation_configs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			description TEXT,
			assign_uuids_to_directories BOOLEAN DEFAULT TRUE,
			examine_contents BOOLEAN DEFAULT FALSE,
			generate_transfer_structure_report BOOLEAN DEFAULT TRUE,
			document_empty_directories BOOLEAN DEFAULT TRUE,
			extract_packages BOOLEAN DEFAULT TRUE,
			delete_packages_after_extraction BOOLEAN DEFAULT FALSE,
			identify_transfer BOOLEAN DEFAULT TRUE,
			identify_submission_and_metadata BOOLEAN DEFAULT TRUE,
			identify_before_normalization BOOLEAN DEFAULT TRUE,
			normalize BOOLEAN DEFAULT TRUE,
			transcribe_files BOOLEAN DEFAULT TRUE,
			perform_policy_checks_on_originals BOOLEAN DEFAULT TRUE,
			perform_policy_checks_on_preservation_derivatives BOOLEAN DEFAULT TRUE,
			perform_policy_checks_on_access_derivatives BOOLEAN DEFAULT TRUE,
			thumbnail_mode INT DEFAULT 1,
			aip_compression_level INT DEFAULT 1,
			aip_compression_algorithm INT DEFAULT 5,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`
	} else if d.dbType == "mysql" {
		createTableSQL = `
		CREATE TABLE IF NOT EXISTS preservation_configs (
			id INT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			assign_uuids_to_directories BOOLEAN DEFAULT TRUE,
			examine_contents BOOLEAN DEFAULT FALSE,
			generate_transfer_structure_report BOOLEAN DEFAULT TRUE,
			document_empty_directories BOOLEAN DEFAULT TRUE,
			extract_packages BOOLEAN DEFAULT TRUE,
			delete_packages_after_extraction BOOLEAN DEFAULT FALSE,
			identify_transfer BOOLEAN DEFAULT TRUE,
			identify_submission_and_metadata BOOLEAN DEFAULT TRUE,
			identify_before_normalization BOOLEAN DEFAULT TRUE,
			normalize BOOLEAN DEFAULT TRUE,
			transcribe_files BOOLEAN DEFAULT TRUE,
			perform_policy_checks_on_originals BOOLEAN DEFAULT TRUE,
			perform_policy_checks_on_preservation_derivatives BOOLEAN DEFAULT TRUE,
			perform_policy_checks_on_access_derivatives BOOLEAN DEFAULT TRUE,
			thumbnail_mode INT DEFAULT 1,
			aip_compression_level INT DEFAULT 1,
			aip_compression_algorithm INT DEFAULT 5,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		);`
	}

	_, err := d.db.Exec(createTableSQL)
	if err != nil {
		return err
	}

	// Create trigger for SQLite to auto-update updated_at field
	if d.dbType == "sqlite3" {
		triggerSQL := `
		CREATE TRIGGER IF NOT EXISTS update_preservation_configs_updated_at
		AFTER UPDATE ON preservation_configs
		BEGIN
			UPDATE preservation_configs SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
		END;`

		_, err = d.db.Exec(triggerSQL)
		if err != nil {
			return err
		}
	}

	// Create default configuration entry
	if err := d.createDefaultConfig(); err != nil {
		return fmt.Errorf("failed to create default configuration: %w", err)
	}

	return nil
}

// createDefaultConfig creates a default preservation configuration if none exists
func (d *Database) createDefaultConfig() error {
	// Check if any configs exist
	var count int
	countQuery := "SELECT COUNT(*) FROM preservation_configs"
	if err := d.db.QueryRow(countQuery).Scan(&count); err != nil {
		return err
	}

	// If configs already exist, don't create default
	if count > 0 {
		return nil
	}

	// Create default configuration
	defaultConfigSQL := `
	INSERT INTO preservation_configs (
		name, description,
		assign_uuids_to_directories, examine_contents, generate_transfer_structure_report,
		document_empty_directories, extract_packages, delete_packages_after_extraction,
		identify_transfer, identify_submission_and_metadata, identify_before_normalization,
		normalize, transcribe_files, perform_policy_checks_on_originals,
		perform_policy_checks_on_preservation_derivatives, perform_policy_checks_on_access_derivatives,
		thumbnail_mode, aip_compression_level, aip_compression_algorithm
	) VALUES (
		'Default Configuration', 'Default preservation configuration with recommended settings',
		true, false, true,
		true, true, false,
		true, true, true,
		true, true, true,
		true, true,
		1, 1, 5
	)`

	_, err := d.db.Exec(defaultConfigSQL)
	return err
}
