package db

import (
	"database/sql"
	"embed"

	"github.com/pressly/goose/v3"
)

// Bakes your .sql files into the compiled binary at build time, accessible as embedMigrations
//
//go:embed sqlite/migrations/*.sql
var embedMigrations embed.FS

func RunMigrations(db *sql.DB) error {
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("sqlite3"); err != nil {
		return err
	}

	return goose.Up(db, "sqlite/migrations")
}
