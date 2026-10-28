package model

import "time"

const (
	CapitalIncome  = "income"
	CapitalNeutral = "neutral"
	CapitalDrain   = "drain"

	RiskLow    = "low"
	RiskMedium = "medium"
	RiskHigh   = "high"
)

// Asset is a manually valued part of net worth: investments, property, a car.
type Asset struct {
	ID              int64
	Name            string
	AssetClass      string
	Currency        string
	Notes           string
	CapitalRole     string
	MonthlyCashFlow *int64 // rough, self-reported
	RiskLevel       string
	CurrentValue    int64      // latest valuation, read-only
	AsOfDate        *time.Time // of the latest valuation, read-only
}

// AssetUpdate holds the fields to change; nil means "leave as is".
type AssetUpdate struct {
	Name                 *string
	AssetClass           *string
	Currency             *string
	Notes                *string
	CapitalRole          *string
	MonthlyCashFlow      *int64
	ClearMonthlyCashFlow bool
	RiskLevel            *string
}

type AssetValuation struct {
	ID       int64
	AssetID  int64
	Value    int64
	AsOfDate time.Time
}
