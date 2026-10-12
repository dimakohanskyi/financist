package graph

import (
	"strconv"

	"github.com/dimakohanskyi/financist/internal/graph/model"
	"github.com/dimakohanskyi/financist/internal/repository"
)

func toAccount(a *repository.Account) *model.Account {
	return &model.Account{
		ID:          strconv.FormatInt(a.ID, 10),
		Name:        a.Name,
		AccountType: model.AccountType(a.AccountType),
		Currency:    model.Currency(a.Currency),
		IsArchived:  a.IsArchived,
	}
}
