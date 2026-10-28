package model

const (
	CategoryIncome  = "income"
	CategoryExpense = "expense"
)

type Category struct {
	ID        int64
	Name      string
	Kind      string
	Icon      *string // lucide icon name
	Color     string  // hex like #2a78d6
	SortOrder int
	IsDefault bool   // seeded by the migration
	ParentID  *int64 // subcategories are one level deep
}

// CategoryUpdate holds the fields to change; nil means "leave as is".
type CategoryUpdate struct {
	Name        *string
	Icon        *string // "" clears it
	Color       *string
	SortOrder   *int
	ParentID    *int64
	ClearParent bool // makes it top-level
}
