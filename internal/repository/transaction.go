package repository

import (
	"context"
	"errors"
	"time"

	"github.com/dimakohanskyi/financist/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// transactionRow is the GORM mapping of the transactions table.
type transactionRow struct {
	ID          int64 `gorm:"primaryKey"`
	AccountID   int64
	ToAccountID *int64
	CategoryID  *int64
	Type        string
	Amount      int64
	Description string
	Merchant    string
	Notes       string
	Date        time.Time
	Tags        []tagRow `gorm:"many2many:transaction_tags;joinForeignKey:TransactionID;joinReferences:TagID"`
}

func (transactionRow) TableName() string { return "transactions" }

func toTransactionRow(t *model.Transaction) transactionRow {
	return transactionRow{
		ID:          t.ID,
		AccountID:   t.AccountID,
		ToAccountID: t.ToAccountID,
		CategoryID:  t.CategoryID,
		Type:        t.Type,
		Amount:      t.Amount,
		Description: t.Description,
		Merchant:    t.Merchant,
		Notes:       t.Notes,
		// stored as text, so one zone keeps date comparisons and ordering right
		Date: t.Date.UTC(),
	}
}

func (r transactionRow) toModel() model.Transaction {
	t := model.Transaction{
		ID:          r.ID,
		AccountID:   r.AccountID,
		ToAccountID: r.ToAccountID,
		CategoryID:  r.CategoryID,
		Type:        r.Type,
		Amount:      r.Amount,
		Description: r.Description,
		Merchant:    r.Merchant,
		Notes:       r.Notes,
		Date:        r.Date,
		Tags:        make([]model.Tag, len(r.Tags)),
	}
	for i, tag := range r.Tags {
		t.Tags[i] = model.Tag(tag)
	}
	return t
}

type TransactionRepo struct{ db *gorm.DB }

func NewTransactionRepo(db *gorm.DB) *TransactionRepo { return &TransactionRepo{db} }

func (r *TransactionRepo) Create(ctx context.Context, t *model.Transaction) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return create(tx, t)
	})
}

// create inserts the transaction with its tags using tx.
func create(tx *gorm.DB, t *model.Transaction) error {
	row := toTransactionRow(t)
	if err := tx.Omit(clause.Associations).Create(&row).Error; err != nil {
		return err
	}
	t.ID = row.ID
	return saveTags(tx, t)
}

// saveTags replaces the transaction's tag links.
func saveTags(tx *gorm.DB, t *model.Transaction) error {
	if err := tx.Where("transaction_id = ?", t.ID).Delete(&transactionTagRow{}).Error; err != nil {
		return err
	}
	if len(t.Tags) == 0 {
		return nil
	}
	links := make([]transactionTagRow, len(t.Tags))
	for i, tag := range t.Tags {
		links[i] = transactionTagRow{TransactionID: t.ID, TagID: tag.ID}
	}
	return tx.Create(&links).Error
}

func (r *TransactionRepo) GetByID(ctx context.Context, id int64) (model.Transaction, error) {
	var row transactionRow
	err := r.db.WithContext(ctx).Preload("Tags").First(&row, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = model.ErrNotFound
	}
	return row.toModel(), err
}

// List returns one page (newest first) and the total number of matches.
func (r *TransactionRepo) List(ctx context.Context, f model.TransactionFilter) ([]model.Transaction, int64, error) {
	q := r.db.WithContext(ctx).Model(&transactionRow{})
	if f.AccountID != nil {
		q = q.Where("account_id = ? OR to_account_id = ?", *f.AccountID, *f.AccountID)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []transactionRow
	if err := q.Preload("Tags").Order("date DESC, id DESC").Limit(f.Limit).Offset(f.Offset).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	out := make([]model.Transaction, len(rows))
	for i, row := range rows {
		out[i] = row.toModel()
	}
	return out, total, nil
}

// Update overwrites every column and the tag set of an existing transaction.
func (r *TransactionRepo) Update(ctx context.Context, t *model.Transaction) error {
	row := toTransactionRow(t)
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&transactionRow{}).Where("id = ?", t.ID).Select("*").Omit("id", "Tags").Updates(&row).Error; err != nil {
			return err
		}
		return saveTags(tx, t)
	})
}

func (r *TransactionRepo) Delete(ctx context.Context, id int64) error {
	res := r.db.WithContext(ctx).Delete(&transactionRow{}, id)
	if res.Error == nil && res.RowsAffected == 0 {
		return model.ErrNotFound
	}
	return res.Error
}
