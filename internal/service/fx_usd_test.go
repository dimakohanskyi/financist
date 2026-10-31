package service

import (
	"context"
	"testing"
	"time"

	"github.com/dimakohanskyi/financist/internal/model"
)

type ledgerStub struct{ ts []model.Transaction }

func (l ledgerStub) Ledger(context.Context) ([]model.Transaction, error) { return l.ts, nil }
func (l ledgerStub) Contributions(context.Context, model.ContributionFilter) ([]model.Contribution, error) {
	return nil, nil
}

// Amounts stay in their own currency unless the request asks for USD.
func TestWrapConvertsOnlyWhenWanted(t *testing.T) {
	day := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	fx := NewFxService(nil, nil)
	fx.cache = model.Rates{"EUR": {{Date: day, Quote: "EUR", Rate: 0.5}}}
	src := ledgerStub{ts: []model.Transaction{{Amount: 100, Currency: "EUR", Date: day}}}

	ctx := context.Background()
	plain, _ := fx.Wrap(src).Ledger(ctx)
	if plain[0].Amount != 100 {
		t.Fatalf("without the switch: got %d, want 100", plain[0].Amount)
	}
	usd, _ := fx.Wrap(src).Ledger(WithUSD(ctx))
	if usd[0].Amount != 200 {
		t.Fatalf("with the switch: got %d, want 200", usd[0].Amount)
	}
}
