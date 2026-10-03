//go:generate go run github.com/99designs/gqlgen generate

package graph

import sqlitedb "github.com/dimakohanskyi/financist/internal/db/sqlite/generated"

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require
// here.

type Resolver struct {
	DB *sqlitedb.Queries
}
