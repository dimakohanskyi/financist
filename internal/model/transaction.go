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
	Splits      []TransactionSplit
}

// TransactionSplit is one category's share of a split transaction.
type TransactionSplit struct {
	ID         int64
	CategoryID *int64 // nil once its category is deleted
	Amount     int64  // minor units (cents), > 0
	Note       string
}

const (
	SortDateDesc   = "date_desc"
	SortAmountDesc = "amount_desc"
	SortAmountAsc  = "amount_asc"
)

type TransactionFilter struct {
	AccountID  *int64 // matches both source and destination of transfers
	CategoryID *int64 // exact category, including split lines
	TagID      *int64
	Type       *string
	Search     string // in description, merchant and notes
	From, To   *time.Time
	Sort       string // Sort* constant, newest first by default
	Limit      int
	Offset     int
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
	Splits        *[]TransactionSplit
}
