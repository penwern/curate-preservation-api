// Package database provides database connectivity and operations for the preservation API.
package database

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"

	_ "github.com/go-sql-driver/mysql" // required for MySQL driver registration
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/mattn/go-sqlite3" // required for SQLite driver registration
	"github.com/penwern/curate-preservation-api/pkg/logger"
)

const (
	// DBTypeSQLite represents the SQLite database type
	DBTypeSQLite = "sqlite3"
	// DBTypeMySQL represents the MySQL database type
	DBTypeMySQL = "mysql"
)

// Embed migration files
//
//go:embed migrations/sqlite3/*.sql
var sqlite3Migrations embed.FS

//go:embed migrations/mysql/*.sql
var mysqlMigrations embed.FS

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

	// Use embedded migrations
	var migrationFS embed.FS
	var migrationPath string

	switch d.dbType {
	case DBTypeSQLite:
		migrationFS = sqlite3Migrations
		migrationPath = "migrations/sqlite3"
	case DBTypeMySQL:
		migrationFS = mysqlMigrations
		migrationPath = "migrations/mysql"
	}

	sourceDriver, err := iofs.New(migrationFS, migrationPath)
	if err != nil {
		return fmt.Errorf("failed to create iofs source driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", sourceDriver, d.dbType, driver)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}
