package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"homeapp/internal/entity"
	"homeapp/internal/pkg/apperror"
	"homeapp/internal/repository"
)

type BillPeriodRepository struct {
	db *gorm.DB
}

func NewBillPeriodRepository(db *gorm.DB) *BillPeriodRepository {
	return &BillPeriodRepository{db: db}
}

func (r *BillPeriodRepository) Create(ctx context.Context, period *entity.BillPeriod) error {
	return dbOrTx(ctx, r.db).WithContext(ctx).Create(period).Error
}

func (r *BillPeriodRepository) CreateBulk(ctx context.Context, periods []*entity.BillPeriod) error {
	if len(periods) == 0 {
		return nil
	}
	return dbOrTx(ctx, r.db).WithContext(ctx).Create(&periods).Error
}

func (r *BillPeriodRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.BillPeriod, error) {
	var p entity.BillPeriod
	err := dbOrTx(ctx, r.db).WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&p).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperror.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *BillPeriodRepository) FindByBillAndPeriod(ctx context.Context, billID uuid.UUID, period string) (*entity.BillPeriod, error) {
	var p entity.BillPeriod
	err := dbOrTx(ctx, r.db).WithContext(ctx).
		Where("bill_id = ? AND period = ? AND deleted_at IS NULL", billID, period).
		First(&p).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperror.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *BillPeriodRepository) ListByBillID(ctx context.Context, billID uuid.UUID) ([]*entity.BillPeriod, error) {
	var periods []*entity.BillPeriod
	err := dbOrTx(ctx, r.db).WithContext(ctx).
		Where("bill_id = ? AND deleted_at IS NULL", billID).
		Order("period ASC").
		Find(&periods).Error
	return periods, err
}

func (r *BillPeriodRepository) LatestByBillID(ctx context.Context, billID uuid.UUID) (*entity.BillPeriod, error) {
	var p entity.BillPeriod
	err := dbOrTx(ctx, r.db).WithContext(ctx).
		Where("bill_id = ? AND deleted_at IS NULL", billID).
		Order("period DESC").
		First(&p).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperror.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *BillPeriodRepository) NextUpcomingByBillID(ctx context.Context, billID uuid.UUID) (*entity.BillPeriod, error) {
	var p entity.BillPeriod
	err := dbOrTx(ctx, r.db).WithContext(ctx).
		Where("bill_id = ? AND status IN ('upcoming', 'overdue') AND deleted_at IS NULL", billID).
		Order("period ASC").
		First(&p).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperror.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *BillPeriodRepository) MarkPaid(ctx context.Context, id, transactionID uuid.UUID, paidAt time.Time) error {
	return dbOrTx(ctx, r.db).WithContext(ctx).Model(&entity.BillPeriod{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":         entity.BillPeriodPaid,
			"transaction_id": transactionID,
			"paid_at":        paidAt,
		}).Error
}

func (r *BillPeriodRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status entity.BillPeriodStatus) error {
	return dbOrTx(ctx, r.db).WithContext(ctx).Model(&entity.BillPeriod{}).
		Where("id = ?", id).
		Update("status", status).Error
}

func (r *BillPeriodRepository) ListDueForReminder(ctx context.Context, today time.Time) ([]*repository.BillPeriodWithBill, error) {
	var items []*repository.BillPeriodWithBill
	err := dbOrTx(ctx, r.db).WithContext(ctx).
		Table("bill_periods bp").
		Select("bp.*, b.name AS bill_name, b.household_id AS household_id, b.category_id AS category_id, b.amount AS amount").
		Joins("JOIN bills b ON b.id = bp.bill_id").
		Where(`bp.status = 'upcoming' AND bp.deleted_at IS NULL
			AND bp.due_date >= ?
			AND bp.due_date <= (?::date + (b.reminder_days_before || ' days')::interval)`,
			today, today).
		Find(&items).Error
	return items, err
}

func (r *BillPeriodRepository) ListOverdue(ctx context.Context, today time.Time) ([]*entity.BillPeriod, error) {
	var periods []*entity.BillPeriod
	err := dbOrTx(ctx, r.db).WithContext(ctx).
		Where("status = 'upcoming' AND due_date < ? AND deleted_at IS NULL", today).
		Find(&periods).Error
	return periods, err
}

func (r *BillPeriodRepository) ListUpcomingForHousehold(ctx context.Context, householdID uuid.UUID, days int) ([]*repository.BillPeriodWithBill, error) {
	today := time.Now().Truncate(24 * time.Hour)
	until := today.AddDate(0, 0, days)

	var items []*repository.BillPeriodWithBill
	err := dbOrTx(ctx, r.db).WithContext(ctx).
		Table("bill_periods bp").
		Select("bp.*, b.name AS bill_name, b.household_id AS household_id, b.category_id AS category_id, b.amount AS amount").
		Joins("JOIN bills b ON b.id = bp.bill_id").
		Where(`b.household_id = ? AND bp.status = 'upcoming' AND bp.deleted_at IS NULL
			AND bp.due_date >= ? AND bp.due_date <= ?`, householdID, today, until).
		Order("bp.due_date ASC").
		Find(&items).Error
	return items, err
}
