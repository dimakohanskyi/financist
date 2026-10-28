package repository

import (
	"context"
	"errors"

	"github.com/dimakohanskyi/financist/internal/model"
	"gorm.io/gorm"
)

// budgetRow is the GORM mapping of the budgets table.
type budgetRow struct {
	ID           int64 `gorm:"primaryKey"`
	CategoryID   int64
	MonthlyLimit int64
}

func (budgetRow) TableName() string { return "budgets" }

type BudgetRepo struct{ db *gorm.DB }

func NewBudgetRepo(db *gorm.DB) *BudgetRepo { return &BudgetRepo{db} }

func (r *BudgetRepo) Create(ctx context.Context, b *model.Budget) error {
	row := budgetRow(*b)
	if err := gorm.G[budgetRow](r.db).Create(ctx, &row); err != nil {
		return err
	}
	b.ID = row.ID
	return nil
}

func (r *BudgetRepo) GetByID(ctx context.Context, id int64) (model.Budget, error) {
	row, err := gorm.G[budgetRow](r.db).Where("id = ?", id).First(ctx)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = model.ErrNotFound
	}
	return model.Budget(row), err
}

// List returns budgets in their categories' display order.
func (r *BudgetRepo) List(ctx context.Context) ([]model.Budget, error) {
	var rows []budgetRow
	err := r.db.WithContext(ctx).Joins("JOIN categories c ON c.id = budgets.category_id").
		Order("c.sort_order, c.id").Find(&rows).Error
	out := make([]model.Budget, len(rows))
	for i, row := range rows {
		out[i] = model.Budget(row)
	}
	return out, err
}

func (r *BudgetRepo) UpdateLimit(ctx context.Context, id, limit int64) (model.Budget, error) {
	if err := r.db.WithContext(ctx).Model(&budgetRow{}).Where("id = ?", id).Update("monthly_limit", limit).Error; err != nil {
		return model.Budget{}, err
	}
	return r.GetByID(ctx, id)
}

func (r *BudgetRepo) Delete(ctx context.Context, id int64) error {
	res := r.db.WithContext(ctx).Delete(&budgetRow{}, id)
	if res.Error == nil && res.RowsAffected == 0 {
		return model.ErrNotFound
	}
	return res.Error
}
