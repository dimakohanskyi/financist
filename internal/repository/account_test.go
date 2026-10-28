package repository

import (
	"context"
	"testing"
	"time"

	financistdb "github.com/dimakohanskyi/financist/internal/db"
	"github.com/dimakohanskyi/financist/internal/model"
	"gorm.io/gorm"
)

// newTestDB returns a fresh in-memory database with all migrations applied.
func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	return openTestDB(t, t.Name())
}

func openTestDB(t *testing.T, name string) *gorm.DB {
	t.Helper()
	db, err := financistdb.Open("file:" + name + "?mode=memory&cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	return db
}

// withAccounts creates accounts 1 and 2 for tests that need foreign keys to hold.
func withAccounts(t *testing.T, db *gorm.DB) *gorm.DB {
	t.Helper()
	for _, name := range []string{"Main", "Savings"} {
		if err := NewAccountRepo(db).Create(context.Background(), &model.Account{Name: name, AccountType: "checking", Currency: "USD"}); err != nil {
			t.Fatal(err)
		}
	}
	return db
}

func newTestRepo(t *testing.T) *AccountRepo { return NewAccountRepo(newTestDB(t)) }

func TestCreate(t *testing.T) {
	a := &model.Account{Name: "Main", AccountType: "checking", Currency: "USD"}
	if err := newTestRepo(t).Create(context.Background(), a); err != nil {
		t.Fatal(err)
	}
	if a.ID == 0 {
		t.Fatal("expected id to be set")
	}
}

func TestGetByID(t *testing.T) {
	r := newTestRepo(t)
	a := &model.Account{Name: "Main", AccountType: "checking", Currency: "USD"}
	if err := r.Create(context.Background(), a); err != nil {
		t.Fatal(err)
	}
	got, err := r.GetByID(context.Background(), a.ID)
	if err != nil || got.Name != "Main" {
		t.Fatalf("got %+v, err %v", got, err)
	}
}

func TestList(t *testing.T) {
	r := newTestRepo(t)
	ctx := context.Background()
	for _, a := range []*model.Account{
		{Name: "Active", AccountType: "cash", Currency: "USD"},
		{Name: "Old", AccountType: "cash", Currency: "USD", IsArchived: true},
	} {
		if err := r.Create(ctx, a); err != nil {
			t.Fatal(err)
		}
	}
	got, err := r.List(ctx, false)
	if err != nil || len(got) != 1 || got[0].Name != "Active" {
		t.Fatalf("got %+v, err %v", got, err)
	}
}

func TestUpdate(t *testing.T) {
	r := newTestRepo(t)
	ctx := context.Background()
	a := &model.Account{Name: "Main", AccountType: "checking", Currency: "USD"}
	if err := r.Create(ctx, a); err != nil {
		t.Fatal(err)
	}
	name, archived := "Renamed", true
	got, err := r.Update(ctx, a.ID, model.AccountUpdate{Name: &name, IsArchived: &archived})
	if err != nil || got.Name != "Renamed" || !got.IsArchived || got.Currency != "USD" {
		t.Fatalf("got %+v, err %v", got, err)
	}
}

func TestArchive(t *testing.T) {
	r := newTestRepo(t)
	ctx := context.Background()
	a := &model.Account{Name: "Main", AccountType: "checking", Currency: "USD"}
	if err := r.Create(ctx, a); err != nil {
		t.Fatal(err)
	}
	if err := r.Archive(ctx, a.ID); err != nil {
		t.Fatal(err)
	}
	got, err := r.List(ctx, false)
	if err != nil || len(got) != 0 {
		t.Fatalf("got %+v, err %v", got, err)
	}
}

func TestBalance(t *testing.T) {
	db := withAccounts(t, newTestDB(t))
	ctx := context.Background()
	to := int64(2)
	for _, tx := range []*model.Transaction{
		{AccountID: 1, Type: model.TransactionIncome, Amount: 1000, Date: time.Now()},
		{AccountID: 1, Type: model.TransactionExpense, Amount: 300, Date: time.Now()},
		{AccountID: 1, ToAccountID: &to, Type: model.TransactionTransfer, Amount: 200, Date: time.Now()},
	} {
		if err := NewTransactionRepo(db).Create(ctx, tx); err != nil {
			t.Fatal(err)
		}
	}
	r := NewAccountRepo(db)
	main, err1 := r.Balance(ctx, 1)
	savings, err2 := r.Balance(ctx, 2)
	if err1 != nil || err2 != nil || main != 500 || savings != 200 {
		t.Fatalf("got %d and %d, errs %v %v", main, savings, err1, err2)
	}
}
