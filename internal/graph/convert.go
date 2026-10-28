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
	return &model.Transaction{
		ID:          strconv.FormatInt(t.ID, 10),
		AccountID:   strconv.FormatInt(t.AccountID, 10),
		ToAccountID: formatOptionalID(t.ToAccountID),
		CategoryID:  formatOptionalID(t.CategoryID),
		Type:        model.TransactionType(t.Type),
		Amount:      int(t.Amount),
		Description: t.Description,
		Date:        t.Date,
	}
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
