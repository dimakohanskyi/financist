package repository

import (
	"context"
	"testing"
	"time"

	"github.com/dimakohanskyi/financist/internal/model"
)

func TestGoalCurrentAmount(t *testing.T) {
	r := NewGoalRepo(newTestDB(t))
	ctx := context.Background()
	g := &model.Goal{Name: "Trip", TargetAmount: 1000}
	if err := r.Create(ctx, g); err != nil {
		t.Fatal(err)
	}
	for _, amount := range []int64{700, -200} {
		if err := r.AddContribution(ctx, &model.GoalContribution{GoalID: g.ID, Amount: amount, Date: time.Now()}); err != nil {
			t.Fatal(err)
		}
	}
	got, err := r.GetByID(ctx, g.ID)
	if err != nil || got.CurrentAmount != 500 {
		t.Fatalf("got %+v, err %v", got, err)
	}
}
