//go:generate go run github.com/99designs/gqlgen generate

package graph

import (
	"github.com/dimakohanskyi/financist/internal/repository"
	"github.com/dimakohanskyi/financist/internal/service"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require
// here.

type Resolver struct {
	AccountRepo *repository.AccountRepo
	TxService   *service.TransactionService
}
