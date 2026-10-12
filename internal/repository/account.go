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
