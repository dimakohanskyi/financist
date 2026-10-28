package service

import (
	"context"
	"errors"
	"testing"

	financistdb "github.com/dimakohanskyi/financist/internal/db"
	"github.com/dimakohanskyi/financist/internal/model"
	"github.com/dimakohanskyi/financist/internal/repository"
	"gorm.io/gorm"
)

// newTestDB returns a fresh in-memory database with all migrations applied.
func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := financistdb.Open("file:" + t.Name() + "?mode=memory&cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	return db
}

func TestCategoryNesting(t *testing.T) {
	svc := NewCategoryService(repository.NewCategoryRepo(newTestDB(t)))
	ctx := context.Background()

	groceries := int64(2) // seeded expense category
	sweets := &model.Category{Name: "Sweets", Kind: model.CategoryExpense, Color: "#1baf7a", ParentID: &groceries}
	if err := svc.Create(ctx, sweets); err != nil {
		t.Fatal(err)
	}
	nested := &model.Category{Name: "Chocolate", Kind: model.CategoryExpense, Color: "#1baf7a", ParentID: &sweets.ID}
	if err := svc.Create(ctx, nested); !errors.Is(err, ErrInvalid) {
		t.Fatalf("two-level nesting accepted, err: %v", err)
	}
}
