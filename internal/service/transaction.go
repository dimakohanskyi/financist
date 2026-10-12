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
}

type TransactionService struct {
	txs      TransactionStore
	accounts AccountGetter
}

func NewTransactionService(txs TransactionStore, accounts AccountGetter) *TransactionService {
	return &TransactionService{txs, accounts}
}

func (s *TransactionService) Create(ctx context.Context, t *model.Transaction) error {
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
	return s.txs.Create(ctx, t)
}
