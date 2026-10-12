package repository

import (
	"context"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func newTestRepo(t *testing.T) *AccountRepo {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&Account{}); err != nil {
		t.Fatal(err)
	}
	return NewAccountRepo(db)
}

func TestCreate(t *testing.T) {
	a := &Account{Name: "Main", AccountType: "checking", Currency: "USD"}
	if err := newTestRepo(t).Create(context.Background(), a); err != nil {
		t.Fatal(err)
	}
	if a.ID == 0 {
		t.Fatal("expected id to be set")
	}
}

func TestGetByID(t *testing.T) {
	r := newTestRepo(t)
	a := &Account{Name: "Main", AccountType: "checking", Currency: "USD"}
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
	for _, a := range []*Account{
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
	a := &Account{Name: "Main", AccountType: "checking", Currency: "USD"}
	if err := r.Create(ctx, a); err != nil {
		t.Fatal(err)
	}
	name, archived := "Renamed", true
	got, err := r.Update(ctx, a.ID, AccountUpdate{Name: &name, IsArchived: &archived})
	if err != nil || got.Name != "Renamed" || !got.IsArchived || got.Currency != "USD" {
		t.Fatalf("got %+v, err %v", got, err)
	}
}
