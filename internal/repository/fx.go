package repository

import (
	"context"
	"time"

	"github.com/dimakohanskyi/financist/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// fxRateRow is the GORM mapping of the fx_rates table.
type fxRateRow struct {
	Date      time.Time
	Quote     string
	Rate      float64
	FetchedAt time.Time
}

func (fxRateRow) TableName() string { return "fx_rates" }

type FxRepo struct{ db *gorm.DB }

func NewFxRepo(db *gorm.DB) *FxRepo { return &FxRepo{db} }

// Save stores the rates, replacing any already held for the same day and quote.
func (r *FxRepo) Save(ctx context.Context, rates []model.FxRate, fetchedAt time.Time) error {
	if len(rates) == 0 {
		return nil
	}
	rows := make([]fxRateRow, len(rates))
	for i, x := range rates {
		rows[i] = fxRateRow{Date: model.Day(x.Date), Quote: x.Quote, Rate: x.Rate, FetchedAt: fetchedAt.UTC()}
	}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "date"}, {Name: "quote"}},
		DoUpdates: clause.AssignmentColumns([]string{"rate", "fetched_at"}),
	}).CreateInBatches(rows, 500).Error
}

// All returns every stored rate, oldest first, and when the newest was fetched (nil if none).
func (r *FxRepo) All(ctx context.Context) ([]model.FxRate, *time.Time, error) {
	var rows []fxRateRow
	if err := r.db.WithContext(ctx).Order("date").Find(&rows).Error; err != nil {
		return nil, nil, err
	}
	out := make([]model.FxRate, len(rows))
	var fetched *time.Time
	for i, row := range rows {
		out[i] = model.FxRate{Date: row.Date, Quote: row.Quote, Rate: row.Rate}
		if fetched == nil || row.FetchedAt.After(*fetched) {
			fetched = &rows[i].FetchedAt
		}
	}
	return out, fetched, nil
}
