package graph

import (
	"strconv"

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
