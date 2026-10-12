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
