package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/dimakohanskyi/financist/internal/model"
	"github.com/dimakohanskyi/financist/internal/repository"
)

func TestSplitTransaction(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	accounts, categories := repository.NewAccountRepo(db), repository.NewCategoryRepo(db)
	if err := accounts.Create(ctx, &model.Account{Name: "Main", AccountType: "checking", Currency: "USD"}); err != nil {
		t.Fatal(err)
	}
	groceries := int64(2) // seeded expense category
	sweets := &model.Category{Name: "Sweets", Kind: model.CategoryExpense, Color: "#1baf7a", ParentID: &groceries}
	if err := categories.Create(ctx, sweets); err != nil {
		t.Fatal(err)
	}
	svc := NewTransactionService(repository.NewTransactionRepo(db), accounts, categories)

	tx := func(parts ...int64) *model.Transaction {
		return &model.Transaction{AccountID: 1, Type: model.TransactionExpense, Amount: 1000, Date: time.Now(), Splits: []model.TransactionSplit{
			{CategoryID: &groceries, Amount: parts[0]},
			{CategoryID: &sweets.ID, Amount: parts[1]},
		}}
	}
	ok := tx(700, 300)
	if err := svc.Create(ctx, ok); err != nil {
		t.Fatal(err)
	}
	if got, err := svc.Get(ctx, ok.ID); err != nil || len(got.Splits) != 2 {
		t.Fatalf("got %+v, err %v", got, err)
	}
	if err := svc.Create(ctx, tx(700, 200)); !errors.Is(err, ErrInvalid) {
		t.Fatalf("splits not adding up accepted, err: %v", err)
	}
}
