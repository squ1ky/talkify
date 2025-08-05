package database

import (
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// MigrationManager handles database migrations
type MigrationManager struct {
	migrate *migrate.Migrate
}

// NewMigrationManager creates a new migration manager
func NewMigrationManager(db *DB) (*MigrationManager, error) {
	if db == nil || db.DB == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres driver: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres",
		driver,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create migrate instance: %v", err)
	}

	return &MigrationManager{migrate: m}, nil
}

// Up applies all pending migrations
func (mm *MigrationManager) Up() error {
	if err := mm.migrate.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	return nil
}

// Down rolls back one migration
func (mm *MigrationManager) Down() error {
	if err := mm.migrate.Steps(-1); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to rollback migration: %w", err)
	}

	return nil
}

// Version returns current migration version
func (mm *MigrationManager) Version() (uint, bool, error) {
	return mm.migrate.Version()
}

// Close closes the migration instance
func (mm *MigrationManager) Close() error {
	sourceErr, dbErr := mm.migrate.Close()
	if sourceErr != nil {
		return sourceErr
	}
	return dbErr
}

// Force sets the migration version without running migrations
// Use with caution - only for fixing broken migration status
func (mm *MigrationManager) Force(version int) error {
	if err := mm.migrate.Force(version); err != nil {
		return fmt.Errorf("failed to force migration: %w", err)
	}

	return nil
}
