package repository

import (
	"context"

	"github.com/dimakohanskyi/financist/internal/model"
	"gorm.io/gorm"
)

// accountRow is the GORM mapping of the accounts table.
type accountRow struct {
	ID          int64 `gorm:"primaryKey"`
	Name        string
	AccountType string
	Currency    string
	IsArchived  bool
}

func (accountRow) TableName() string { return "accounts" }

func toRow(a *model.Account) accountRow { return accountRow(*a) }

func (r accountRow) toModel() model.Account { return model.Account(r) }

type AccountRepo struct{ db *gorm.DB }

func NewAccountRepo(db *gorm.DB) *AccountRepo { return &AccountRepo{db} }

func (r *AccountRepo) Create(ctx context.Context, a *model.Account) error {
	row := toRow(a)
	if err := gorm.G[accountRow](r.db).Create(ctx, &row); err != nil {
		return err
	}
	a.ID = row.ID
	return nil
}

func (r *AccountRepo) GetByID(ctx context.Context, id int64) (model.Account, error) {
	row, err := gorm.G[accountRow](r.db).Where("id = ?", id).First(ctx)
	return row.toModel(), err
}

func (r *AccountRepo) List(ctx context.Context, includeArchived bool) ([]model.Account, error) {
	q := gorm.G[accountRow](r.db).Order("id")
	if !includeArchived {
		q = q.Where("is_archived = ?", false)
	}
	rows, err := q.Find(ctx)
	out := make([]model.Account, len(rows))
	for i, row := range rows {
		out[i] = row.toModel()
	}
	return out, err
}

func (r *AccountRepo) Update(ctx context.Context, id int64, u model.AccountUpdate) (model.Account, error) {
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
		if err := r.db.WithContext(ctx).Model(&accountRow{}).Where("id = ?", id).Updates(fields).Error; err != nil {
			return model.Account{}, err
		}
	}
	return r.GetByID(ctx, id)
}

// Archive soft-deletes the account.
func (r *AccountRepo) Archive(ctx context.Context, id int64) error {
	archived := true
	_, err := r.Update(ctx, id, model.AccountUpdate{IsArchived: &archived})
	return err
}
