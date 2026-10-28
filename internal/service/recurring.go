package service

import (
	"context"
	"time"

	"github.com/dimakohanskyi/financist/internal/model"
)

type RecurringStore interface {
	Create(ctx context.Context, r *model.Recurring) error
	GetByID(ctx context.Context, id int64) (model.Recurring, error)
	List(ctx context.Context) ([]model.Recurring, error)
	Update(ctx context.Context, r *model.Recurring) error
	Delete(ctx context.Context, id int64) error
	Post(ctx context.Context, id int64, t *model.Transaction) error
}

// RecurringService checks templates with the same rules as the transactions they post.
type RecurringService struct {
	store RecurringStore
	txs   *TransactionService
}

func NewRecurringService(store RecurringStore, txs *TransactionService) *RecurringService {
	return &RecurringService{store, txs}
}

func (s *RecurringService) Create(ctx context.Context, r *model.Recurring) error {
	r.Description = capitalizeFirst(r.Description)
	r.AnchorDate = model.Day(r.AnchorDate)
	r.IsActive = true
	if err := s.validate(ctx, r); err != nil {
		return err
	}
	return s.store.Create(ctx, r)
}

func (s *RecurringService) Get(ctx context.Context, id int64) (model.Recurring, error) {
	return s.store.GetByID(ctx, id)
}

func (s *RecurringService) List(ctx context.Context) ([]model.Recurring, error) {
	return s.store.List(ctx)
}

// Update applies the patch to the stored template and re-validates the result.
func (s *RecurringService) Update(ctx context.Context, id int64, u model.RecurringUpdate) (model.Recurring, error) {
	r, err := s.store.GetByID(ctx, id)
	if err != nil {
		return r, err
	}
	if u.AccountID != nil {
		r.AccountID = *u.AccountID
	}
	if u.ToAccountID != nil {
		r.ToAccountID = u.ToAccountID
	}
	if u.CategoryID != nil {
		r.CategoryID = u.CategoryID
	}
	if u.ClearCategory {
		r.CategoryID = nil
	}
	if u.Type != nil {
		r.Type = *u.Type
	}
	if u.Amount != nil {
		r.Amount = *u.Amount
	}
	if u.Description != nil {
		r.Description = capitalizeFirst(*u.Description)
	}
	if u.Merchant != nil {
		r.Merchant = *u.Merchant
	}
	if u.Notes != nil {
		r.Notes = *u.Notes
	}
	if u.Frequency != nil {
		r.Frequency = *u.Frequency
	}
	if u.AnchorDate != nil {
		r.AnchorDate = model.Day(*u.AnchorDate)
	}
	if u.IsActive != nil {
		r.IsActive = *u.IsActive
	}
	if r.Type != model.TransactionTransfer {
		r.ToAccountID = nil
	} else {
		r.CategoryID = nil
	}
	if err := s.validate(ctx, &r); err != nil {
		return r, err
	}
	return r, s.store.Update(ctx, &r)
}

func (s *RecurringService) Delete(ctx context.Context, id int64) error {
	return s.store.Delete(ctx, id)
}

// Post creates a transaction dated today from the template; this is the
// only thing that moves its schedule forward.
func (s *RecurringService) Post(ctx context.Context, id int64) (model.Recurring, error) {
	r, err := s.store.GetByID(ctx, id)
	if err != nil {
		return r, err
	}
	today := model.Day(time.Now())
	t := r.Transaction(today)
	if err := s.txs.validate(ctx, &t); err != nil {
		return r, err
	}
	if err := s.store.Post(ctx, id, &t); err != nil {
		return r, err
	}
	r.LastPostedDate = &today
	return r, nil
}

func (s *RecurringService) validate(ctx context.Context, r *model.Recurring) error {
	t := r.Transaction(r.AnchorDate)
	return s.txs.validate(ctx, &t)
}
