package usecase

import (
	"context"
	"crypto/rand"
	"errors"
	"math/big"
	"time"

	"family-finance-api/internal/entity"
	"family-finance-api/internal/pkg/apperror"
	"family-finance-api/internal/repository"

	"github.com/google/uuid"
)

const invitationExpiryDays = 7
const invitationCodeLength = 8
const invitationCodeCharset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // tanpa karakter ambigu: 0/O, 1/I

type HouseholdUsecase struct {
	householdRepo repository.HouseholdRepository
}

func NewHouseholdUsecase(householdRepo repository.HouseholdRepository) *HouseholdUsecase {
	return &HouseholdUsecase{householdRepo: householdRepo}
}

// CreateHousehold membuat rumah tangga baru dan menjadikan user sebagai owner.
func (u *HouseholdUsecase) CreateHousehold(ctx context.Context, userID uuid.UUID, name string) (*entity.Household, error) {
	if _, err := u.householdRepo.FindMemberByUserID(ctx, userID); err == nil {
		return nil, apperror.ErrAlreadyInHousehold
	} else if !errors.Is(err, apperror.ErrNotFound) {
		return nil, err
	}

	household := &entity.Household{Name: name, CreatedBy: userID}
	if err := u.householdRepo.Create(ctx, household); err != nil {
		return nil, err
	}

	member := &entity.HouseholdMember{
		HouseholdID: household.ID,
		UserID:      userID,
		Role:        entity.RoleOwner,
		JoinedAt:    time.Now(),
	}
	if err := u.householdRepo.CreateMember(ctx, member); err != nil {
		return nil, err
	}

	return household, nil
}

// JoinHousehold memvalidasi kode undangan lalu menambahkan user sebagai member.
func (u *HouseholdUsecase) JoinHousehold(ctx context.Context, userID uuid.UUID, code string) (*entity.Household, error) {
	if _, err := u.householdRepo.FindMemberByUserID(ctx, userID); err == nil {
		return nil, apperror.ErrAlreadyInHousehold
	} else if !errors.Is(err, apperror.ErrNotFound) {
		return nil, err
	}

	invitation, err := u.householdRepo.FindInvitationByCode(ctx, code)
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			return nil, apperror.ErrInvitationInvalid
		}
		return nil, err
	}

	if invitation.Status != entity.InvitationActive || time.Now().After(invitation.ExpiresAt) {
		return nil, apperror.ErrInvitationInvalid
	}

	member := &entity.HouseholdMember{
		HouseholdID: invitation.HouseholdID,
		UserID:      userID,
		Role:        entity.RoleMember,
		JoinedAt:    time.Now(),
	}
	if err := u.householdRepo.CreateMember(ctx, member); err != nil {
		return nil, err
	}

	now := time.Now()
	invitation.Status = entity.InvitationUsed
	invitation.UsedBy = &userID
	invitation.UsedAt = &now
	if err := u.householdRepo.UpdateInvitation(ctx, invitation); err != nil {
		return nil, err
	}

	return u.householdRepo.FindByID(ctx, invitation.HouseholdID)
}

// CreateInvitation men-generate kode undangan baru. Hanya owner yang boleh melakukan ini.
// Otomatis expire kode aktif lama jika ada.
func (u *HouseholdUsecase) CreateInvitation(ctx context.Context, userID uuid.UUID) (*entity.HouseholdInvitation, error) {
	member, err := u.householdRepo.FindMemberByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if member.Role != entity.RoleOwner {
		return nil, apperror.ErrForbidden
	}

	// Auto-expire kode aktif lama sebelum membuat kode baru
	if oldInv, err := u.householdRepo.FindActiveInvitationByHouseholdID(ctx, member.HouseholdID); err == nil && oldInv != nil {
		oldInv.Status = entity.InvitationExpired
		_ = u.householdRepo.UpdateInvitation(ctx, oldInv)
	}

	code, err := generateInvitationCode()
	if err != nil {
		return nil, err
	}

	invitation := &entity.HouseholdInvitation{
		HouseholdID: member.HouseholdID,
		Code:        code,
		CreatedBy:   userID,
		ExpiresAt:   time.Now().AddDate(0, 0, invitationExpiryDays),
		Status:      entity.InvitationActive,
	}
	if err := u.householdRepo.CreateInvitation(ctx, invitation); err != nil {
		return nil, err
	}

	return invitation, nil
}

// GetHouseholdDetail mengembalikan detail household dengan role dan member count.
func (u *HouseholdUsecase) GetHouseholdDetail(ctx context.Context, userID uuid.UUID) (map[string]interface{}, error) {
	member, err := u.householdRepo.FindMemberByUserID(ctx, userID)
	if err != nil {
		return nil, apperror.ErrNotFound
	}

	household, err := u.householdRepo.FindByID(ctx, member.HouseholdID)
	if err != nil {
		return nil, err
	}

	members, err := u.householdRepo.FindMembersByHouseholdID(ctx, member.HouseholdID)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"id":            household.ID,
		"name":          household.Name,
		"role":          member.Role,
		"member_count":  len(members),
	}, nil
}

// UpdateHouseholdName update nama household. Hanya owner yang boleh.
func (u *HouseholdUsecase) UpdateHouseholdName(ctx context.Context, userID uuid.UUID, householdID uuid.UUID, newName string) error {
	member, err := u.householdRepo.FindMemberByUserID(ctx, userID)
	if err != nil {
		return apperror.ErrNotFound
	}
	if member.HouseholdID != householdID {
		return apperror.ErrForbidden
	}
	if member.Role != entity.RoleOwner {
		return apperror.ErrForbidden
	}

	household, err := u.householdRepo.FindByID(ctx, householdID)
	if err != nil {
		return err
	}

	household.Name = newName
	return u.householdRepo.Update(ctx, household)
}

// GetMembers mengembalikan list anggota dengan detail user.
func (u *HouseholdUsecase) GetMembers(ctx context.Context, userID uuid.UUID) ([]map[string]interface{}, error) {
	member, err := u.householdRepo.FindMemberByUserID(ctx, userID)
	if err != nil {
		return nil, apperror.ErrNotFound
	}

	members, err := u.householdRepo.FindMembersByHouseholdID(ctx, member.HouseholdID)
	if err != nil {
		return nil, err
	}

	var result []map[string]interface{}
	for _, m := range members {
		result = append(result, map[string]interface{}{
			"user_id":   m.UserID,
			"name":      m.User.Name,
			"email":     m.User.Email,
			"avatar_url": m.User.AvatarURL,
			"role":      m.Role,
			"joined_at": m.JoinedAt,
		})
	}

	return result, nil
}

// RemoveMember mengeluarkan anggota dari household. Hanya owner yang boleh.
// Tidak boleh mengeluarkan sole owner.
func (u *HouseholdUsecase) RemoveMember(ctx context.Context, userID uuid.UUID, targetUserID uuid.UUID) error {
	member, err := u.householdRepo.FindMemberByUserID(ctx, userID)
	if err != nil {
		return apperror.ErrNotFound
	}
	if member.Role != entity.RoleOwner {
		return apperror.ErrForbidden
	}

	targetMember, err := u.householdRepo.FindMemberByUserID(ctx, targetUserID)
	if err != nil {
		return apperror.ErrNotFound
	}
	if targetMember.HouseholdID != member.HouseholdID {
		return apperror.ErrForbidden
	}

	// Cegah mengeluarkan sole owner
	if targetMember.Role == entity.RoleOwner {
		members, _ := u.householdRepo.FindMembersByHouseholdID(ctx, member.HouseholdID)
		ownerCount := 0
		for _, m := range members {
			if m.Role == entity.RoleOwner {
				ownerCount++
			}
		}
		if ownerCount <= 1 {
			return apperror.ErrCannotRemoveSoleOwner
		}
	}

	return u.householdRepo.DeleteMember(ctx, targetMember.ID)
}

// GetActiveInvitation mengembalikan kode undangan aktif jika ada.
func (u *HouseholdUsecase) GetActiveInvitation(ctx context.Context, userID uuid.UUID) (*entity.HouseholdInvitation, error) {
	member, err := u.householdRepo.FindMemberByUserID(ctx, userID)
	if err != nil {
		return nil, apperror.ErrNotFound
	}

	return u.householdRepo.FindActiveInvitationByHouseholdID(ctx, member.HouseholdID)
}

func generateInvitationCode() (string, error) {
	result := make([]byte, invitationCodeLength)
	for i := range result {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(invitationCodeCharset))))
		if err != nil {
			return "", err
		}
		result[i] = invitationCodeCharset[n.Int64()]
	}
	return string(result), nil
}
