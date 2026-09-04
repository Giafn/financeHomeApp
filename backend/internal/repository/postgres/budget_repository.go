package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"homeapp/internal/entity"
	"homeapp/internal/pkg/apperror"
	cycleperiod "homeapp/internal/pkg/period"
	"homeapp/internal/repository"
)

type BudgetRepository struct {
	db *gorm.DB
}

func NewBudgetRepository(db *gorm.DB) *BudgetRepository {
	return &BudgetRepository{db: db}
}

func (r *BudgetRepository) Create(ctx context.Context, budget *entity.Budget) error {
	return dbOrTx(ctx, r.db).WithContext(ctx).Create(budget).Error
}

func (r *BudgetRepository) FindByID(ctx context.Context, id, householdID uuid.UUID) (*entity.Budget, error) {
	var b entity.Budget
	err := dbOrTx(ctx, r.db).WithContext(ctx).
		Where("id = ? AND household_id = ? AND deleted_at IS NULL", id, householdID).
		First(&b).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperror.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *BudgetRepository) FindByCategoryPeriod(ctx context.Context, householdID, categoryID uuid.UUID, period string) (*entity.Budget, error) {
	var b entity.Budget
	err := dbOrTx(ctx, r.db).WithContext(ctx).
		Where("household_id = ? AND category_id = ? AND period = ? AND deleted_at IS NULL", householdID, categoryID, period).
		First(&b).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperror.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *BudgetRepository) ListByPeriod(ctx context.Context, householdID uuid.UUID, period string, cycleStartDay int) ([]*repository.BudgetWithSpent, error) {
	start, end, err := cycleperiod.CyclePeriodWindow(period, cycleStartDay)
	if err != nil {
		return nil, fmt.Errorf("format period tidak valid: %w", err)
	}

	var items []*repository.BudgetWithSpent
	err = dbOrTx(ctx, r.db).WithContext(ctx).
		Table("budgets b").
		Select(`b.*, c.name AS category_name, COALESCE(t.spent, 0) AS spent`).
		Joins("JOIN categories c ON c.id = b.category_id").
		Joins(`LEFT JOIN (
			SELECT category_id, SUM(amount) AS spent
			FROM transactions
			WHERE household_id = ? AND type = 'expense' AND deleted_at IS NULL
			  AND transaction_date >= ? AND transaction_date < ?
			GROUP BY category_id
		) t ON t.category_id = b.category_id`, householdID, start, end).
		Where("b.household_id = ? AND b.period = ? AND b.deleted_at IS NULL", householdID, period).
		Order("c.name").
		Find(&items).Error
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (r *BudgetRepository) ListRawByPeriod(ctx context.Context, householdID uuid.UUID, period string) ([]*entity.Budget, error) {
	var items []*entity.Budget
	err := dbOrTx(ctx, r.db).WithContext(ctx).
		Where("household_id = ? AND period = ? AND deleted_at IS NULL", householdID, period).
		Find(&items).Error
	return items, err
}

func (r *BudgetRepository) Update(ctx context.Context, budget *entity.Budget) error {
	return dbOrTx(ctx, r.db).WithContext(ctx).Model(budget).
		Select("amount").
		Updates(budget).Error
}

func (r *BudgetRepository) Delete(ctx context.Context, id, householdID uuid.UUID) error {
	return dbOrTx(ctx, r.db).WithContext(ctx).
		Where("id = ? AND household_id = ?", id, householdID).
		Delete(&entity.Budget{}).Error
}

