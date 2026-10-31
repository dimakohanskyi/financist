package model

import "time"

// Budget is a monthly spending limit on an expense category and its subcategories.
type Budget struct {
	ID           int64
	CategoryID   int64
	MonthlyLimit int64 // minor units (cents)
}

type BudgetStatus struct {
	Budget
	Spent int64
}

// Contribution is what one transaction put into one category: the whole
// amount for a plain transaction, one line's amount for a split.
type Contribution struct {
	TransactionID int64
	CategoryID    int64
	Amount        int64
	Currency      string
	Date          time.Time
}

type ContributionFilter struct {
	Type      string // "" for any
	From, To  *time.Time
	AccountID *int64 // nil for every account
}

// MonthRange returns the first and last moment of a month in UTC.
func MonthRange(year, month int) (from, to time.Time) {
	from = time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	return from, from.AddDate(0, 1, 0).Add(-time.Nanosecond)
}
