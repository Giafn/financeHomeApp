package repository

import (
	"context"

	"github.com/google/uuid"
	"homeapp/internal/entity"
)

// BudgetWithSpent gabungan budget + nama kategori + agregasi expense bulan berjalan.
type BudgetWithSpent struct {
	entity.Budget
	CategoryName string  `json:"category_name"`
	Spent        float64 `json:"spent"`
}

type BudgetRepository interface {
	Create(ctx context.Context, budget *entity.Budget) error
	FindByID(ctx context.Context, id, householdID uuid.UUID) (*entity.Budget, error)
	FindByCategoryPeriod(ctx context.Context, householdID, categoryID uuid.UUID, period string) (*entity.Budget, error)
	// ListByPeriod — cycleStartDay dari household menentukan window transaksi yang dihitung
	// sebagai "spent" (lihat internal/pkg/period). cycleStartDay<=1 = window kalender biasa.
	ListByPeriod(ctx context.Context, householdID uuid.UUID, period string, cycleStartDay int) ([]*BudgetWithSpent, error)
	// ListRawByPeriod dipakai job auto-copy — baris budget mentah tanpa agregasi spent.
	ListRawByPeriod(ctx context.Context, householdID uuid.UUID, period string) ([]*entity.Budget, error)
	Update(ctx context.Context, budget *entity.Budget) error
	Delete(ctx context.Context, id, householdID uuid.UUID) error
}
