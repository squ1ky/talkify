package database

import (
	"database/sql"
	"fmt"
	"github.com/squ1ky/talkify/internal/config"
	"time"
)

// DB wraps sql.DB to provide additional functionality
type DB struct {
	*sql.DB
}

// Connect establishes connection to PostgreSQL database
func Connect(cfg *config.DatabaseConfig) (*DB, error) {
	db, err := sql.Open("postgres", cfg.GetDSN())
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{DB: db}, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	if db.DB != nil {
		return db.DB.Close()
	}

	return nil
}

// Ping tests database connection
func (db *DB) Ping() error {
	if db.DB == nil {
		return fmt.Errorf("database connection is nil")
	}

	return db.DB.Ping()
}

// BeginTx starts a database transaction
func (db *DB) BeginTx() (*sql.Tx, error) {
	return db.DB.Begin()
}

// Health checks database health status
func (db *DB) Health() error {
	if err := db.Ping(); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	var result int
	err := db.QueryRow("SELECT 1").Scan(&result)
	if err != nil {
		return fmt.Errorf("database query test failed: %w", err)
	}

	if result != 1 {
		return fmt.Errorf("database query returned unexpected result: %d", result)
	}

	return nil
}
