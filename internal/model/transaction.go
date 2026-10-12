package model

import "time"

const (
	TransactionIncome   = "income"
	TransactionExpense  = "expense"
	TransactionTransfer = "transfer"
)

type Transaction struct {
	ID          int64
	AccountID   int64
	ToAccountID *int64 // only for transfers
	Type        string
	Amount      int64 // minor units (cents), always > 0
	Description string
	Date        time.Time
}
