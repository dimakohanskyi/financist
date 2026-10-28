package repository

import (
	"context"
	"testing"
	"time"

	"github.com/dimakohanskyi/financist/internal/model"
)

func TestAssetValuations(t *testing.T) {
	r := NewAssetRepo(newTestDB(t))
	ctx := context.Background()
	jan := time.Date(2025, 1, 1, 9, 0, 0, 0, time.UTC)
	a := &model.Asset{Name: "Flat", AssetClass: "real_estate", Currency: "USD", CapitalRole: model.CapitalIncome, RiskLevel: model.RiskLow}
	if err := r.Create(ctx, a, 100, jan); err != nil {
		t.Fatal(err)
	}
	for _, v := range []model.AssetValuation{
		{AssetID: a.ID, Value: 150, AsOfDate: jan.AddDate(0, 1, 0)},
		{AssetID: a.ID, Value: 120, AsOfDate: jan.AddDate(0, 1, 0).Add(5 * time.Hour)}, // same day overwrites
	} {
		if err := r.AddValuation(ctx, v); err != nil {
			t.Fatal(err)
		}
	}
	got, err := r.GetByID(ctx, a.ID)
	history, _ := r.Valuations(ctx, &a.ID)
	if err != nil || got.CurrentValue != 120 || len(history) != 2 {
		t.Fatalf("got %+v with %d valuations, err %v", got, len(history), err)
	}
}
