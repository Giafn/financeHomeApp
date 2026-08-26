package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"homeapp/internal/entity"
	"homeapp/internal/pkg/apperror"
	"homeapp/internal/repository"
)

const dateLayout = "2006-01-02"

// GoalRecalculator dipanggil setiap kali transaksi ber-goal_id berubah (create/update/delete)
// supaya progress goal (Phase 09) selalu sinkron dengan transaksi yang benar-benar ada.
// Interface sempit di sini (bukan import usecase.GoalUsecase langsung) supaya tidak ada
// dependency dua arah antara transaction & goal usecase.
type GoalRecalculator interface {
	RecalculateStatus(ctx context.Context, goalID uuid.UUID) error
}

type TransactionUsecase struct {
	transactionRepo repository.TransactionRepository
	accountRepo     repository.AccountRepository
	categoryRepo    repository.CategoryRepository
	householdRepo   repository.HouseholdRepository
	goalRecalc      GoalRecalculator
}

func NewTransactionUsecase(
	transactionRepo repository.TransactionRepository,
	accountRepo repository.AccountRepository,
	categoryRepo repository.CategoryRepository,
	householdRepo repository.HouseholdRepository,
	goalRecalc GoalRecalculator,
) *TransactionUsecase {
	return &TransactionUsecase{
		transactionRepo: transactionRepo,
		accountRepo:     accountRepo,
		categoryRepo:    categoryRepo,
		householdRepo:   householdRepo,
		goalRecalc:      goalRecalc,
	}
}

// recalcGoal memanggil goalRecalc kalau goalID tidak nil — no-op aman dipanggil dengan nil
// (transaksi tanpa goal_id) tanpa perlu nil-check berulang di tiap call site.
func (u *TransactionUsecase) recalcGoal(ctx context.Context, goalID *uuid.UUID) error {
	if goalID == nil || u.goalRecalc == nil {
		return nil
	}
	return u.goalRecalc.RecalculateStatus(ctx, *goalID)
}

type TransactionInput struct {
	Type                 string
	AccountID            uuid.UUID
	DestinationAccountID *uuid.UUID
	CategoryID           *uuid.UUID
	Amount               float64
	Description          *string
	TransactionDate      string
	AttachmentURL        *string
	GoalID               *uuid.UUID
	// BillPeriodID diisi internal oleh BillUsecase (Phase 10) lewat POST /bill-periods/{id}/pay —
	// tidak diexpose lewat DTO transaksi umum, beda dengan GoalID yang memang diisi client.
	BillPeriodID *uuid.UUID
}

// validate menegakkan aturan bisnis inti Phase 06 §4: kategori wajib+cocok tipe untuk
// income/expense, akun tujuan wajib+beda untuk transfer, dan semua akun harus milik
// household ini serta aktif.
func (u *TransactionUsecase) validate(ctx context.Context, householdID uuid.UUID, input TransactionInput) error {
	account, err := u.accountRepo.FindByID(ctx, input.AccountID)
	if err != nil {
		return err
	}
	if account.HouseholdID != householdID {
		return apperror.ErrNotFound
	}
	if !account.IsActive {
		return apperror.ErrAccountInactive
	}

	switch entity.TransactionType(input.Type) {
	case entity.TransactionIncome, entity.TransactionExpense:
		if input.CategoryID == nil {
			return apperror.ErrCategoryRequired
		}
		category, err := u.categoryRepo.FindByID(ctx, *input.CategoryID, householdID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return apperror.ErrNotFound
			}
			return err
		}
		if string(category.Type) != input.Type {
			return apperror.ErrCategoryTypeMismatch
		}
	case entity.TransactionTransfer:
		if input.DestinationAccountID == nil {
			return apperror.ErrDestinationRequired
		}
		if *input.DestinationAccountID == input.AccountID {
			return apperror.ErrTransferSameAccount
		}
		destAccount, err := u.accountRepo.FindByID(ctx, *input.DestinationAccountID)
		if err != nil {
			return err
		}
		if destAccount.HouseholdID != householdID {
			return apperror.ErrNotFound
		}
		if !destAccount.IsActive {
			return apperror.ErrAccountInactive
		}
	}

	return nil
}

func (u *TransactionUsecase) CreateTransaction(ctx context.Context, userID uuid.UUID, input TransactionInput) (*repository.TransactionListItem, error) {
	member, err := u.householdRepo.FindMemberByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if err := u.validate(ctx, member.HouseholdID, input); err != nil {
		return nil, err
	}

	txDate, err := time.Parse(dateLayout, input.TransactionDate)
	if err != nil {
		return nil, err
	}

	transaction := &entity.Transaction{
		HouseholdID:          member.HouseholdID,
		Type:                 entity.TransactionType(input.Type),
		AccountID:            input.AccountID,
		DestinationAccountID: input.DestinationAccountID,
		CategoryID:           input.CategoryID,
		Amount:               input.Amount,
		Description:          input.Description,
		TransactionDate:      txDate,
		AttachmentURL:        input.AttachmentURL,
		GoalID:               input.GoalID,
		BillPeriodID:         input.BillPeriodID,
		CreatedBy:            userID,
	}

	if err := u.transactionRepo.Create(ctx, transaction); err != nil {
		return nil, err
	}

	if err := u.recalcGoal(ctx, input.GoalID); err != nil {
		return nil, err
	}

	return u.transactionRepo.FindByID(ctx, transaction.ID, member.HouseholdID)
}

func (u *TransactionUsecase) GetTransaction(ctx context.Context, userID, transactionID uuid.UUID) (*repository.TransactionListItem, error) {
	member, err := u.householdRepo.FindMemberByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return u.transactionRepo.FindByID(ctx, transactionID, member.HouseholdID)
}

type ListTransactionsInput struct {
	Type       string
	AccountID  *uuid.UUID
	CategoryID *uuid.UUID
	CreatedBy  *uuid.UUID
	DateFrom   *time.Time
	DateTo     *time.Time
	Page       int
	Limit      int
}

func (u *TransactionUsecase) ListTransactions(ctx context.Context, userID uuid.UUID, input ListTransactionsInput) ([]*repository.TransactionListItem, int64, error) {
	member, err := u.householdRepo.FindMemberByUserID(ctx, userID)
	if err != nil {
		return nil, 0, err
	}

	filter := repository.TransactionFilter{
		Type:       input.Type,
		AccountID:  input.AccountID,
		CategoryID: input.CategoryID,
		CreatedBy:  input.CreatedBy,
		DateFrom:   input.DateFrom,
		DateTo:     input.DateTo,
		Page:       input.Page,
		Limit:      input.Limit,
	}

	return u.transactionRepo.List(ctx, member.HouseholdID, filter)
}

func (u *TransactionUsecase) UpdateTransaction(ctx context.Context, userID, transactionID uuid.UUID, input TransactionInput) (*repository.TransactionListItem, error) {
	member, err := u.householdRepo.FindMemberByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	existing, err := u.transactionRepo.FindByID(ctx, transactionID, member.HouseholdID)
	if err != nil {
		return nil, err
	}

	if err := u.validate(ctx, member.HouseholdID, input); err != nil {
		return nil, err
	}

	txDate, err := time.Parse(dateLayout, input.TransactionDate)
	if err != nil {
		return nil, err
	}

	oldGoalID := existing.GoalID

	updated := &entity.Transaction{
		BaseModel:            existing.BaseModel,
		HouseholdID:          existing.HouseholdID, // immutable
		Type:                 entity.TransactionType(input.Type),
		AccountID:            input.AccountID,
		DestinationAccountID: input.DestinationAccountID,
		CategoryID:           input.CategoryID,
		Amount:               input.Amount,
		Description:          input.Description,
		TransactionDate:      txDate,
		AttachmentURL:        input.AttachmentURL,
		GoalID:               input.GoalID,
		CreatedBy:            existing.CreatedBy, // immutable
	}

	if err := u.transactionRepo.Update(ctx, updated); err != nil {
		return nil, err
	}

	// Recalc goal lama (kalau goal_id dilepas/diganti) dan goal baru (kalau ditambah/diganti) —
	// keduanya bisa berbeda dan keduanya butuh sinkronisasi ulang.
	if err := u.recalcGoal(ctx, oldGoalID); err != nil {
		return nil, err
	}
	if input.GoalID != nil && (oldGoalID == nil || *oldGoalID != *input.GoalID) {
		if err := u.recalcGoal(ctx, input.GoalID); err != nil {
			return nil, err
		}
	}

	return u.transactionRepo.FindByID(ctx, transactionID, member.HouseholdID)
}

func (u *TransactionUsecase) DeleteTransaction(ctx context.Context, userID, transactionID uuid.UUID) error {
	member, err := u.householdRepo.FindMemberByUserID(ctx, userID)
	if err != nil {
		return err
	}

	existing, err := u.transactionRepo.FindByID(ctx, transactionID, member.HouseholdID)
	if err != nil {
		return err
	}

	linkedToPaidBill, err := u.transactionRepo.IsLinkedToPaidBillPeriod(ctx, transactionID)
	if err != nil {
		return err
	}
	if linkedToPaidBill {
		return apperror.ErrBillPeriodPaidConflict
	}

	if err := u.transactionRepo.Delete(ctx, transactionID, member.HouseholdID); err != nil {
		return err
	}

	return u.recalcGoal(ctx, existing.GoalID)
}

func (u *TransactionUsecase) GetQuickSelect(ctx context.Context, userID uuid.UUID) (*uuid.UUID, *uuid.UUID, error) {
	member, err := u.householdRepo.FindMemberByUserID(ctx, userID)
	if err != nil {
		return nil, nil, err
	}

	return u.transactionRepo.LastUsedAccountAndCategory(ctx, member.HouseholdID, userID)
}
