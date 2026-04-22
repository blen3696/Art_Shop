// Package migrate wraps golang-migrate to run schema migrations at startup.
//
// Migrations live under this package's schema/ directory as pairs of
// <version>_<name>.up.sql / .down.sql files. They're embedded into the binary
// via //go:embed so production deploys don't need a separate migrations
// directory on disk or a CLI install — the server self-migrates on boot.
//
// Seed data lives under backend/migrations/seeds/ (outside this package) and
// is NOT run here. Seeds are one-shot fixtures applied manually via psql.
package migrate

import (
	"embed"
	"errors"
	"fmt"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"gorm.io/gorm"
)

// migrationFiles bundles every file under schema/ into the binary at build
// time. Must be in the same package as the //go:embed directive — that's why
// the SQL files live beside this file instead of at the project root.
//
//go:embed schema/*.sql
var migrationFiles embed.FS

// Up applies every pending migration to the database, in version order.
// Returns nil both when migrations ran AND when the DB was already current —
// "no change" is not an error.
func Up(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("migrate: get sql.DB: %w", err)
	}

	// The postgres driver for golang-migrate wants a *sql.DB, not a DSN, so
	// we reuse the connection GORM already has. This also means migrations
	// go through the same connection pool (important for Supabase pooler).
	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("migrate: build postgres driver: %w", err)
	}

	// iofs turns our embed.FS into a migrate.Source. "schema" is the directory
	// within the FS, matching the //go:embed pattern above.
	source, err := iofs.New(migrationFiles, "schema")
	if err != nil {
		return fmt.Errorf("migrate: build iofs source: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", source, "postgres", driver)
	if err != nil {
		return fmt.Errorf("migrate: init migrator: %w", err)
	}

	slog.Info("migrate: applying pending migrations")
	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			slog.Info("migrate: database already up to date")
			return nil
		}
		return fmt.Errorf("migrate: up: %w", err)
	}

	version, dirty, _ := m.Version()
	slog.Info("migrate: complete", "version", version, "dirty", dirty)
	return nil
}