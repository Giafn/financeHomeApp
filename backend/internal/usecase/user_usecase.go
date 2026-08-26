package usecase

import (
	"context"
	"errors"

	"homeapp/internal/pkg/apperror"
	"homeapp/internal/pkg/hash"
	"homeapp/internal/repository"

	"github.com/google/uuid"
)

type UserUsecase struct {
	userRepo      repository.UserRepository
	householdRepo repository.HouseholdRepository
}

func NewUserUsecase(userRepo repository.UserRepository, householdRepo repository.HouseholdRepository) *UserUsecase {
	return &UserUsecase{
		userRepo:      userRepo,
		householdRepo: householdRepo,
	}
}

type UserProfile struct {
	ID         uuid.UUID      `json:"id"`
	Name       string         `json:"name"`
	Email      string         `json:"email"`
	AvatarURL  *string        `json:"avatar_url"`
	Household  *HouseholdInfo `json:"household"`
}

type HouseholdInfo struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
	Role string    `json:"role"`
}

func (u *UserUsecase) GetProfile(ctx context.Context, userID uuid.UUID) (*UserProfile, error) {
	user, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	profile := &UserProfile{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		AvatarURL: user.AvatarURL,
	}

	member, err := u.householdRepo.FindMemberByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			profile.Household = nil
			return profile, nil
		}
		return nil, err
	}

	household, err := u.householdRepo.FindByID(ctx, member.HouseholdID)
	if err != nil {
		return nil, err
	}

	profile.Household = &HouseholdInfo{
		ID:   household.ID,
		Name: household.Name,
		Role: string(member.Role),
	}

	return profile, nil
}

type UpdateProfileInput struct {
	Name      *string
	AvatarURL *string
}

func (u *UserUsecase) UpdateProfile(ctx context.Context, userID uuid.UUID, in UpdateProfileInput) (*UserProfile, error) {
	user, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if in.Name != nil {
		user.Name = *in.Name
	}
	if in.AvatarURL != nil {
		user.AvatarURL = in.AvatarURL
	}

	if err := u.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	profile := &UserProfile{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		AvatarURL: user.AvatarURL,
	}

	member, err := u.householdRepo.FindMemberByUserID(ctx, userID)
	if err == nil {
		household, err := u.householdRepo.FindByID(ctx, member.HouseholdID)
		if err == nil {
			profile.Household = &HouseholdInfo{
				ID:   household.ID,
				Name: household.Name,
				Role: string(member.Role),
			}
		}
	}

	return profile, nil
}

type ChangePasswordInput struct {
	OldPassword string
	NewPassword string
}

func (u *UserUsecase) ChangePassword(ctx context.Context, userID uuid.UUID, in ChangePasswordInput) error {
	user, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}

	if user.PasswordHash == nil || !hash.CheckPassword(in.OldPassword, *user.PasswordHash) {
		return apperror.ErrInvalidCredential
	}

	hashed, err := hash.HashPassword(in.NewPassword)
	if err != nil {
		return err
	}

	user.PasswordHash = &hashed
	return u.userRepo.Update(ctx, user)
}
