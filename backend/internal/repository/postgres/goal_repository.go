package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"homeapp/internal/entity"
	"homeapp/internal/pkg/apperror"
)

type GoalRepository struct {
	db *gorm.DB
}

func NewGoalRepository(db *gorm.DB) *GoalRepository {
	return &GoalRepository{db: db}
}

func (r *GoalRepository) Create(ctx context.Context, goal *entity.Goal) error {
	return dbOrTx(ctx, r.db).WithContext(ctx).Create(goal).Error
}

func (r *GoalRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.Goal, error) {
	var g entity.Goal
	err := dbOrTx(ctx, r.db).WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&g).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperror.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &g, nil
}

func (r *GoalRepository) FindByIDAndHousehold(ctx context.Context, id, householdID uuid.UUID) (*entity.Goal, error) {
	var g entity.Goal
	err := dbOrTx(ctx, r.db).WithContext(ctx).
		Where("id = ? AND household_id = ? AND deleted_at IS NULL", id, householdID).
		First(&g).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperror.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &g, nil
}

func (r *GoalRepository) ListByHouseholdAndStatus(ctx context.Context, householdID uuid.UUID, status string) ([]*entity.Goal, error) {
	var goals []*entity.Goal
	query := dbOrTx(ctx, r.db).WithContext(ctx).Where("household_id = ? AND deleted_at IS NULL", householdID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	err := query.Order("created_at DESC").Find(&goals).Error
	return goals, err
}

func (r *GoalRepository) Update(ctx context.Context, goal *entity.Goal) error {
	return dbOrTx(ctx, r.db).WithContext(ctx).Model(goal).
		Select("name", "icon", "target_amount", "target_date", "status").
		Updates(goal).Error
}

func (r *GoalRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status entity.GoalStatus) error {
	return dbOrTx(ctx, r.db).WithContext(ctx).Model(&entity.Goal{}).
		Where("id = ?", id).
		Update("status", status).Error
}

func (r *GoalRepository) Delete(ctx context.Context, id, householdID uuid.UUID) error {
	return dbOrTx(ctx, r.db).WithContext(ctx).
		Where("id = ? AND household_id = ?", id, householdID).
		Delete(&entity.Goal{}).Error
}

func (r *GoalRepository) SumTransferAmount(ctx context.Context, goalID, linkedAccountID uuid.UUID, asIncoming bool) (float64, error) {
	accountColumn := "account_id"
	if asIncoming {
		accountColumn = "destination_account_id"
	}

	var sum float64
	err := dbOrTx(ctx, r.db).WithContext(ctx).
		Table("transactions").
		Where("goal_id = ? AND type = 'transfer' AND deleted_at IS NULL AND "+accountColumn+" = ?", goalID, linkedAccountID).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&sum).Error
	return sum, err
}

func (r *GoalRepository) HasContributions(ctx context.Context, goalID uuid.UUID) (bool, error) {
	var count int64
	err := dbOrTx(ctx, r.db).WithContext(ctx).
		Table("transactions").
		Where("goal_id = ? AND deleted_at IS NULL", goalID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
