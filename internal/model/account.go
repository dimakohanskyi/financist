package model

type Account struct {
	ID          int64
	Name        string
	AccountType string
	Currency    string
	Color       *string // hex like #2a78d6
	IsArchived  bool
}

// AccountUpdate holds the fields to change; nil means "leave as is".
type AccountUpdate struct {
	Name        *string
	AccountType *string
	Currency    *string
	Color       *string // "" clears it
	IsArchived  *bool
}
