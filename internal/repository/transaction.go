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
	Tags        []tagRow   `gorm:"many2many:transaction_tags;joinForeignKey:TransactionID;joinReferences:TagID"`
	Splits      []splitRow `gorm:"foreignKey:TransactionID"`
}

func (transactionRow) TableName() string { return "transactions" }

// splitRow is the GORM mapping of the transaction_splits table.
type splitRow struct {
	ID            int64 `gorm:"primaryKey"`
	TransactionID int64
	CategoryID    *int64
	Amount        int64
	Note          string
}

func (splitRow) TableName() string { return "transaction_splits" }

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
		Splits:      make([]model.TransactionSplit, len(r.Splits)),
	}
	for i, tag := range r.Tags {
		t.Tags[i] = model.Tag(tag)
	}
	for i, sp := range r.Splits {
		t.Splits[i] = model.TransactionSplit{ID: sp.ID, CategoryID: sp.CategoryID, Amount: sp.Amount, Note: sp.Note}
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

// create inserts the transaction with its tags and splits using tx.
func create(tx *gorm.DB, t *model.Transaction) error {
	row := toTransactionRow(t)
	if err := tx.Omit(clause.Associations).Create(&row).Error; err != nil {
		return err
	}
	t.ID = row.ID
	return saveLinks(tx, t)
}

// saveLinks replaces the transaction's tags and splits.
func saveLinks(tx *gorm.DB, t *model.Transaction) error {
	if err := saveTags(tx, t); err != nil {
		return err
	}
	if err := tx.Where("transaction_id = ?", t.ID).Delete(&splitRow{}).Error; err != nil {
		return err
	}
	if len(t.Splits) == 0 {
		return nil
	}
	rows := make([]splitRow, len(t.Splits))
	for i, sp := range t.Splits {
		rows[i] = splitRow{TransactionID: t.ID, CategoryID: sp.CategoryID, Amount: sp.Amount, Note: sp.Note}
	}
	if err := tx.Create(&rows).Error; err != nil {
		return err
	}
	for i := range rows {
		t.Splits[i].ID = rows[i].ID
	}
	return nil
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
	err := withLinks(r.db.WithContext(ctx)).First(&row, id).Error
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
	if err := withLinks(q).Order("date DESC, id DESC").Limit(f.Limit).Offset(f.Offset).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	out := make([]model.Transaction, len(rows))
	for i, row := range rows {
		out[i] = row.toModel()
	}
	return out, total, nil
}

func withLinks(q *gorm.DB) *gorm.DB {
	return q.Preload("Tags").Preload("Splits", func(db *gorm.DB) *gorm.DB { return db.Order("id") })
}

// Update overwrites every column, tag and split of an existing transaction.
func (r *TransactionRepo) Update(ctx context.Context, t *model.Transaction) error {
	row := toTransactionRow(t)
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&transactionRow{}).Where("id = ?", t.ID).Select("*").Omit("id", "Tags", "Splits").Updates(&row).Error; err != nil {
			return err
		}
		return saveLinks(tx, t)
	})
}

func (r *TransactionRepo) Delete(ctx context.Context, id int64) error {
	res := r.db.WithContext(ctx).Delete(&transactionRow{}, id)
	if res.Error == nil && res.RowsAffected == 0 {
		return model.ErrNotFound
	}
	return res.Error
}
