package repository

import (
	"context"
	"strings"

	"github.com/dimakohanskyi/financist/internal/model"
	"gorm.io/gorm"
)

// tagRow is the GORM mapping of the tags table.
type tagRow struct {
	ID   int64 `gorm:"primaryKey"`
	Name string
}

func (tagRow) TableName() string { return "tags" }

// transactionTagRow is the GORM mapping of the transaction_tags join table.
type transactionTagRow struct {
	TransactionID int64 `gorm:"primaryKey"`
	TagID         int64 `gorm:"primaryKey"`
}

func (transactionTagRow) TableName() string { return "transaction_tags" }

type TagRepo struct{ db *gorm.DB }

func NewTagRepo(db *gorm.DB) *TagRepo { return &TagRepo{db} }

// Create returns the existing tag when one has the same name ignoring case,
// so "Georgia" and "georgia" typed on two transactions end up as one tag.
func (r *TagRepo) Create(ctx context.Context, name string) (model.Tag, error) {
	row := tagRow{Name: strings.TrimSpace(name)}
	err := r.db.WithContext(ctx).Where("name = ?", row.Name).FirstOrCreate(&row).Error
	return model.Tag(row), err
}

func (r *TagRepo) List(ctx context.Context) ([]model.Tag, error) {
	rows, err := gorm.G[tagRow](r.db).Order("name").Find(ctx)
	out := make([]model.Tag, len(rows))
	for i, row := range rows {
		out[i] = model.Tag(row)
	}
	return out, err
}

func (r *TagRepo) Delete(ctx context.Context, id int64) error {
	res := r.db.WithContext(ctx).Delete(&tagRow{}, id)
	if res.Error == nil && res.RowsAffected == 0 {
		return model.ErrNotFound
	}
	return res.Error
}
