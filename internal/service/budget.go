package service

import (
	"context"
	"fmt"

	"github.com/dimakohanskyi/financist/internal/model"
)

type BudgetStore interface {
	Create(ctx context.Context, b *model.Budget) error
	List(ctx context.Context) ([]model.Budget, error)
	UpdateLimit(ctx context.Context, id, limit int64) (model.Budget, error)
	Delete(ctx context.Context, id int64) error
}

type CategoryLister interface {
	GetByID(ctx context.Context, id int64) (model.Category, error)
	List(ctx context.Context, kind *string) ([]model.Category, error)
}

type ContributionSource interface {
	Contributions(ctx context.Context, f model.ContributionFilter) ([]model.Contribution, error)
}

type BudgetService struct {
	store      BudgetStore
	categories CategoryLister
	txs        ContributionSource
}

func NewBudgetService(store BudgetStore, categories CategoryLister, txs ContributionSource) *BudgetService {
	return &BudgetService{store, categories, txs}
}

func (s *BudgetService) Create(ctx context.Context, b *model.Budget) error {
	if b.MonthlyLimit <= 0 {
		return fmt.Errorf("%w: monthly limit must be positive", ErrInvalid)
	}
	c, err := s.categories.GetByID(ctx, b.CategoryID)
	if err != nil {
		return err
	}
	if c.Kind != model.CategoryExpense {
		return fmt.Errorf("%w: budgets are for expense categories", ErrInvalid)
	}
	return s.store.Create(ctx, b)
}

func (s *BudgetService) List(ctx context.Context) ([]model.Budget, error) {
	return s.store.List(ctx)
}

func (s *BudgetService) UpdateLimit(ctx context.Context, id, limit int64) (model.Budget, error) {
	if limit <= 0 {
		return model.Budget{}, fmt.Errorf("%w: monthly limit must be positive", ErrInvalid)
	}
	return s.store.UpdateLimit(ctx, id, limit)
}

func (s *BudgetService) Delete(ctx context.Context, id int64) error {
	return s.store.Delete(ctx, id)
}

// Status compares each budget with the month's spending. A budget covers its
// category's subcategories too; a subcategory with its own budget is tracked
// against both.
func (s *BudgetService) Status(ctx context.Context, year, month int) ([]model.BudgetStatus, error) {
	budgets, err := s.store.List(ctx)
	if err != nil || len(budgets) == 0 {
		return nil, err
	}
	categories, err := s.categories.List(ctx, nil)
	if err != nil {
		return nil, err
	}
	from, to := model.MonthRange(year, month)
	contributions, err := s.txs.Contributions(ctx, model.ContributionFilter{Type: model.TransactionExpense, From: &from, To: &to})
	if err != nil {
		return nil, err
	}

	spent := map[int64]int64{} // by category and, rolled up, by its parent
	parentOf := map[int64]int64{}
	for _, c := range categories {
		if c.ParentID != nil {
			parentOf[c.ID] = *c.ParentID
		}
	}
	for _, c := range contributions {
		spent[c.CategoryID] += c.Amount
		if parent, ok := parentOf[c.CategoryID]; ok {
			spent[parent] += c.Amount
		}
	}

	out := make([]model.BudgetStatus, len(budgets))
	for i, b := range budgets {
		out[i] = model.BudgetStatus{Budget: b, Spent: spent[b.CategoryID]}
	}
	return out, nil
}
