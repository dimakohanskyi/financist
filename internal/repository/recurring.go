package repository

import (
	"context"
	"errors"
	"time"

	"github.com/dimakohanskyi/financist/internal/model"
	"gorm.io/gorm"
)

// recurringRow is the GORM mapping of the recurring_transactions table.
type recurringRow struct {
	ID             int64 `gorm:"primaryKey"`
	AccountID      int64
	ToAccountID    *int64
	CategoryID     *int64
	Type           string
	Amount         int64
	Description    string
	Merchant       string
	Notes          string
	Frequency      string
	AnchorDate     time.Time
	LastPostedDate *time.Time
	IsActive       bool
}

func (recurringRow) TableName() string { return "recurring_transactions" }

type RecurringRepo struct{ db *gorm.DB }

func NewRecurringRepo(db *gorm.DB) *RecurringRepo { return &RecurringRepo{db} }

func (r *RecurringRepo) Create(ctx context.Context, rec *model.Recurring) error {
	row := recurringRow(*rec)
	if err := gorm.G[recurringRow](r.db).Create(ctx, &row); err != nil {
		return err
	}
	rec.ID = row.ID
	return nil
}

func (r *RecurringRepo) GetByID(ctx context.Context, id int64) (model.Recurring, error) {
	row, err := gorm.G[recurringRow](r.db).Where("id = ?", id).First(ctx)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = model.ErrNotFound
	}
	return model.Recurring(row), err
}

func (r *RecurringRepo) List(ctx context.Context) ([]model.Recurring, error) {
	rows, err := gorm.G[recurringRow](r.db).Order("id").Find(ctx)
	out := make([]model.Recurring, len(rows))
	for i, row := range rows {
		out[i] = model.Recurring(row)
	}
	return out, err
}

// Update overwrites every column of an existing template.
func (r *RecurringRepo) Update(ctx context.Context, rec *model.Recurring) error {
	row := recurringRow(*rec)
	return r.db.WithContext(ctx).Model(&recurringRow{}).Where("id = ?", rec.ID).Select("*").Omit("id").Updates(&row).Error
}

func (r *RecurringRepo) Delete(ctx context.Context, id int64) error {
	res := r.db.WithContext(ctx).Delete(&recurringRow{}, id)
	if res.Error == nil && res.RowsAffected == 0 {
		return model.ErrNotFound
	}
	return res.Error
}

// Post saves t and marks the template posted on t's date, both or neither.
func (r *RecurringRepo) Post(ctx context.Context, id int64, t *model.Transaction) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := create(tx, t); err != nil {
			return err
		}
		return tx.Model(&recurringRow{}).Where("id = ?", id).Update("last_posted_date", t.Date).Error
	})
}
