package repository

import (
	"context"
	"errors"
	"time"

	"github.com/dimakohanskyi/financist/internal/model"
	"gorm.io/gorm"
)

// transactionRow is the GORM mapping of the transactions table.
type transactionRow struct {
	ID          int64 `gorm:"primaryKey"`
	AccountID   int64
	ToAccountID *int64
	Type        string
	Amount      int64
	Description string
	Date        time.Time
}

func (transactionRow) TableName() string { return "transactions" }

type TransactionRepo struct{ db *gorm.DB }

func NewTransactionRepo(db *gorm.DB) *TransactionRepo { return &TransactionRepo{db} }

func (r *TransactionRepo) Create(ctx context.Context, t *model.Transaction) error {
	row := transactionRow(*t)
	if err := gorm.G[transactionRow](r.db).Create(ctx, &row); err != nil {
		return err
	}
	t.ID = row.ID
	return nil
}

func (r *TransactionRepo) GetByID(ctx context.Context, id int64) (model.Transaction, error) {
	row, err := gorm.G[transactionRow](r.db).Where("id = ?", id).First(ctx)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = model.ErrNotFound
	}
	return model.Transaction(row), err
}

// List returns one page (newest first) and the total number of matches.
func (r *TransactionRepo) List(ctx context.Context, f model.TransactionFilter) ([]model.Transaction, int64, error) {
	q := r.db.WithContext(ctx).Model(&transactionRow{})
	if f.AccountID != nil {
		q = q.Where("account_id = ? OR to_account_id = ?", *f.AccountID, *f.AccountID)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []transactionRow
	if err := q.Order("date DESC, id DESC").Limit(f.Limit).Offset(f.Offset).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	out := make([]model.Transaction, len(rows))
	for i, row := range rows {
		out[i] = model.Transaction(row)
	}
	return out, total, nil
}
