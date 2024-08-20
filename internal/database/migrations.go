package database

import (
	"database/sql"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/rs/zerolog/log"
)

func RunMigrations(db *sql.DB, driverName, dbName string, migrationPaths []string) error {
	var driver database.Driver
	var err error
	switch driverName {
	case "postgres":
		driver, err = postgres.WithInstance(db, &postgres.Config{})
	case "sqlite3":
		driver, err = sqlite3.WithInstance(db, &sqlite3.Config{})
	default:
		return fmt.Errorf("unsupported database driver: %s", driverName)
	}
	if err != nil {
		return fmt.Errorf("failed to create database driver: %w", err)
	}
	for _, migrationsDir := range migrationPaths {
		m, err := migrate.NewWithDatabaseInstance(
			fmt.Sprintf("file://%s", migrationsDir),
			dbName,
			driver,
		)
		if err != nil {
			return fmt.Errorf("failed to create migrate instance for %s: %w", migrationsDir, err)
		}
		version, dirty, err := m.Version()
		if err != nil && err != migrate.ErrNilVersion {
			return fmt.Errorf("failed to get migration version for %s: %w", migrationsDir, err)
		}
		if dirty {
			log.Warn().Msgf("Database is in a dirty state for %s. Forcing version %d", migrationsDir, version)
			if err := m.Force(int(version)); err != nil {
				return fmt.Errorf("failed to force version for %s: %w", migrationsDir, err)
			}
		}
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			return fmt.Errorf("failed to apply migrations for %s: %w", migrationsDir, err)
		}
		log.Info().Msgf("Migrations applied successfully for %s", migrationsDir)
	}
	return nil
}
