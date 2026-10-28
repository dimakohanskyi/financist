package service

import (
	"context"
	"testing"
	"time"

	"github.com/dimakohanskyi/financist/internal/model"
	"github.com/dimakohanskyi/financist/internal/repository"
)

func TestRecurringPost(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	accounts, txRepo := repository.NewAccountRepo(db), repository.NewTransactionRepo(db)
	if err := accounts.Create(ctx, &model.Account{Name: "Main", AccountType: "checking", Currency: "USD"}); err != nil {
		t.Fatal(err)
	}
	svc := NewRecurringService(repository.NewRecurringRepo(db), NewTransactionService(txRepo, accounts, repository.NewCategoryRepo(db)))

	rent := &model.Recurring{AccountID: 1, Type: model.TransactionExpense, Amount: 50000, Frequency: model.FrequencyMonthly, AnchorDate: time.Now()}
	if err := svc.Create(ctx, rent); err != nil {
		t.Fatal(err)
	}
	posted, err := svc.Post(ctx, rent.ID)
	if err != nil {
		t.Fatal(err)
	}
	today := model.Day(time.Now())
	if _, total, _ := txRepo.List(ctx, model.TransactionFilter{Limit: 10}); total != 1 || !posted.NextDue().Equal(model.AddMonths(today, 1)) {
		t.Fatalf("got %d transactions, next due %v", total, posted.NextDue())
	}
	jan31 := time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC)
	if got := model.AddMonths(jan31, 1); got.Day() != 28 {
		t.Fatalf("Jan 31 + 1 month = %v, want Feb 28", got)
	}
}
