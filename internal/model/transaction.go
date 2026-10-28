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
	CategoryID  *int64 // kind matches Type; never on transfers
	Type        string
	Amount      int64 // minor units (cents), always > 0
	Description string
	Merchant    string
	Notes       string
	Date        time.Time
	Tags        []Tag // only IDs matter when writing
}

type TransactionFilter struct {
	AccountID *int64 // matches both source and destination of transfers
	Limit     int
	Offset    int
}

// TransactionUpdate holds the fields to change; nil means "leave as is".
type TransactionUpdate struct {
	AccountID     *int64
	ToAccountID   *int64
	CategoryID    *int64
	ClearCategory bool
	Type          *string
	Amount        *int64
	Description   *string
	Merchant      *string
	Notes         *string
	Date          *time.Time
	TagIDs        *[]int64 // replaces the whole set; empty clears it
}
