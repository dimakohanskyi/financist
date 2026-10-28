package repository

import (
	"context"
	"errors"
	"time"

	"github.com/dimakohanskyi/financist/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// assetRow is the GORM mapping of the assets table.
type assetRow struct {
	ID              int64 `gorm:"primaryKey"`
	Name            string
	AssetClass      string
	Currency        string
	Notes           string
	CapitalRole     string
	MonthlyCashFlow *int64
	RiskLevel       string
	CurrentValue    int64      `gorm:"->" json:"-"` // filled by withLatest
	AsOfDate        *time.Time `gorm:"->" json:"-"`
}

func (assetRow) TableName() string { return "assets" }

// valuationRow is the GORM mapping of the asset_valuations table.
type valuationRow struct {
	ID       int64 `gorm:"primaryKey"`
	AssetID  int64
	Value    int64
	AsOfDate time.Time
}

func (valuationRow) TableName() string { return "asset_valuations" }

type AssetRepo struct{ db *gorm.DB }

func NewAssetRepo(db *gorm.DB) *AssetRepo { return &AssetRepo{db} }

// withLatest joins each asset's latest valuation.
func (r *AssetRepo) withLatest(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).Model(&assetRow{}).
		Select("assets.*, COALESCE(v.value, 0) AS current_value, v.as_of_date").
		Joins(`LEFT JOIN asset_valuations v ON v.id = (
			SELECT id FROM asset_valuations WHERE asset_id = assets.id ORDER BY as_of_date DESC LIMIT 1)`)
}

// Create saves the asset with its first valuation.
func (r *AssetRepo) Create(ctx context.Context, a *model.Asset, value int64, asOf time.Time) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		row := assetRow(*a)
		if err := tx.Create(&row).Error; err != nil {
			return err
		}
		a.ID = row.ID
		return tx.Create(&valuationRow{AssetID: row.ID, Value: value, AsOfDate: model.Day(asOf)}).Error
	})
}

func (r *AssetRepo) GetByID(ctx context.Context, id int64) (model.Asset, error) {
	var row assetRow
	err := r.withLatest(ctx).Where("assets.id = ?", id).Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = model.ErrNotFound
	}
	return model.Asset(row), err
}

func (r *AssetRepo) List(ctx context.Context) ([]model.Asset, error) {
	var rows []assetRow
	err := r.withLatest(ctx).Order("assets.name, assets.id").Find(&rows).Error
	out := make([]model.Asset, len(rows))
	for i, row := range rows {
		out[i] = model.Asset(row)
	}
	return out, err
}

func (r *AssetRepo) Update(ctx context.Context, id int64, u model.AssetUpdate) (model.Asset, error) {
	fields := map[string]any{}
	for col, v := range map[string]*string{
		"name": u.Name, "asset_class": u.AssetClass, "currency": u.Currency,
		"notes": u.Notes, "capital_role": u.CapitalRole, "risk_level": u.RiskLevel,
	} {
		if v != nil {
			fields[col] = *v
		}
	}
	if u.MonthlyCashFlow != nil {
		fields["monthly_cash_flow"] = *u.MonthlyCashFlow
	}
	if u.ClearMonthlyCashFlow {
		fields["monthly_cash_flow"] = nil
	}
	if len(fields) > 0 {
		if err := r.db.WithContext(ctx).Model(&assetRow{}).Where("id = ?", id).Updates(fields).Error; err != nil {
			return model.Asset{}, err
		}
	}
	return r.GetByID(ctx, id)
}

func (r *AssetRepo) Delete(ctx context.Context, id int64) error {
	res := r.db.WithContext(ctx).Delete(&assetRow{}, id)
	if res.Error == nil && res.RowsAffected == 0 {
		return model.ErrNotFound
	}
	return res.Error
}

// AddValuation records the value as of a day; the same day again overwrites it.
func (r *AssetRepo) AddValuation(ctx context.Context, v model.AssetValuation) error {
	row := valuationRow{AssetID: v.AssetID, Value: v.Value, AsOfDate: model.Day(v.AsOfDate)}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "asset_id"}, {Name: "as_of_date"}},
		DoUpdates: clause.AssignmentColumns([]string{"value"}),
	}).Create(&row).Error
}

// Valuations returns the history of one asset, or of all when assetID is nil, oldest first.
func (r *AssetRepo) Valuations(ctx context.Context, assetID *int64) ([]model.AssetValuation, error) {
	q := gorm.G[valuationRow](r.db).Order("as_of_date, id")
	if assetID != nil {
		q = q.Where("asset_id = ?", *assetID)
	}
	rows, err := q.Find(ctx)
	out := make([]model.AssetValuation, len(rows))
	for i, row := range rows {
		out[i] = model.AssetValuation(row)
	}
	return out, err
}
