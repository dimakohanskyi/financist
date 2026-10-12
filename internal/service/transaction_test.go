package service

import (
	"context"
	"errors"
	"testing"

	"github.com/dimakohanskyi/financist/internal/model"
)

type fakeAccounts struct{ archived bool }

func (f fakeAccounts) GetByID(context.Context, int64) (model.Account, error) {
	return model.Account{IsArchived: f.archived}, nil
}

type fakeTxs struct{}

func (fakeTxs) Create(context.Context, *model.Transaction) error { return nil }
func (fakeTxs) GetByID(context.Context, int64) (model.Transaction, error) {
	return model.Transaction{}, nil
}

func TestCreate(t *testing.T) {
	tx := func() *model.Transaction {
		return &model.Transaction{AccountID: 1, Type: model.TransactionExpense, Amount: 100}
	}
	ctx := context.Background()

	if err := NewTransactionService(fakeTxs{}, fakeAccounts{}).Create(ctx, tx()); err != nil {
		t.Fatalf("valid transaction rejected: %v", err)
	}
	err := NewTransactionService(fakeTxs{}, fakeAccounts{archived: true}).Create(ctx, tx())
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("archived account accepted, err: %v", err)
	}
}
