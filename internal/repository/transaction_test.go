package repository

import (
	"context"
	"testing"
	"time"

	"github.com/dimakohanskyi/financist/internal/model"
)

func TestTransactionCreate(t *testing.T) {
	db := newTestDB(t)
	tx := &model.Transaction{AccountID: 1, Type: model.TransactionExpense, Amount: 500, Date: time.Now()}
	if err := NewTransactionRepo(db).Create(context.Background(), tx); err != nil {
		t.Fatal(err)
	}
	if tx.ID == 0 {
		t.Fatal("expected id to be set")
	}
}

func TestTransactionGetByID(t *testing.T) {
	r := NewTransactionRepo(newTestDB(t))
	ctx := context.Background()
	tx := &model.Transaction{AccountID: 1, Type: model.TransactionIncome, Amount: 700, Date: time.Now()}
	if err := r.Create(ctx, tx); err != nil {
		t.Fatal(err)
	}
	got, err := r.GetByID(ctx, tx.ID)
	if err != nil || got.Amount != 700 {
		t.Fatalf("got %+v, err %v", got, err)
	}
}

func TestTransactionList(t *testing.T) {
	r := NewTransactionRepo(newTestDB(t))
	ctx := context.Background()
	for _, acc := range []int64{1, 1, 2} {
		tx := &model.Transaction{AccountID: acc, Type: model.TransactionExpense, Amount: 100, Date: time.Now()}
		if err := r.Create(ctx, tx); err != nil {
			t.Fatal(err)
		}
	}
	one := int64(1)
	got, total, err := r.List(ctx, model.TransactionFilter{AccountID: &one, Limit: 1})
	if err != nil || len(got) != 1 || total != 2 {
		t.Fatalf("got %d items, total %d, err %v", len(got), total, err)
	}
}
