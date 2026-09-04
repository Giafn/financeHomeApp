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

type TransactionRepository struct {
	db *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (r *TransactionRepository) Create(ctx context.Context, transaction *entity.Transaction) error {
	return dbOrTx(ctx, r.db).WithContext(ctx).Create(transaction).Error
}

// baseListQuery join transactions dengan accounts/categories/users supaya nama tersedia
// tanpa request tambahan dari frontend (lihat kontrak API Phase 06).
func (r *TransactionRepository) baseListQuery(ctx context.Context) *gorm.DB {
	return dbOrTx(ctx, r.db).WithContext(ctx).
		Table("transactions t").
		Select(`t.*, a.name AS account_name, c.name AS category_name, u.name AS created_by_name`).
		Joins("JOIN accounts a ON a.id = t.account_id").
		Joins("LEFT JOIN categories c ON c.id = t.category_id").
		Joins("JOIN users u ON u.id = t.created_by").
		Where("t.deleted_at IS NULL")
}

func (r *TransactionRepository) FindByID(ctx context.Context, id, householdID uuid.UUID) (*repository.TransactionListItem, error) {
	var item repository.TransactionListItem
	err := r.baseListQuery(ctx).
		Where("t.id = ? AND t.household_id = ?", id, householdID).
		First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperror.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *TransactionRepository) List(ctx context.Context, householdID uuid.UUID, filter repository.TransactionFilter) ([]*repository.TransactionListItem, int64, error) {
	query := r.baseListQuery(ctx).Where("t.household_id = ?", householdID)

	if filter.Type != "" {
		query = query.Where("t.type = ?", filter.Type)
	}
	if filter.AccountID != nil {
		query = query.Where("(t.account_id = ? OR t.destination_account_id = ?)", *filter.AccountID, *filter.AccountID)
	}
	if filter.CategoryID != nil {
		query = query.Where("t.category_id = ?", *filter.CategoryID)
	}
	if filter.CreatedBy != nil {
		query = query.Where("t.created_by = ?", *filter.CreatedBy)
	}
	if filter.GoalID != nil {
		query = query.Where("t.goal_id = ?", *filter.GoalID)
	}
	if filter.DateFrom != nil {
		query = query.Where("t.transaction_date >= ?", *filter.DateFrom)
	}
	if filter.DateTo != nil {
		query = query.Where("t.transaction_date <= ?", *filter.DateTo)
	}

	var total int64
	countQuery := query.Session(&gorm.Session{})
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page := filter.Page
	if page < 1 {
		page = 1
	}
	limit := filter.Limit
	if limit < 1 {
		limit = 20
	}
	offset := (page - 1) * limit

	var items []*repository.TransactionListItem
	err := query.Order("t.transaction_date DESC, t.created_at DESC").
		Limit(limit).Offset(offset).
		Find(&items).Error
	if err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (r *TransactionRepository) Update(ctx context.Context, transaction *entity.Transaction) error {
	return dbOrTx(ctx, r.db).WithContext(ctx).Model(transaction).
		Select("type", "account_id", "destination_account_id", "category_id", "amount", "admin_fee", "description", "transaction_date", "attachment_url", "goal_id").
		Updates(transaction).Error
}

func (r *TransactionRepository) Delete(ctx context.Context, id, householdID uuid.UUID) error {
	return dbOrTx(ctx, r.db).WithContext(ctx).
		Where("id = ? AND household_id = ?", id, householdID).
		Delete(&entity.Transaction{}).Error
}

func (r *TransactionRepository) LastUsedAccountAndCategory(ctx context.Context, householdID, userID uuid.UUID) (*uuid.UUID, *uuid.UUID, error) {
	var t entity.Transaction
	err := dbOrTx(ctx, r.db).WithContext(ctx).
		Where("household_id = ? AND created_by = ?", householdID, userID).
		Order("created_at DESC").
		First(&t).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil, nil
	}
	if err != nil {
		return nil, nil, err
	}
	return &t.AccountID, t.CategoryID, nil
}

func (r *TransactionRepository) IsLinkedToPaidBillPeriod(ctx context.Context, transactionID uuid.UUID) (bool, error) {
	var count int64
	err := dbOrTx(ctx, r.db).WithContext(ctx).
		Table("bill_periods").
		Where("transaction_id = ? AND status = ?", transactionID, entity.BillPeriodPaid).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *TransactionRepository) MonthlyTrend(ctx context.Context, householdID uuid.UUID, months int) ([]repository.MonthlyTrendItem, error) {
	now := time.Now()
	rangeStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC).AddDate(0, -(months - 1), 0)
	rangeEnd := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC).AddDate(0, 1, 0)

	type row struct {
		Month   string
		Income  float64
		Expense float64
	}
	var rows []row
	err := dbOrTx(ctx, r.db).WithContext(ctx).
		Table("transactions").
		Select(`to_char(transaction_date, 'YYYY-MM') AS month,
			COALESCE(SUM(CASE WHEN type = 'income' THEN amount ELSE 0 END), 0) AS income,
			COALESCE(SUM(CASE WHEN type = 'expense' THEN amount ELSE 0 END), 0) AS expense`).
		Where("household_id = ? AND deleted_at IS NULL AND transaction_date >= ? AND transaction_date < ?", householdID, rangeStart, rangeEnd).
		Group("month").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}

	byMonth := make(map[string]row, len(rows))
	for _, r := range rows {
		byMonth[r.Month] = r
	}

	// Zero-fill bulan yang tidak punya transaksi supaya grafik tetap kontinu (tidak ada gap).
	result := make([]repository.MonthlyTrendItem, 0, months)
	for i := months - 1; i >= 0; i-- {
		m := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC).AddDate(0, -i, 0).Format("2006-01")
		if r, ok := byMonth[m]; ok {
			result = append(result, repository.MonthlyTrendItem{Month: m, Income: r.Income, Expense: r.Expense})
		} else {
			result = append(result, repository.MonthlyTrendItem{Month: m})
		}
	}
	return result, nil
}

func (r *TransactionRepository) MemberBreakdown(ctx context.Context, householdID uuid.UUID, start, end time.Time) ([]repository.MemberBreakdownItem, error) {
	var items []repository.MemberBreakdownItem
	err := dbOrTx(ctx, r.db).WithContext(ctx).
		Table("household_members hm").
		Select(`hm.user_id, u.name,
			COALESCE(SUM(CASE WHEN t.type = 'expense' THEN t.amount ELSE 0 END), 0) AS total_expense,
			COALESCE(SUM(CASE WHEN t.type = 'income' THEN t.amount ELSE 0 END), 0) AS total_income`).
		Joins("JOIN users u ON u.id = hm.user_id").
		Joins(`LEFT JOIN transactions t ON t.created_by = hm.user_id AND t.household_id = hm.household_id
			AND t.deleted_at IS NULL AND t.transaction_date >= ? AND t.transaction_date < ?`, start, end).
		Where("hm.household_id = ? AND hm.deleted_at IS NULL", householdID).
		Group("hm.user_id, u.name").
		Order("total_expense DESC").
		Find(&items).Error
	return items, err
}

func (r *TransactionRepository) CategoryBreakdown(ctx context.Context, householdID uuid.UUID, start, end time.Time, txType string) ([]repository.CategoryBreakdownItem, error) {
	var items []repository.CategoryBreakdownItem
	err := dbOrTx(ctx, r.db).WithContext(ctx).
		Table("transactions t").
		Select("t.category_id, c.name AS category_name, SUM(t.amount) AS total").
		Joins("JOIN categories c ON c.id = t.category_id").
		Where(`t.household_id = ? AND t.type = ? AND t.deleted_at IS NULL
			AND t.transaction_date >= ? AND t.transaction_date < ?`, householdID, txType, start, end).
		Group("t.category_id, c.name").
		Order("total DESC").
		Find(&items).Error
	return items, err
}

func (r *TransactionRepository) TotalByType(ctx context.Context, householdID uuid.UUID, start, end time.Time, txType string) (float64, error) {
	var total float64
	err := dbOrTx(ctx, r.db).WithContext(ctx).
		Table("transactions").
		Select("COALESCE(SUM(amount), 0)").
		Where(`household_id = ? AND type = ? AND deleted_at IS NULL
			AND transaction_date >= ? AND transaction_date < ?`, householdID, txType, start, end).
		Scan(&total).Error
	return total, err
}
