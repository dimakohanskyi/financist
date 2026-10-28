package model

import "time"

const (
	FrequencyWeekly  = "weekly"
	FrequencyMonthly = "monthly"
	FrequencyYearly  = "yearly"
)

// Recurring is a template for a transaction that repeats on a schedule.
type Recurring struct {
	ID             int64
	AccountID      int64
	ToAccountID    *int64
	CategoryID     *int64
	Type           string
	Amount         int64
	Description    string
	Merchant       string
	Notes          string
	Frequency      string
	AnchorDate     time.Time  // first due date
	LastPostedDate *time.Time // the schedule only moves when posted
	IsActive       bool
}

// RecurringUpdate holds the fields to change; nil means "leave as is".
type RecurringUpdate struct {
	AccountID     *int64
	ToAccountID   *int64
	CategoryID    *int64
	ClearCategory bool
	Type          *string
	Amount        *int64
	Description   *string
	Merchant      *string
	Notes         *string
	Frequency     *string
	AnchorDate    *time.Time
	IsActive      *bool
}

// NextDue is the anchor date until the first post, then one period after the last one.
func (r Recurring) NextDue() time.Time {
	if r.LastPostedDate == nil {
		return r.AnchorDate
	}
	switch r.Frequency {
	case FrequencyWeekly:
		return r.LastPostedDate.AddDate(0, 0, 7)
	case FrequencyYearly:
		return AddMonths(*r.LastPostedDate, 12)
	default:
		return AddMonths(*r.LastPostedDate, 1)
	}
}

// Transaction builds the transaction this template posts on date.
func (r Recurring) Transaction(date time.Time) Transaction {
	return Transaction{
		AccountID:   r.AccountID,
		ToAccountID: r.ToAccountID,
		CategoryID:  r.CategoryID,
		Type:        r.Type,
		Amount:      r.Amount,
		Description: r.Description,
		Merchant:    r.Merchant,
		Notes:       r.Notes,
		Date:        date,
	}
}

// Day truncates t to midnight UTC, for fields that are dates rather than moments.
func Day(t time.Time) time.Time {
	y, m, d := t.UTC().Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

// AddMonths clamps to the month's last day: Jan 31 + 1 month is Feb 28, not Mar 3.
func AddMonths(t time.Time, n int) time.Time {
	first := time.Date(t.Year(), t.Month()+time.Month(n), 1, 0, 0, 0, 0, t.Location())
	last := first.AddDate(0, 1, -1).Day()
	return time.Date(first.Year(), first.Month(), min(t.Day(), last), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
}
