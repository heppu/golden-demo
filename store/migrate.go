package store

import (
	"embed"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.sql
var fs embed.FS

func (s *Store) migrate() error {
	source, err := iofs.New(fs, "migrations")
	if err != nil {
		return fmt.Errorf("failed to create migration fs from embedded fs: %w", err)
	}

	driver, err := postgres.WithInstance(s.db.DB, &postgres.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithInstance("iofs", source, "postgres", driver)
	if err != nil {
		return err
	}

	dbInfo(m.Version())
	err = m.Up()
	if errors.Is(err, migrate.ErrNoChange) {
		slog.Info("already at latest migration")
		return nil
	}

	dbInfo(m.Version())
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func dbInfo(version uint, dirty bool, err error) {
	l := slog.With(slog.Uint64("current_version", uint64(version)))
	l = l.With(slog.Bool("dirty", dirty))
	if err != nil {
		l = l.With(slog.String("error", err.Error()))
	}
	l.Info("DB info")
}
