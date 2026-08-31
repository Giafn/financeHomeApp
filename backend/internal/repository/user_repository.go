package repository

import (
	"context"

	"homeapp/internal/entity"

	"github.com/google/uuid"
)

// UserRepository adalah kontrak akses data User.
// Layer usecase HANYA bergantung pada interface ini, tidak tahu detail implementasi (GORM/Postgres).
type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	FindByEmail(ctx context.Context, email string) (*entity.User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	FindByVerificationToken(ctx context.Context, token string) (*entity.User, error)
	Update(ctx context.Context, user *entity.User) error
	Delete(ctx context.Context, id uuid.UUID) error
}
