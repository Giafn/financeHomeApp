package usecase

import (
	"context"
	"errors"

	"homeapp/internal/entity"
	"homeapp/internal/pkg/apperror"
	"homeapp/internal/repository"

	"github.com/google/uuid"
)

const (
	AccountTypeBank    = "bank"
	AccountTypeEwallet = "ewallet"
	AccountTypeCash    = "cash"
	AccountTypeOther   = "other"
)

type AccountUsecase struct {
	householdRepo repository.HouseholdRepository
	accountRepo   repository.AccountRepository
}

func NewAccountUsecase(householdRepo repository.HouseholdRepository, accountRepo repository.AccountRepository) *AccountUsecase {
	return &AccountUsecase{householdRepo: householdRepo, accountRepo: accountRepo}
}

func (u *AccountUsecase) CreateAccount(ctx context.Context, userID uuid.UUID, name string, accountType string, initialBalance float64, ownerType string) (*entity.Account, error) {
	member, err := u.householdRepo.FindMemberByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if ownerType == "" {
		ownerType = string(entity.AccountOwnerHousehold)
	}

	if entity.AccountOwnerType(ownerType) == entity.AccountOwnerHousehold && member.Role != entity.RoleOwner {
		return nil, apperror.ErrForbidden
	}

	account := &entity.Account{
		HouseholdID:    member.HouseholdID,
		Name:           name,
		Type:           entity.AccountType(accountType),
		InitialBalance: initialBalance,
		IsActive:       true,
		OwnerType:      entity.AccountOwnerType(ownerType),
		CreatedBy:      userID,
	}
	if account.OwnerType == entity.AccountOwnerPersonal {
		account.OwnerUserID = &userID
	}

	if err := u.accountRepo.Create(ctx, account); err != nil {
		return nil, err
	}

	return account, nil
}

// GetAccount mengembalikan detail akun. Verifikasi household ownership.
func (u *AccountUsecase) GetAccount(ctx context.Context, userID uuid.UUID, accountID uuid.UUID) (*entity.Account, error) {
	member, err := u.householdRepo.FindMemberByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	account, err := u.accountRepo.FindByID(ctx, accountID)
	if err != nil {
		return nil, err
	}

	if account.HouseholdID != member.HouseholdID {
		return nil, apperror.ErrForbidden
	}

	return account, nil
}

// ListAccounts mengembalikan list akun household user. Tambahkan current_balance ke response.
func (u *AccountUsecase) ListAccounts(ctx context.Context, userID uuid.UUID, includeInactive bool) ([]map[string]interface{}, error) {
	member, err := u.householdRepo.FindMemberByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	accounts, err := u.accountRepo.ListByHouseholdID(ctx, member.HouseholdID, includeInactive)
	if err != nil {
		return nil, err
	}

	var result []map[string]interface{}
	for _, acc := range accounts {
		// Hitung current_balance (sementara = initial_balance)
		balance, _ := u.accountRepo.CalculateBalance(ctx, acc.ID)

		isOwnedByMe := acc.OwnerType != entity.AccountOwnerPersonal || (acc.OwnerUserID != nil && *acc.OwnerUserID == userID)

		result = append(result, map[string]interface{}{
			"id":              acc.ID,
			"name":            acc.Name,
			"type":            acc.Type,
			"initial_balance": acc.InitialBalance,
			"current_balance": balance,
			"is_active":       acc.IsActive,
			"owner_type":      acc.OwnerType,
			"owner_user_id":   acc.OwnerUserID,
			"is_owned_by_me":  isOwnedByMe,
		})
	}

	return result, nil
}

// UpdateAccount update nama, type, atau is_active. Tidak boleh ubah initial_balance.
func (u *AccountUsecase) UpdateAccount(ctx context.Context, userID uuid.UUID, accountID uuid.UUID, updates map[string]interface{}) error {
	member, err := u.householdRepo.FindMemberByUserID(ctx, userID)
	if err != nil {
		return err
	}

	account, err := u.accountRepo.FindByID(ctx, accountID)
	if err != nil {
		return err
	}

	if account.HouseholdID != member.HouseholdID {
		return apperror.ErrForbidden
	}

	if account.OwnerType == entity.AccountOwnerPersonal {
		if account.OwnerUserID == nil || *account.OwnerUserID != userID {
			return apperror.ErrPersonalAccountForbidden
		}
	} else if member.Role != entity.RoleOwner {
		return apperror.ErrForbidden
	}

	// Update hanya field yang diizinkan
	if name, ok := updates["name"].(string); ok && name != "" {
		account.Name = name
	}
	if accType, ok := updates["type"].(string); ok && accType != "" {
		account.Type = entity.AccountType(accType)
	}
	if isActive, ok := updates["is_active"].(bool); ok {
		account.IsActive = isActive
	}
	if ownerType, ok := updates["owner_type"].(string); ok && ownerType != "" {
		account.OwnerType = entity.AccountOwnerType(ownerType)
		if account.OwnerType == entity.AccountOwnerPersonal {
			account.OwnerUserID = &userID
		} else {
			account.OwnerUserID = nil
		}
	}

	return u.accountRepo.Update(ctx, account)
}

// GetAccountBalance mengembalikan current_balance akun untuk dipakai di phase lain.
func (u *AccountUsecase) GetAccountBalance(ctx context.Context, userID uuid.UUID, accountID uuid.UUID) (float64, error) {
	// Verifikasi akun milik household user
	_, err := u.GetAccount(ctx, userID, accountID)
	if err != nil {
		if !errors.Is(err, apperror.ErrForbidden) {
			return 0, apperror.ErrNotFound
		}
		return 0, err
	}

	return u.accountRepo.CalculateBalance(ctx, accountID)
}
