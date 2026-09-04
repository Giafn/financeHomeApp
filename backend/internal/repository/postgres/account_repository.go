package postgres

import (
	"context"
	"errors"

	"homeapp/internal/entity"
	"homeapp/internal/pkg/apperror"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type accountRepository struct {
	db *gorm.DB
}

// NewAccountRepository mengembalikan implementasi AccountRepository berbasis GORM/Postgres.
func NewAccountRepository(db *gorm.DB) *accountRepository {
	return &accountRepository{db: db}
}

func (r *accountRepository) Create(ctx context.Context, account *entity.Account) error {
	return r.db.WithContext(ctx).Create(account).Error
}

func (r *accountRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.Account, error) {
	var a entity.Account
	err := r.db.WithContext(ctx).First(&a, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperror.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *accountRepository) ListByHouseholdID(ctx context.Context, householdID uuid.UUID, includeInactive bool) ([]entity.Account, error) {
	var accounts []entity.Account
	query := r.db.WithContext(ctx).Where("household_id = ?", householdID)
	if !includeInactive {
		query = query.Where("is_active = ?", true)
	}
	err := query.Find(&accounts).Error
	return accounts, err
}

func (r *accountRepository) Update(ctx context.Context, account *entity.Account) error {
	return r.db.WithContext(ctx).Save(account).Error
}

// CalculateBalance menghitung current_balance = initial_balance
//   + SUM(income masuk ke akun ini)
//   - SUM(expense keluar dari akun ini)
//   - SUM(transfer keluar dari akun ini)
//   + SUM(transfer masuk ke akun ini sebagai destination)
func (r *accountRepository) CalculateBalance(ctx context.Context, accountID uuid.UUID) (float64, error) {
	var acc entity.Account
	err := r.db.WithContext(ctx).First(&acc, "id = ?", accountID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, apperror.ErrNotFound
	}
	if err != nil {
		return 0, err
	}

	var delta float64
	err = r.db.WithContext(ctx).Raw(`
		SELECT
			COALESCE(SUM(CASE WHEN type = 'income' AND account_id = ? THEN amount ELSE 0 END), 0)
			- COALESCE(SUM(CASE WHEN type = 'expense' AND account_id = ? THEN amount ELSE 0 END), 0)
			- COALESCE(SUM(CASE WHEN type = 'transfer' AND account_id = ? THEN amount + admin_fee ELSE 0 END), 0)
			+ COALESCE(SUM(CASE WHEN type = 'transfer' AND destination_account_id = ? THEN amount ELSE 0 END), 0)
		FROM transactions
		WHERE deleted_at IS NULL
		  AND (account_id = ? OR destination_account_id = ?)
	`, accountID, accountID, accountID, accountID, accountID, accountID).Scan(&delta).Error
	if err != nil {
		return 0, err
	}

	return acc.InitialBalance + delta, nil
}
