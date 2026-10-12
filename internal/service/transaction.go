package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/dimakohanskyi/financist/internal/model"
)

var ErrInvalid = errors.New("invalid transaction")

type AccountGetter interface {
	GetByID(ctx context.Context, id int64) (model.Account, error)
}

type TransactionStore interface {
	Create(ctx context.Context, t *model.Transaction) error
	GetByID(ctx context.Context, id int64) (model.Transaction, error)
	List(ctx context.Context, f model.TransactionFilter) ([]model.Transaction, int64, error)
	Update(ctx context.Context, t *model.Transaction) error
}

type TransactionService struct {
	txs      TransactionStore
	accounts AccountGetter
}

func NewTransactionService(txs TransactionStore, accounts AccountGetter) *TransactionService {
	return &TransactionService{txs, accounts}
}

func (s *TransactionService) Create(ctx context.Context, t *model.Transaction) error {
	if err := s.validate(ctx, t); err != nil {
		return err
	}
	return s.txs.Create(ctx, t)
}

func (s *TransactionService) validate(ctx context.Context, t *model.Transaction) error {
	if t.Amount <= 0 {
		return fmt.Errorf("%w: amount must be positive", ErrInvalid)
	}
	switch t.Type {
	case model.TransactionIncome, model.TransactionExpense:
		if t.ToAccountID != nil {
			return fmt.Errorf("%w: toAccountId is only for transfers", ErrInvalid)
		}
	case model.TransactionTransfer:
		if t.ToAccountID == nil || *t.ToAccountID == t.AccountID {
			return fmt.Errorf("%w: transfer needs a different toAccountId", ErrInvalid)
		}
	default:
		return fmt.Errorf("%w: unknown type %q", ErrInvalid, t.Type)
	}

	ids := []int64{t.AccountID}
	if t.ToAccountID != nil {
		ids = append(ids, *t.ToAccountID)
	}
	for _, id := range ids {
		acc, err := s.accounts.GetByID(ctx, id)
		if err != nil {
			return err
		}
		if acc.IsArchived {
			return fmt.Errorf("%w: account %d is archived", ErrInvalid, id)
		}
	}
	return nil
}

// Update applies the patch to the stored transaction and re-validates the result.
func (s *TransactionService) Update(ctx context.Context, id int64, u model.TransactionUpdate) (model.Transaction, error) {
	t, err := s.txs.GetByID(ctx, id)
	if err != nil {
		return t, err
	}
	if u.AccountID != nil {
		t.AccountID = *u.AccountID
	}
	if u.Type != nil {
		t.Type = *u.Type
	}
	if u.Amount != nil {
		t.Amount = *u.Amount
	}
	if u.Description != nil {
		t.Description = *u.Description
	}
	if u.Date != nil {
		t.Date = *u.Date
	}
	if u.ToAccountID != nil {
		t.ToAccountID = u.ToAccountID
	}
	if t.Type != model.TransactionTransfer {
		t.ToAccountID = nil // only transfers have a destination
	}
	if err := s.validate(ctx, &t); err != nil {
		return t, err
	}
	return t, s.txs.Update(ctx, &t)
}

func (s *TransactionService) Get(ctx context.Context, id int64) (model.Transaction, error) {
	return s.txs.GetByID(ctx, id)
}

const (
	defaultLimit = 50
	maxLimit     = 200
)

func (s *TransactionService) List(ctx context.Context, f model.TransactionFilter) ([]model.Transaction, int64, error) {
	if f.Limit <= 0 {
		f.Limit = defaultLimit
	}
	f.Limit = min(f.Limit, maxLimit)
	f.Offset = max(f.Offset, 0)
	return s.txs.List(ctx, f)
}
