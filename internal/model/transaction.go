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

type TransactionFilter struct {
	AccountID *int64 // matches both source and destination of transfers
	Limit     int
	Offset    int
}

// TransactionUpdate holds the fields to change; nil means "leave as is".
type TransactionUpdate struct {
	AccountID   *int64
	ToAccountID *int64
	Type        *string
	Amount      *int64
	Description *string
	Date        *time.Time
}
