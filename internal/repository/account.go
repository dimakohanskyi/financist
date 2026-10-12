package repository

import (
	"context"

	"gorm.io/gorm"
)

type Account struct {
	ID          int64 `gorm:"primaryKey"`
	Name        string
	AccountType string
	Currency    string
	IsArchived  bool
}

type AccountRepo struct{ db *gorm.DB }

func NewAccountRepo(db *gorm.DB) *AccountRepo { return &AccountRepo{db} }

func (r *AccountRepo) Create(ctx context.Context, a *Account) error {
	return gorm.G[Account](r.db).Create(ctx, a)
}

func (r *AccountRepo) GetByID(ctx context.Context, id int64) (Account, error) {
	return gorm.G[Account](r.db).Where("id = ?", id).First(ctx)
}

func (r *AccountRepo) List(ctx context.Context, includeArchived bool) ([]Account, error) {
	q := gorm.G[Account](r.db).Order("id")
	if !includeArchived {
		q = q.Where("is_archived = ?", false)
	}
	return q.Find(ctx)
}
