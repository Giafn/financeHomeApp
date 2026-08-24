package postgres

import (
	"context"
	"errors"

	"family-finance-api/internal/entity"
	"family-finance-api/internal/pkg/apperror"

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

// CalculateBalance menghitung current_balance = initial_balance + SUM(transaksi terkait).
// Untuk Phase 04, belum ada transaksi, jadi hasilnya = initial_balance.
// Implementasi penuh di Phase 06 saat ada tabel transaction.
func (r *accountRepository) CalculateBalance(ctx context.Context, accountID uuid.UUID) (float64, error) {
	var acc entity.Account
	err := r.db.WithContext(ctx).First(&acc, "id = ?", accountID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, apperror.ErrNotFound
	}
	if err != nil {
		return 0, err
	}

	// Sementara Phase 04: return initial_balance (transaksi belum ada).
	// TODO Phase 06: tambah SUM(transaksi) ke formula ini
	return acc.InitialBalance, nil
}
