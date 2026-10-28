package repository

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/dimakohanskyi/financist/internal/model"
)

func TestBackupRoundTrip(t *testing.T) {
	src := withAccounts(t, newTestDB(t))
	ctx := context.Background()
	tag, _ := NewTagRepo(src).Create(ctx, "trip")
	groceries := int64(2)
	tx := &model.Transaction{AccountID: 1, Type: model.TransactionExpense, Amount: 100, CategoryID: &groceries, Date: time.Now(), Tags: []model.Tag{tag}}
	if err := NewTransactionRepo(src).Create(ctx, tx); err != nil {
		t.Fatal(err)
	}
	b, err := NewBackupRepo(src).Export(ctx)
	if err != nil {
		t.Fatal(err)
	}
	data, _ := json.Marshal(b)

	dst := openTestDB(t, t.Name()+"_dst")
	var restored Backup
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatal(err)
	}
	if err := NewBackupRepo(dst).Import(ctx, &restored); err != nil {
		t.Fatal(err)
	}
	got, err := NewTransactionRepo(dst).GetByID(ctx, tx.ID)
	if err != nil || len(got.Tags) != 1 || *got.CategoryID != groceries {
		t.Fatalf("got %+v, err %v", got, err)
	}

	restored.Transactions[0].AccountID = 99 // unknown account: nothing may change
	if err := NewBackupRepo(dst).Import(ctx, &restored); err == nil {
		t.Fatal("backup with a broken reference accepted")
	}
	if _, err := NewTransactionRepo(dst).GetByID(ctx, tx.ID); err != nil {
		t.Fatalf("failed import changed data: %v", err)
	}
}
