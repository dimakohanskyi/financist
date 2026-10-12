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
