package db

import (
	"database/sql"
	"embed"
	"strings"

	"github.com/glebarez/sqlite"
	"github.com/pressly/goose/v3"
	"gorm.io/gorm"
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

// Open connects to the database at dsn (foreign keys on) and migrates it to the latest version.
func Open(dsn string) (*gorm.DB, error) {
	sep := "?"
	if strings.Contains(dsn, "?") {
		sep = "&"
	}
	gormDB, err := gorm.Open(sqlite.Open(dsn+sep+"_pragma=foreign_keys(1)"), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Migrations get their own connection with foreign keys off: SQLite needs that
	// to rebuild a table other tables reference (https://sqlite.org/lang_altertable.html).
	migrationDB, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	defer migrationDB.Close()
	return gormDB, RunMigrations(migrationDB)
}
