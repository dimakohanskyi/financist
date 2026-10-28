package model

type Settings struct {
	Currency                        string
	NegativeCashFlowThresholdMonths int
	NetWorthDeclineThresholdMonths  int
	RiskyAllocationThresholdPercent int
	IdleCashThresholdAmount         int64 // minor units (cents)
	IdleCashThresholdDays           int
}

// SettingsUpdate holds the fields to change; nil means "leave as is".
type SettingsUpdate struct {
	Currency                        *string
	NegativeCashFlowThresholdMonths *int
	NetWorthDeclineThresholdMonths  *int
	RiskyAllocationThresholdPercent *int
	IdleCashThresholdAmount         *int64
	IdleCashThresholdDays           *int
}
