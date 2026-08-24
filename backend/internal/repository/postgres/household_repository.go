package postgres

import (
	"context"
	"errors"

	"family-finance-api/internal/entity"
	"family-finance-api/internal/pkg/apperror"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type householdRepository struct {
	db *gorm.DB
}

// NewHouseholdRepository mengembalikan implementasi HouseholdRepository berbasis GORM/Postgres.
func NewHouseholdRepository(db *gorm.DB) *householdRepository {
	return &householdRepository{db: db}
}

func (r *householdRepository) Create(ctx context.Context, household *entity.Household) error {
	return r.db.WithContext(ctx).Create(household).Error
}

func (r *householdRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.Household, error) {
	var h entity.Household
	err := r.db.WithContext(ctx).First(&h, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperror.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &h, nil
}

func (r *householdRepository) CreateMember(ctx context.Context, member *entity.HouseholdMember) error {
	return r.db.WithContext(ctx).Create(member).Error
}

func (r *householdRepository) FindMemberByUserID(ctx context.Context, userID uuid.UUID) (*entity.HouseholdMember, error) {
	var m entity.HouseholdMember
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperror.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *householdRepository) FindMembersByHouseholdID(ctx context.Context, householdID uuid.UUID) ([]entity.HouseholdMember, error) {
	var members []entity.HouseholdMember
	err := r.db.WithContext(ctx).Preload("User").Where("household_id = ?", householdID).Find(&members).Error
	return members, err
}

func (r *householdRepository) CreateInvitation(ctx context.Context, invitation *entity.HouseholdInvitation) error {
	return r.db.WithContext(ctx).Create(invitation).Error
}

func (r *householdRepository) FindInvitationByCode(ctx context.Context, code string) (*entity.HouseholdInvitation, error) {
	var inv entity.HouseholdInvitation
	err := r.db.WithContext(ctx).Where("code = ?", code).First(&inv).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperror.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

func (r *householdRepository) UpdateInvitation(ctx context.Context, invitation *entity.HouseholdInvitation) error {
	return r.db.WithContext(ctx).Save(invitation).Error
}

func (r *householdRepository) Update(ctx context.Context, household *entity.Household) error {
	return r.db.WithContext(ctx).Save(household).Error
}

func (r *householdRepository) DeleteMember(ctx context.Context, memberID uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.HouseholdMember{}, "id = ?", memberID).Error
}

func (r *householdRepository) FindActiveInvitationByHouseholdID(ctx context.Context, householdID uuid.UUID) (*entity.HouseholdInvitation, error) {
	var inv entity.HouseholdInvitation
	err := r.db.WithContext(ctx).Where("household_id = ? AND status = ?", householdID, "active").First(&inv).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &inv, nil
}
