package service

import (
	"context"
	"fmt"

	"github.com/dimakohanskyi/financist/internal/model"
)

type CategoryStore interface {
	Create(ctx context.Context, c *model.Category) error
	GetByID(ctx context.Context, id int64) (model.Category, error)
	List(ctx context.Context, kind *string) ([]model.Category, error)
	Update(ctx context.Context, id int64, u model.CategoryUpdate) (model.Category, error)
	Delete(ctx context.Context, id int64) error
	HasChildren(ctx context.Context, id int64) (bool, error)
	InUse(ctx context.Context, id int64) (bool, error)
}

type CategoryService struct{ store CategoryStore }

func NewCategoryService(store CategoryStore) *CategoryService { return &CategoryService{store} }

func (s *CategoryService) Create(ctx context.Context, c *model.Category) error {
	if c.ParentID != nil {
		if err := s.validateParent(ctx, *c.ParentID, c.Kind, 0); err != nil {
			return err
		}
	}
	c.IsDefault = false
	return s.store.Create(ctx, c)
}

func (s *CategoryService) Get(ctx context.Context, id int64) (model.Category, error) {
	return s.store.GetByID(ctx, id)
}

func (s *CategoryService) List(ctx context.Context, kind *string) ([]model.Category, error) {
	return s.store.List(ctx, kind)
}

func (s *CategoryService) Update(ctx context.Context, id int64, u model.CategoryUpdate) (model.Category, error) {
	if u.ParentID != nil && !u.ClearParent {
		c, err := s.store.GetByID(ctx, id)
		if err != nil {
			return c, err
		}
		if err := s.validateParent(ctx, *u.ParentID, c.Kind, id); err != nil {
			return c, err
		}
		hasChildren, err := s.store.HasChildren(ctx, id)
		if err != nil {
			return c, err
		}
		if hasChildren {
			return c, fmt.Errorf("%w: a category with subcategories can't become a subcategory", ErrInvalid)
		}
	}
	return s.store.Update(ctx, id, u)
}

// Delete removes the category; its subcategories become top-level and its
// transactions uncategorized. Default categories can only go once unused,
// so a year of history doesn't silently lose its category.
func (s *CategoryService) Delete(ctx context.Context, id int64) error {
	c, err := s.store.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if c.IsDefault {
		used, err := s.store.InUse(ctx, id)
		if err != nil {
			return err
		}
		if used {
			return fmt.Errorf("%w: default categories can only be deleted once they have no transactions", ErrInvalid)
		}
	}
	return s.store.Delete(ctx, id)
}

// validateParent keeps subcategories one level deep and of the parent's kind.
func (s *CategoryService) validateParent(ctx context.Context, parentID int64, kind string, id int64) error {
	if parentID == id {
		return fmt.Errorf("%w: a category can't be its own parent", ErrInvalid)
	}
	parent, err := s.store.GetByID(ctx, parentID)
	if err != nil {
		return err
	}
	if parent.ParentID != nil {
		return fmt.Errorf("%w: subcategories can only be one level deep", ErrInvalid)
	}
	if parent.Kind != kind {
		return fmt.Errorf("%w: a subcategory must have its parent's kind", ErrInvalid)
	}
	return nil
}
