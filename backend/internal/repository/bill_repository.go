package repository

import (
	"context"

	"github.com/google/uuid"
	"homeapp/internal/entity"
)

type BillRepository interface {
	Create(ctx context.Context, bill *entity.Bill) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Bill, error)
	FindByIDAndHousehold(ctx context.Context, id, householdID uuid.UUID) (*entity.Bill, error)
	ListByHousehold(ctx context.Context, householdID uuid.UUID, isActive *bool) ([]*entity.Bill, error)
	// ListIndefiniteActive bill aktif tanpa end_period — dipakai job bill-period-generator (Phase 10 §2).
	ListIndefiniteActive(ctx context.Context) ([]*entity.Bill, error)
	Update(ctx context.Context, bill *entity.Bill) error
}
