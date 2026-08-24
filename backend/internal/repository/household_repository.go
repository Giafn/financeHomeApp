package repository

import (
	"context"

	"family-finance-api/internal/entity"

	"github.com/google/uuid"
)

// HouseholdRepository adalah kontrak akses data Household, HouseholdMember, dan HouseholdInvitation.
type HouseholdRepository interface {
	Create(ctx context.Context, household *entity.Household) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Household, error)
	Update(ctx context.Context, household *entity.Household) error

	CreateMember(ctx context.Context, member *entity.HouseholdMember) error
	FindMemberByUserID(ctx context.Context, userID uuid.UUID) (*entity.HouseholdMember, error)
	FindMembersByHouseholdID(ctx context.Context, householdID uuid.UUID) ([]entity.HouseholdMember, error)
	DeleteMember(ctx context.Context, memberID uuid.UUID) error

	CreateInvitation(ctx context.Context, invitation *entity.HouseholdInvitation) error
	FindInvitationByCode(ctx context.Context, code string) (*entity.HouseholdInvitation, error)
	FindActiveInvitationByHouseholdID(ctx context.Context, householdID uuid.UUID) (*entity.HouseholdInvitation, error)
	UpdateInvitation(ctx context.Context, invitation *entity.HouseholdInvitation) error
}
