package repository

import (
	"context"

	"github.com/dimakohanskyi/financist/internal/model"
	"gorm.io/gorm"
)

// settingsRow is the GORM mapping of the single-row app_settings table.
type settingsRow struct {
	ID                              int64 `gorm:"primaryKey"`
	Currency                        string
	NegativeCashFlowThresholdMonths int
	NetWorthDeclineThresholdMonths  int
	RiskyAllocationThresholdPercent int
	IdleCashThresholdAmount         int64
	IdleCashThresholdDays           int
}

func (settingsRow) TableName() string { return "app_settings" }

type SettingsRepo struct{ db *gorm.DB }

func NewSettingsRepo(db *gorm.DB) *SettingsRepo { return &SettingsRepo{db} }

func (r *SettingsRepo) Get(ctx context.Context) (model.Settings, error) {
	row, err := gorm.G[settingsRow](r.db).Where("id = 1").First(ctx)
	return model.Settings{
		Currency:                        row.Currency,
		NegativeCashFlowThresholdMonths: row.NegativeCashFlowThresholdMonths,
		NetWorthDeclineThresholdMonths:  row.NetWorthDeclineThresholdMonths,
		RiskyAllocationThresholdPercent: row.RiskyAllocationThresholdPercent,
		IdleCashThresholdAmount:         row.IdleCashThresholdAmount,
		IdleCashThresholdDays:           row.IdleCashThresholdDays,
	}, err
}

func (r *SettingsRepo) Update(ctx context.Context, u model.SettingsUpdate) (model.Settings, error) {
	fields := map[string]any{}
	if u.Currency != nil {
		fields["currency"] = *u.Currency
	}
	for col, v := range map[string]*int{
		"negative_cash_flow_threshold_months": u.NegativeCashFlowThresholdMonths,
		"net_worth_decline_threshold_months":  u.NetWorthDeclineThresholdMonths,
		"risky_allocation_threshold_percent":  u.RiskyAllocationThresholdPercent,
		"idle_cash_threshold_days":            u.IdleCashThresholdDays,
	} {
		if v != nil {
			fields[col] = *v
		}
	}
	if u.IdleCashThresholdAmount != nil {
		fields["idle_cash_threshold_amount"] = *u.IdleCashThresholdAmount
	}
	if len(fields) > 0 {
		if err := r.db.WithContext(ctx).Model(&settingsRow{}).Where("id = 1").Updates(fields).Error; err != nil {
			return model.Settings{}, err
		}
	}
	return r.Get(ctx)
}
