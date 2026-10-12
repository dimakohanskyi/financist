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

// AccountUpdate holds the fields to change; nil means "leave as is".
type AccountUpdate struct {
	Name        *string
	AccountType *string
	Currency    *string
	IsArchived  *bool
}

func (r *AccountRepo) Update(ctx context.Context, id int64, u AccountUpdate) (Account, error) {
	fields := map[string]any{}
	if u.Name != nil {
		fields["name"] = *u.Name
	}
	if u.AccountType != nil {
		fields["account_type"] = *u.AccountType
	}
	if u.Currency != nil {
		fields["currency"] = *u.Currency
	}
	if u.IsArchived != nil {
		fields["is_archived"] = *u.IsArchived
	}
	if len(fields) > 0 {
		if err := r.db.WithContext(ctx).Model(&Account{}).Where("id = ?", id).Updates(fields).Error; err != nil {
			return Account{}, err
		}
	}
	return r.GetByID(ctx, id)
}

// Archive soft-deletes the account.
func (r *AccountRepo) Archive(ctx context.Context, id int64) error {
	archived := true
	_, err := r.Update(ctx, id, AccountUpdate{IsArchived: &archived})
	return err
}
