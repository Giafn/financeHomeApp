package repository

import (
	"context"

	"homeapp/internal/entity"

	"github.com/google/uuid"
)

// AccountRepository adalah kontrak akses data Account.
type AccountRepository interface {
	Create(ctx context.Context, account *entity.Account) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Account, error)
	ListByHouseholdID(ctx context.Context, householdID uuid.UUID, includeInactive bool) ([]entity.Account, error)
	Update(ctx context.Context, account *entity.Account) error
	CalculateBalance(ctx context.Context, accountID uuid.UUID) (float64, error)
}
