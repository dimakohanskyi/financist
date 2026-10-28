package repository

import (
	"context"
	"testing"
	"time"

	"github.com/dimakohanskyi/financist/internal/model"
)

func TestTagDedupAndLink(t *testing.T) {
	db := withAccounts(t, newTestDB(t))
	ctx := context.Background()
	tags := NewTagRepo(db)
	a, err1 := tags.Create(ctx, "Georgia")
	b, err2 := tags.Create(ctx, " georgia ")
	if err1 != nil || err2 != nil || a.ID != b.ID {
		t.Fatalf("got %+v and %+v, errs %v %v", a, b, err1, err2)
	}

	txs := NewTransactionRepo(db)
	tx := &model.Transaction{AccountID: 1, Type: model.TransactionExpense, Amount: 100, Date: time.Now(), Tags: []model.Tag{{ID: a.ID}}}
	if err := txs.Create(ctx, tx); err != nil {
		t.Fatal(err)
	}
	got, err := txs.GetByID(ctx, tx.ID)
	if err != nil || len(got.Tags) != 1 || got.Tags[0].Name != "Georgia" {
		t.Fatalf("got %+v, err %v", got, err)
	}
}
