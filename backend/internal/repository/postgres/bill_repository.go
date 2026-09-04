package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"homeapp/internal/entity"
	"homeapp/internal/pkg/apperror"
)

type BillRepository struct {
	db *gorm.DB
}

func NewBillRepository(db *gorm.DB) *BillRepository {
	return &BillRepository{db: db}
}

func (r *BillRepository) Create(ctx context.Context, bill *entity.Bill) error {
	return dbOrTx(ctx, r.db).WithContext(ctx).Create(bill).Error
}

func (r *BillRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.Bill, error) {
	var b entity.Bill
	err := dbOrTx(ctx, r.db).WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&b).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperror.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *BillRepository) FindByIDAndHousehold(ctx context.Context, id, householdID uuid.UUID) (*entity.Bill, error) {
	var b entity.Bill
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

func (r *BillRepository) ListByHousehold(ctx context.Context, householdID uuid.UUID, isActive *bool) ([]*entity.Bill, error) {
	var bills []*entity.Bill
	query := dbOrTx(ctx, r.db).WithContext(ctx).Where("household_id = ? AND deleted_at IS NULL", householdID)
	if isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}
	err := query.Order("created_at DESC").Find(&bills).Error
	return bills, err
}

func (r *BillRepository) ListIndefiniteActive(ctx context.Context) ([]*entity.Bill, error) {
	var bills []*entity.Bill
	err := dbOrTx(ctx, r.db).WithContext(ctx).
		Where("end_period IS NULL AND is_active = true AND deleted_at IS NULL").
		Find(&bills).Error
	return bills, err
}

func (r *BillRepository) Update(ctx context.Context, bill *entity.Bill) error {
	return dbOrTx(ctx, r.db).WithContext(ctx).Model(bill).
		Select("is_active", "name", "amount", "category_id", "reminder_days_before", "due_day", "end_period").
		Updates(bill).Error
}

func (r *BillRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	return dbOrTx(ctx, r.db).WithContext(ctx).Delete(&entity.Bill{}, "id = ?", id).Error
}
