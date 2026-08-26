package repository

import (
	"context"

	"github.com/google/uuid"
	"homeapp/internal/entity"
)

type CategoryRepository interface {
	Create(ctx context.Context, category *entity.Category) error
	FindByID(ctx context.Context, id, householdID uuid.UUID) (*entity.Category, error)
	ListByHousehold(ctx context.Context, householdID uuid.UUID, categoryType string, includeArchived bool) ([]*entity.Category, error)
	Update(ctx context.Context, category *entity.Category) error
	Archive(ctx context.Context, id, householdID uuid.UUID) error
	Unarchive(ctx context.Context, id, householdID uuid.UUID) error
}
