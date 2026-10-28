package model

import "time"

// CategoryAmount is a top-level category's total with its subcategories rolled in.
type CategoryAmount struct {
	CategoryID       *int64 // nil for the slice holding everything past the top few
	Amount           int64
	Percent          float64
	TransactionCount int
	Children         []CategoryAmount // set only when more than one category fed Amount
}

type Dashboard struct {
	Year, Month        int
	Income             int64
	Spent              int64
	TransferredOut     int64
	SpendingByCategory []CategoryAmount
}

type CashFlowMonth struct {
	Year, Month     int
	Income, Expense int64
}

type CashFlow struct {
	From, To *time.Time // the data's own span when not given
	Months   []CashFlowMonth
}

type MonthAmount struct {
	Year, Month int
	Amount      int64
}

type CategorySpending struct {
	CategoryID       int64
	From, To         *time.Time
	Total            int64
	TransactionCount int
	AveragePerMonth  int64
	Months           []MonthAmount
}

type CategoryRanking struct {
	From, To *time.Time
	Total    int64
	Items    []CategoryAmount
}

type DatedAmount struct {
	Date   time.Time
	Amount int64
}

type NetWorthPart struct {
	Key     string // "cash", an asset class, or "asset:<id>"
	Name    string
	Amount  int64
	Percent float64
}

type CapitalRoleSummary struct {
	Role            string
	TotalValue      int64
	MonthlyCashFlow int64
	Count           int
}

type RiskLevelSummary struct {
	Level      string
	TotalValue int64
	Percent    float64        // of all capital
	Items      []NetWorthPart // percent of this tier each
}

type NetWorth struct {
	Current       int64
	ChangeAmount  int64
	ChangePercent *float64 // nil when the range starts at zero
	Series        []DatedAmount
	Breakdown     []NetWorthPart
	CapitalRoles  []CapitalRoleSummary
	RiskLevels    []RiskLevelSummary
}

// Notice is an alert or a piece of advice; clients word it from Key and Params.
type Notice struct {
	Key    string
	Level  string // severity for alerts, tone for advice
	Params map[string]any
}

// Percent is part of whole, 0 when whole is 0.
func Percent(part, whole int64) float64 {
	if whole == 0 {
		return 0
	}
	return float64(part) / float64(whole) * 100
}
