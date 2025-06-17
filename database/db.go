// Package database provides database connectivity and operations for the preservation API.
package database

import (
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"runtime"

	_ "github.com/go-sql-driver/mysql" // required for MySQL driver registration
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file" // required for file-based migrations
	_ "github.com/mattn/go-sqlite3"                      // required for SQLite driver registration
	"github.com/penwern/curate-preservation-api/pkg/logger"
)

const (
	// DBTypeSQLite represents the SQLite database type
	DBTypeSQLite = "sqlite3"
	// DBTypeMySQL represents the MySQL database type
	DBTypeMySQL = "mysql"
)

// Database represents a database connection
type Database struct {
	db     *sql.DB
	dbType string
}

// New creates a new database connection
func New(dbType, connString string) (*Database, error) {
	if dbType != DBTypeSQLite && dbType != DBTypeMySQL {
		return nil, errors.New("unsupported database type, must be 'sqlite3' or 'mysql'")
	}

	logger.Info("Connecting to %s database: %s", dbType, connString)
	db, err := sql.Open(dbType, connString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Successfully connected to %s database", dbType)

	database := &Database{
		db:     db,
		dbType: dbType,
	}

	// Run migrations
	logger.Info("Running database migrations...")
	if err := database.runMigrations(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	logger.Info("Database migrations completed successfully")
	return database, nil
}

// Close closes the database connection
func (d *Database) Close() error {
	return d.db.Close()
}

// getMigrationsPath returns the absolute path to the migrations directory
func getMigrationsPath(dbType string) (string, error) {
	// Get the current file's directory
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "", errors.New("failed to get current file path")
	}

	// Get the directory containing this file (database package)
	currentDir := filepath.Dir(filename)

	// Build path to migrations directory
	migrationsDir := filepath.Join(currentDir, "migrations", dbType)

	// Convert to file:// URL format
	return "file://" + filepath.ToSlash(migrationsDir), nil
}

// runMigrations runs all pending database migrations
func (d *Database) runMigrations() error {
	var driver database.Driver
	var err error

	switch d.dbType {
	case DBTypeSQLite:
		driver, err = sqlite3.WithInstance(d.db, &sqlite3.Config{})
		if err != nil {
			return fmt.Errorf("failed to create sqlite3 driver: %w", err)
		}
	case DBTypeMySQL:
		driver, err = mysql.WithInstance(d.db, &mysql.Config{})
		if err != nil {
			return fmt.Errorf("failed to create mysql driver: %w", err)
		}
	default:
		return errors.New("unsupported database type for migrations")
	}

	migrationsPath, err := getMigrationsPath(d.dbType)
	if err != nil {
		return fmt.Errorf("failed to get migrations path: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		migrationsPath,
		d.dbType,
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}
