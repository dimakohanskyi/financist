//go:generate go run github.com/99designs/gqlgen generate

package graph

import (
	"context"

	"github.com/dimakohanskyi/financist/internal/graph/model"
	"github.com/dimakohanskyi/financist/internal/repository"
	"github.com/dimakohanskyi/financist/internal/service"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require
// here.

type Resolver struct {
	AccountRepo  *repository.AccountRepo
	TxService    *service.TransactionService
	CategorySvc  *service.CategoryService
	TagRepo      *repository.TagRepo
	Recurring    *service.RecurringService
	BudgetSvc    *service.BudgetService
	GoalRepo     *repository.GoalRepo
	AssetRepo    *repository.AssetRepo
	SettingsRepo *repository.SettingsRepo
}

// category loads a category for a nested field. ponytail: one query per
// parent object; add a dataloader if lists get long.
func (r *Resolver) category(ctx context.Context, id *string) (*model.Category, error) {
	n, err := parseOptionalID(id)
	if err != nil || n == nil {
		return nil, err
	}
	c, err := r.CategorySvc.Get(ctx, *n)
	if err != nil {
		return nil, err
	}
	return toCategory(&c), nil
}
