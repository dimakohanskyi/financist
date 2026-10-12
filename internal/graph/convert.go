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
		IsArchived:  a.IsArchived,
	}
}

func parseID(id string) (int64, error) { return strconv.ParseInt(id, 10, 64) }

func toTransaction(t *domain.Transaction) *model.Transaction {
	out := &model.Transaction{
		ID:          strconv.FormatInt(t.ID, 10),
		AccountID:   strconv.FormatInt(t.AccountID, 10),
		Type:        model.TransactionType(t.Type),
		Amount:      int(t.Amount),
		Description: t.Description,
		Date:        t.Date,
	}
	if t.ToAccountID != nil {
		to := strconv.FormatInt(*t.ToAccountID, 10)
		out.ToAccountID = &to
	}
	return out
}
