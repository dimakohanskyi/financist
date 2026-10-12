package model

type Account struct {
	ID          int64
	Name        string
	AccountType string
	Currency    string
	IsArchived  bool
}

// AccountUpdate holds the fields to change; nil means "leave as is".
type AccountUpdate struct {
	Name        *string
	AccountType *string
	Currency    *string
	IsArchived  *bool
}
