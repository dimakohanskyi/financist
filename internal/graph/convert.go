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
