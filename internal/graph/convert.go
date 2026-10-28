package graph

import (
	"strconv"
	"time"

	"github.com/dimakohanskyi/financist/internal/graph/model"
	domain "github.com/dimakohanskyi/financist/internal/model"
)

func toAccount(a *domain.Account) *model.Account {
	return &model.Account{
		ID:          strconv.FormatInt(a.ID, 10),
		Name:        a.Name,
		AccountType: model.AccountType(a.AccountType),
		Currency:    model.Currency(a.Currency),
		Color:       a.Color,
		IsArchived:  a.IsArchived,
	}
}

func parseID(id string) (int64, error) { return strconv.ParseInt(id, 10, 64) }

func toTransaction(t *domain.Transaction) *model.Transaction {
	out := &model.Transaction{
		ID:          strconv.FormatInt(t.ID, 10),
		AccountID:   strconv.FormatInt(t.AccountID, 10),
		ToAccountID: formatOptionalID(t.ToAccountID),
		CategoryID:  formatOptionalID(t.CategoryID),
		Type:        model.TransactionType(t.Type),
		Amount:      int(t.Amount),
		Description: t.Description,
		Merchant:    t.Merchant,
		Notes:       t.Notes,
		Date:        t.Date,
		Tags:        make([]*model.Tag, len(t.Tags)),
		Splits:      make([]*model.TransactionSplit, len(t.Splits)),
	}
	for i := range t.Tags {
		out.Tags[i] = toTag(&t.Tags[i])
	}
	for i, sp := range t.Splits {
		out.Splits[i] = &model.TransactionSplit{
			ID:         strconv.FormatInt(sp.ID, 10),
			CategoryID: formatOptionalID(sp.CategoryID),
			Amount:     int(sp.Amount),
			Note:       sp.Note,
		}
	}
	return out
}

// parseOptionalID parses a nullable ID; nil stays nil.
func parseOptionalID(id *string) (*int64, error) {
	if id == nil {
		return nil, nil
	}
	n, err := parseID(*id)
	return &n, err
}

func formatOptionalID(id *int64) *string {
	if id == nil {
		return nil
	}
	s := strconv.FormatInt(*id, 10)
	return &s
}

func toCategory(c *domain.Category) *model.Category {
	return &model.Category{
		ID:        strconv.FormatInt(c.ID, 10),
		Name:      c.Name,
		Kind:      model.CategoryKind(c.Kind),
		Icon:      c.Icon,
		Color:     c.Color,
		SortOrder: int32(c.SortOrder),
		IsDefault: c.IsDefault,
		ParentID:  formatOptionalID(c.ParentID),
	}
}

func toTag(t *domain.Tag) *model.Tag {
	return &model.Tag{ID: strconv.FormatInt(t.ID, 10), Name: t.Name}
}

func parseIDs(ids []string) ([]int64, error) {
	out := make([]int64, len(ids))
	for i, id := range ids {
		n, err := parseID(id)
		if err != nil {
			return nil, err
		}
		out[i] = n
	}
	return out, nil
}

func parseSplits(in []*model.TransactionSplitInput) ([]domain.TransactionSplit, error) {
	out := make([]domain.TransactionSplit, len(in))
	for i, sp := range in {
		id, err := parseID(sp.CategoryID)
		if err != nil {
			return nil, err
		}
		out[i] = domain.TransactionSplit{CategoryID: &id, Amount: int64(sp.Amount)}
		if sp.Note != nil {
			out[i].Note = *sp.Note
		}
	}
	return out, nil
}

func toNewTransaction(input *model.CreateTransactionInput) (*domain.Transaction, error) {
	accountID, err := parseID(input.AccountID)
	if err != nil {
		return nil, err
	}
	t := &domain.Transaction{
		AccountID: accountID,
		Type:      string(input.Type),
		Amount:    int64(input.Amount),
		Date:      input.Date,
	}
	if t.ToAccountID, err = parseOptionalID(input.ToAccountID); err != nil {
		return nil, err
	}
	if t.CategoryID, err = parseOptionalID(input.CategoryID); err != nil {
		return nil, err
	}
	if input.Description != nil {
		t.Description = *input.Description
	}
	if input.Merchant != nil {
		t.Merchant = *input.Merchant
	}
	if input.Notes != nil {
		t.Notes = *input.Notes
	}
	tagIDs, err := parseIDs(input.TagIds)
	if err != nil {
		return nil, err
	}
	for _, id := range tagIDs {
		t.Tags = append(t.Tags, domain.Tag{ID: id})
	}
	if t.Splits, err = parseSplits(input.Splits); err != nil {
		return nil, err
	}
	return t, nil
}

func toRecurring(r *domain.Recurring) *model.RecurringTransaction {
	today := domain.Day(time.Now())
	next := r.NextDue()
	return &model.RecurringTransaction{
		ID:             strconv.FormatInt(r.ID, 10),
		AccountID:      strconv.FormatInt(r.AccountID, 10),
		ToAccountID:    formatOptionalID(r.ToAccountID),
		CategoryID:     formatOptionalID(r.CategoryID),
		Type:           model.TransactionType(r.Type),
		Amount:         int(r.Amount),
		Description:    r.Description,
		Merchant:       r.Merchant,
		Notes:          r.Notes,
		Frequency:      model.RecurringFrequency(r.Frequency),
		AnchorDate:     r.AnchorDate,
		LastPostedDate: r.LastPostedDate,
		IsActive:       r.IsActive,
		NextDueDate:    next,
		IsDue:          r.IsActive && !next.After(today),
		DaysUntilDue:   int32(next.Sub(today).Hours() / 24),
	}
}

func toBudget(b *domain.Budget) *model.Budget {
	return &model.Budget{
		ID:           strconv.FormatInt(b.ID, 10),
		CategoryID:   strconv.FormatInt(b.CategoryID, 10),
		MonthlyLimit: int(b.MonthlyLimit),
	}
}

// percent is part of whole, 0 when whole is 0.
func percent(part, whole int64) float64 {
	if whole == 0 {
		return 0
	}
	return float64(part) / float64(whole) * 100
}

// yearMonth fills in the current month for missing arguments.
func yearMonth(year, month *int32) (int, int) {
	now := time.Now().UTC()
	y, m := now.Year(), int(now.Month())
	if year != nil {
		y = int(*year)
	}
	if month != nil {
		m = int(*month)
	}
	return y, m
}
