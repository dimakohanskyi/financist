package service

import (
	"context"
	"testing"
	"time"

	"github.com/dimakohanskyi/financist/internal/model"
	"github.com/dimakohanskyi/financist/internal/repository"
)

func TestReports(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	accounts, categories, txRepo, assets := repository.NewAccountRepo(db), repository.NewCategoryRepo(db), repository.NewTransactionRepo(db), repository.NewAssetRepo(db)
	for _, a := range []*model.Account{
		{Name: "Main", AccountType: "checking", Currency: "USD"},
		{Name: "Card", AccountType: "credit_card", Currency: "USD"}, // not cash
	} {
		if err := accounts.Create(ctx, a); err != nil {
			t.Fatal(err)
		}
	}
	groceries, salary := int64(2), int64(9) // seeded categories
	sweets := &model.Category{Name: "Sweets", Kind: model.CategoryExpense, Color: "#1baf7a", ParentID: &groceries}
	if err := categories.Create(ctx, sweets); err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	card := int64(2)
	for _, tx := range []*model.Transaction{
		{AccountID: 1, Type: model.TransactionIncome, Amount: 5000, CategoryID: &salary, Date: now},
		{AccountID: 1, Type: model.TransactionExpense, Amount: 300, CategoryID: &groceries, Date: now},
		{AccountID: 1, Type: model.TransactionExpense, Amount: 200, CategoryID: &sweets.ID, Date: now},
		{AccountID: 1, ToAccountID: &card, Type: model.TransactionTransfer, Amount: 1000, Date: now},
	} {
		if err := txRepo.Create(ctx, tx); err != nil {
			t.Fatal(err)
		}
	}
	if err := assets.Create(ctx, &model.Asset{Name: "Car", AssetClass: "vehicles", Currency: "USD", CapitalRole: model.CapitalDrain, RiskLevel: model.RiskMedium}, 2000, now); err != nil {
		t.Fatal(err)
	}
	svc := NewReportService(txRepo, accounts, categories, assets, repository.NewSettingsRepo(db), NewBudgetService(repository.NewBudgetRepo(db), categories, txRepo))

	d, err := svc.Dashboard(ctx, now.Year(), int(now.Month()), nil)
	if err != nil || d.Spent != 500 || d.TransferredOut != 1000 || len(d.SpendingByCategory) != 1 || len(d.SpendingByCategory[0].Children) != 2 {
		t.Fatalf("dashboard %+v, err %v", d, err)
	}
	// the card only received a transfer: no income, no spending
	d, err = svc.Dashboard(ctx, now.Year(), int(now.Month()), &card)
	if err != nil || d.Income != 0 || d.Spent != 0 || d.TransferredOut != 0 || len(d.SpendingByCategory) != 0 {
		t.Fatalf("card dashboard %+v, err %v", d, err)
	}
	// cash: 5000 - 500 - 1000 moved to the card = 3500, plus the car
	nw, err := svc.NetWorth(ctx, "last_30_days")
	if err != nil || nw.Current != 5500 || nw.Breakdown[0].Amount != 3500 {
		t.Fatalf("net worth %+v, err %v", nw, err)
	}
	if _, err := svc.Alerts(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Advice(ctx); err != nil {
		t.Fatal(err)
	}
}
