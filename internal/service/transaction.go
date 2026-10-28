package service

import (
	"context"
	"errors"
	"fmt"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/dimakohanskyi/financist/internal/model"
)

var ErrInvalid = errors.New("invalid input")

type AccountGetter interface {
	GetByID(ctx context.Context, id int64) (model.Account, error)
}

type CategoryGetter interface {
	GetByID(ctx context.Context, id int64) (model.Category, error)
}

type TransactionStore interface {
	Create(ctx context.Context, t *model.Transaction) error
	GetByID(ctx context.Context, id int64) (model.Transaction, error)
	List(ctx context.Context, f model.TransactionFilter) ([]model.Transaction, int64, error)
	Update(ctx context.Context, t *model.Transaction) error
	Delete(ctx context.Context, id int64) error
	CreateMany(ctx context.Context, ts []*model.Transaction) error
	YearRange(ctx context.Context) (from, to int, err error)
}

type TransactionService struct {
	txs        TransactionStore
	accounts   AccountGetter
	categories CategoryGetter
}

func NewTransactionService(txs TransactionStore, accounts AccountGetter, categories CategoryGetter) *TransactionService {
	return &TransactionService{txs, accounts, categories}
}

func (s *TransactionService) Create(ctx context.Context, t *model.Transaction) error {
	t.Description = capitalizeFirst(t.Description)
	if err := s.validate(ctx, t); err != nil {
		return err
	}
	return s.txs.Create(ctx, t)
}

const maxBulk = 5000

// CreateMany validates every transaction before inserting any, so one bad
// row of a CSV import never leaves half a statement behind.
func (s *TransactionService) CreateMany(ctx context.Context, ts []*model.Transaction) error {
	if len(ts) == 0 || len(ts) > maxBulk {
		return fmt.Errorf("%w: send 1 to %d transactions", ErrInvalid, maxBulk)
	}
	for i, t := range ts {
		t.Description = capitalizeFirst(t.Description)
		if err := s.validate(ctx, t); err != nil {
			return fmt.Errorf("transaction %d: %w", i, err)
		}
	}
	return s.txs.CreateMany(ctx, ts)
}

// Years lists every year from the first transaction through the current one,
// gaps included, for a year picker.
func (s *TransactionService) Years(ctx context.Context) ([]int, error) {
	from, to, err := s.txs.YearRange(ctx)
	if err != nil {
		return nil, err
	}
	now := time.Now().Year()
	if from == 0 {
		return []int{now}, nil
	}
	var years []int
	for y := from; y <= max(to, now); y++ {
		years = append(years, y)
	}
	return years, nil
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
		if t.CategoryID != nil {
			return fmt.Errorf("%w: transfers have no category", ErrInvalid)
		}
	default:
		return fmt.Errorf("%w: unknown type %q", ErrInvalid, t.Type)
	}

	if t.CategoryID != nil {
		if _, err := s.categoryFor(ctx, *t.CategoryID, t.Type); err != nil {
			return err
		}
	}
	if err := s.validateSplits(ctx, t); err != nil {
		return err
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

// categoryFor fetches the category and checks it fits the transaction type:
// an income category on an expense would misclassify it in every report.
func (s *TransactionService) categoryFor(ctx context.Context, id int64, txType string) (model.Category, error) {
	c, err := s.categories.GetByID(ctx, id)
	if err != nil {
		return c, err
	}
	if c.Kind != txType {
		return c, fmt.Errorf("%w: category %q is for %s", ErrInvalid, c.Name, c.Kind)
	}
	return c, nil
}

// validateSplits checks a split divides the whole amount across one parent
// category and its direct subcategories, so it rolls up into one place.
func (s *TransactionService) validateSplits(ctx context.Context, t *model.Transaction) error {
	if len(t.Splits) == 0 {
		return nil
	}
	if t.Type == model.TransactionTransfer {
		return fmt.Errorf("%w: transfers can't be split", ErrInvalid)
	}
	if t.CategoryID != nil {
		return fmt.Errorf("%w: a split transaction has no categoryId of its own", ErrInvalid)
	}
	if len(t.Splits) < 2 {
		return fmt.Errorf("%w: a split needs at least 2 parts", ErrInvalid)
	}
	var sum int64
	var top *int64
	for _, sp := range t.Splits {
		if sp.Amount <= 0 {
			return fmt.Errorf("%w: split amounts must be positive", ErrInvalid)
		}
		sum += sp.Amount
		if sp.CategoryID == nil {
			continue // its category was deleted since; only the sum still matters
		}
		c, err := s.categoryFor(ctx, *sp.CategoryID, t.Type)
		if err != nil {
			return err
		}
		root := c.ID
		if c.ParentID != nil {
			root = *c.ParentID
		}
		if top != nil && *top != root {
			return fmt.Errorf("%w: split categories must share one parent", ErrInvalid)
		}
		top = &root
	}
	if sum != t.Amount {
		return fmt.Errorf("%w: splits add up to %d, not %d", ErrInvalid, sum, t.Amount)
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
		t.Description = capitalizeFirst(*u.Description)
	}
	if u.Merchant != nil {
		t.Merchant = *u.Merchant
	}
	if u.Notes != nil {
		t.Notes = *u.Notes
	}
	if u.TagIDs != nil {
		t.Tags = make([]model.Tag, len(*u.TagIDs))
		for i, id := range *u.TagIDs {
			t.Tags[i] = model.Tag{ID: id}
		}
	}
	if u.Date != nil {
		t.Date = *u.Date
	}
	if u.ToAccountID != nil {
		t.ToAccountID = u.ToAccountID
	}
	// a category and splits are alternatives: setting one drops the other
	if u.CategoryID != nil {
		t.CategoryID = u.CategoryID
		if u.Splits == nil {
			t.Splits = nil
		}
	}
	if u.Splits != nil {
		t.Splits = *u.Splits
		if len(t.Splits) > 0 && u.CategoryID == nil {
			t.CategoryID = nil
		}
	}
	if u.ClearCategory {
		t.CategoryID = nil
	}
	if t.Type != model.TransactionTransfer {
		t.ToAccountID = nil // only transfers have a destination
	} else {
		t.CategoryID, t.Splits = nil, nil // and only non-transfers have categories
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

func (s *TransactionService) Delete(ctx context.Context, id int64) error {
	return s.txs.Delete(ctx, id)
}

// capitalizeFirst uppercases only the first letter: "iPhone 15" stays as is,
// "продукти" becomes "Продукти".
func capitalizeFirst(s string) string {
	if s == "" {
		return s
	}
	r, size := utf8.DecodeRuneInString(s)
	return string(unicode.ToUpper(r)) + s[size:]
}
