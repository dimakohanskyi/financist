package repository

import (
	"context"
	"errors"

	"github.com/dimakohanskyi/financist/internal/model"
	"gorm.io/gorm"
)

// categoryRow is the GORM mapping of the categories table.
type categoryRow struct {
	ID        int64 `gorm:"primaryKey"`
	Name      string
	Kind      string
	Icon      *string
	Color     string
	SortOrder int
	IsDefault bool
	ParentID  *int64
}

func (categoryRow) TableName() string { return "categories" }

type CategoryRepo struct{ db *gorm.DB }

func NewCategoryRepo(db *gorm.DB) *CategoryRepo { return &CategoryRepo{db} }

func (r *CategoryRepo) Create(ctx context.Context, c *model.Category) error {
	row := categoryRow(*c)
	if err := gorm.G[categoryRow](r.db).Create(ctx, &row); err != nil {
		return err
	}
	c.ID = row.ID
	return nil
}

func (r *CategoryRepo) GetByID(ctx context.Context, id int64) (model.Category, error) {
	row, err := gorm.G[categoryRow](r.db).Where("id = ?", id).First(ctx)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = model.ErrNotFound
	}
	return model.Category(row), err
}

// List returns categories in their display order, optionally of one kind.
func (r *CategoryRepo) List(ctx context.Context, kind *string) ([]model.Category, error) {
	q := gorm.G[categoryRow](r.db).Order("sort_order, id")
	if kind != nil {
		q = q.Where("kind = ?", *kind)
	}
	rows, err := q.Find(ctx)
	out := make([]model.Category, len(rows))
	for i, row := range rows {
		out[i] = model.Category(row)
	}
	return out, err
}

func (r *CategoryRepo) Update(ctx context.Context, id int64, u model.CategoryUpdate) (model.Category, error) {
	fields := map[string]any{}
	if u.Name != nil {
		fields["name"] = *u.Name
	}
	if u.Icon != nil {
		fields["icon"] = nullIfEmpty(*u.Icon)
	}
	if u.Color != nil {
		fields["color"] = *u.Color
	}
	if u.SortOrder != nil {
		fields["sort_order"] = *u.SortOrder
	}
	if u.ParentID != nil {
		fields["parent_id"] = *u.ParentID
	}
	if u.ClearParent {
		fields["parent_id"] = nil
	}
	if len(fields) > 0 {
		if err := r.db.WithContext(ctx).Model(&categoryRow{}).Where("id = ?", id).Updates(fields).Error; err != nil {
			return model.Category{}, err
		}
	}
	return r.GetByID(ctx, id)
}

func (r *CategoryRepo) Delete(ctx context.Context, id int64) error {
	res := r.db.WithContext(ctx).Delete(&categoryRow{}, id)
	if res.Error == nil && res.RowsAffected == 0 {
		return model.ErrNotFound
	}
	return res.Error
}

func (r *CategoryRepo) HasChildren(ctx context.Context, id int64) (bool, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&categoryRow{}).Where("parent_id = ?", id).Limit(1).Count(&n).Error
	return n > 0, err
}

// InUse reports whether any transaction or split line points at the category.
func (r *CategoryRepo) InUse(ctx context.Context, id int64) (bool, error) {
	var used bool
	err := r.db.WithContext(ctx).Raw(`SELECT EXISTS(SELECT 1 FROM transactions WHERE category_id = ?)`, id).Scan(&used).Error
	return used, err
}
