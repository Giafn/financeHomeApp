package repository

import (
	"context"

	"github.com/google/uuid"
	"homeapp/internal/entity"
)

type GoalRepository interface {
	Create(ctx context.Context, goal *entity.Goal) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Goal, error)
	FindByIDAndHousehold(ctx context.Context, id, householdID uuid.UUID) (*entity.Goal, error)
	ListByHouseholdAndStatus(ctx context.Context, householdID uuid.UUID, status string) ([]*entity.Goal, error)
	Update(ctx context.Context, goal *entity.Goal) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status entity.GoalStatus) error
	Delete(ctx context.Context, id, householdID uuid.UUID) error
	// SumTransferAmount jumlah transfer ber-goal_id ini yang account_id/destination_account_id-nya
	// cocok linkedAccountID — asIncoming=true untuk kontribusi (destination), false untuk penarikan (source).
	SumTransferAmount(ctx context.Context, goalID, linkedAccountID uuid.UUID, asIncoming bool) (float64, error)
	// HasContributions dipakai validasi delete — goal dengan histori transaksi tidak boleh dihapus begitu saja.
	HasContributions(ctx context.Context, goalID uuid.UUID) (bool, error)
}
