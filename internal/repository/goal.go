package repository

import (
	"context"
	"errors"
	"time"

	"github.com/dimakohanskyi/financist/internal/model"
	"gorm.io/gorm"
)

// goalRow is the GORM mapping of the goals table.
type goalRow struct {
	ID            int64 `gorm:"primaryKey"`
	Name          string
	TargetAmount  int64
	TargetDate    *time.Time
	CurrentAmount int64 `gorm:"->" json:"-"` // computed by withCurrent
}

func (goalRow) TableName() string { return "goals" }

// goalContributionRow is the GORM mapping of the goal_contributions table.
type goalContributionRow struct {
	ID     int64 `gorm:"primaryKey"`
	GoalID int64
	Amount int64
	Date   time.Time
	Note   string
}

func (goalContributionRow) TableName() string { return "goal_contributions" }

type GoalRepo struct{ db *gorm.DB }

func NewGoalRepo(db *gorm.DB) *GoalRepo { return &GoalRepo{db} }

func (r *GoalRepo) withCurrent(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).Model(&goalRow{}).
		Select("goals.*, (SELECT COALESCE(SUM(amount), 0) FROM goal_contributions WHERE goal_id = goals.id) AS current_amount")
}

func (r *GoalRepo) Create(ctx context.Context, g *model.Goal) error {
	row := goalRow{Name: g.Name, TargetAmount: g.TargetAmount, TargetDate: g.TargetDate}
	if row.TargetDate != nil {
		d := model.Day(*row.TargetDate)
		row.TargetDate = &d
	}
	if err := gorm.G[goalRow](r.db).Create(ctx, &row); err != nil {
		return err
	}
	g.ID = row.ID
	return nil
}

func (r *GoalRepo) GetByID(ctx context.Context, id int64) (model.Goal, error) {
	var row goalRow
	err := r.withCurrent(ctx).Where("id = ?", id).Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = model.ErrNotFound
	}
	return model.Goal(row), err
}

func (r *GoalRepo) List(ctx context.Context) ([]model.Goal, error) {
	var rows []goalRow
	err := r.withCurrent(ctx).Order("id").Find(&rows).Error
	out := make([]model.Goal, len(rows))
	for i, row := range rows {
		out[i] = model.Goal(row)
	}
	return out, err
}

func (r *GoalRepo) Update(ctx context.Context, id int64, u model.GoalUpdate) (model.Goal, error) {
	fields := map[string]any{}
	if u.Name != nil {
		fields["name"] = *u.Name
	}
	if u.TargetAmount != nil {
		fields["target_amount"] = *u.TargetAmount
	}
	if u.TargetDate != nil {
		fields["target_date"] = model.Day(*u.TargetDate)
	}
	if u.ClearTargetDate {
		fields["target_date"] = nil
	}
	if len(fields) > 0 {
		if err := r.db.WithContext(ctx).Model(&goalRow{}).Where("id = ?", id).Updates(fields).Error; err != nil {
			return model.Goal{}, err
		}
	}
	return r.GetByID(ctx, id)
}

func (r *GoalRepo) Delete(ctx context.Context, id int64) error {
	res := r.db.WithContext(ctx).Delete(&goalRow{}, id)
	if res.Error == nil && res.RowsAffected == 0 {
		return model.ErrNotFound
	}
	return res.Error
}

func (r *GoalRepo) AddContribution(ctx context.Context, c *model.GoalContribution) error {
	row := goalContributionRow(*c)
	row.Date = model.Day(row.Date)
	if err := gorm.G[goalContributionRow](r.db).Create(ctx, &row); err != nil {
		return err
	}
	c.ID = row.ID
	return nil
}
