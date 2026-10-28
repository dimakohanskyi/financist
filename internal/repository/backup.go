package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const backupVersion = 1

// Backup is a full snapshot of the database; its JSON is the backup file format.
type Backup struct {
	Version           int
	ExportedAt        time.Time
	Accounts          []accountRow
	Categories        []categoryRow
	Tags              []tagRow
	Transactions      []transactionRow
	TransactionTags   []transactionTagRow
	TransactionSplits []splitRow
	RecurringTxs      []recurringRow
	Budgets           []budgetRow
	Goals             []goalRow
	GoalContributions []goalContributionRow
	Assets            []assetRow
	AssetValuations   []valuationRow
	Settings          []settingsRow
}

type BackupRepo struct{ db *gorm.DB }

func NewBackupRepo(db *gorm.DB) *BackupRepo { return &BackupRepo{db} }

func (r *BackupRepo) Export(ctx context.Context) (*Backup, error) {
	b := &Backup{Version: backupVersion, ExportedAt: time.Now().UTC()}
	db := r.db.WithContext(ctx)
	for _, dst := range []any{
		&b.Accounts, &b.Categories, &b.Tags, &b.Transactions, &b.TransactionTags, &b.TransactionSplits,
		&b.RecurringTxs, &b.Budgets, &b.Goals, &b.GoalContributions, &b.Assets, &b.AssetValuations, &b.Settings,
	} {
		if err := db.Find(dst).Error; err != nil {
			return nil, err
		}
	}
	return b, nil
}

// Import replaces all data with the backup. It runs in one transaction with
// foreign keys checked at commit, so a broken file changes nothing.
func (r *BackupRepo) Import(ctx context.Context, b *Backup) error {
	if b.Version != backupVersion {
		return fmt.Errorf("unsupported backup version %d", b.Version)
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("PRAGMA defer_foreign_keys = ON").Error; err != nil {
			return err
		}
		// children first
		for _, table := range []string{
			"app_settings", "asset_valuations", "assets", "goal_contributions", "goals", "budgets",
			"recurring_transactions", "transaction_splits", "transaction_tags", "transactions", "tags", "categories", "accounts",
		} {
			if err := tx.Exec("DELETE FROM " + table).Error; err != nil {
				return err
			}
		}
		for _, rows := range []any{
			b.Accounts, b.Categories, b.Tags, b.Transactions, b.TransactionTags, b.TransactionSplits,
			b.RecurringTxs, b.Budgets, b.Goals, b.GoalContributions, b.Assets, b.AssetValuations, b.Settings,
		} {
			if err := insertAll(tx, rows); err != nil {
				return err
			}
		}
		if len(b.Settings) == 0 {
			return tx.Exec("INSERT INTO app_settings(id) VALUES (1)").Error
		}
		return nil
	})
}

func insertAll(tx *gorm.DB, rows any) error {
	err := tx.Omit(clause.Associations).CreateInBatches(rows, 500).Error
	if errors.Is(err, gorm.ErrEmptySlice) {
		return nil
	}
	return err
}
