package model

import "time"

type Goal struct {
	ID            int64
	Name          string
	TargetAmount  int64 // minor units (cents)
	TargetDate    *time.Time
	CurrentAmount int64 // sum of contributions, read-only
}

// GoalUpdate holds the fields to change; nil means "leave as is".
type GoalUpdate struct {
	Name            *string
	TargetAmount    *int64
	TargetDate      *time.Time
	ClearTargetDate bool
}

// GoalContribution is a deposit toward a goal, or a withdrawal when negative.
type GoalContribution struct {
	ID     int64
	GoalID int64
	Amount int64
	Date   time.Time
	Note   string
}
