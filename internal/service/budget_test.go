package service

import (
	"context"
	"testing"
	"time"

	"github.com/dimakohanskyi/financist/internal/model"
	"github.com/dimakohanskyi/financist/internal/repository"
)

func TestBudgetStatusRollsUpSubcategoriesAndSplits(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	accounts, categories, txRepo := repository.NewAccountRepo(db), repository.NewCategoryRepo(db), repository.NewTransactionRepo(db)
	if err := accounts.Create(ctx, &model.Account{Name: "Main", AccountType: "checking", Currency: "USD"}); err != nil {
		t.Fatal(err)
	}
	groceries := int64(2) // seeded expense category
	sweets := &model.Category{Name: "Sweets", Kind: model.CategoryExpense, Color: "#1baf7a", ParentID: &groceries}
	if err := categories.Create(ctx, sweets); err != nil {
		t.Fatal(err)
	}
	svc := NewBudgetService(repository.NewBudgetRepo(db), categories, txRepo)
	if err := svc.Create(ctx, &model.Budget{CategoryID: groceries, MonthlyLimit: 500}); err != nil {
		t.Fatal(err)
	}

	now := time.Now().UTC()
	for _, tx := range []*model.Transaction{
		{AccountID: 1, Type: model.TransactionExpense, Amount: 300, CategoryID: &groceries, Date: now},
		{AccountID: 1, Type: model.TransactionExpense, Amount: 400, Date: now, Splits: []model.TransactionSplit{
			{CategoryID: &groceries, Amount: 100},
			{CategoryID: &sweets.ID, Amount: 300},
		}},
		{AccountID: 1, Type: model.TransactionExpense, Amount: 999, CategoryID: &groceries, Date: now.AddDate(0, -2, 0)},
	} {
		if err := txRepo.Create(ctx, tx); err != nil {
			t.Fatal(err)
		}
	}
	got, err := svc.Status(ctx, now.Year(), int(now.Month()))
	if err != nil || len(got) != 1 || got[0].Spent != 700 {
		t.Fatalf("got %+v, err %v", got, err)
	}
}
