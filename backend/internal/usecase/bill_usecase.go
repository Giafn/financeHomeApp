package usecase

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"homeapp/internal/entity"
	"homeapp/internal/pkg/apperror"
	"homeapp/internal/repository"
)

type BillUsecase struct {
	billRepo       repository.BillRepository
	billPeriodRepo repository.BillPeriodRepository
	categoryRepo   repository.CategoryRepository
	accountRepo    repository.AccountRepository
	householdRepo  repository.HouseholdRepository
	transactionUC  *TransactionUsecase
	txManager      repository.TxManager
}

func NewBillUsecase(
	billRepo repository.BillRepository,
	billPeriodRepo repository.BillPeriodRepository,
	categoryRepo repository.CategoryRepository,
	accountRepo repository.AccountRepository,
	householdRepo repository.HouseholdRepository,
	transactionUC *TransactionUsecase,
	txManager repository.TxManager,
) *BillUsecase {
	return &BillUsecase{
		billRepo:       billRepo,
		billPeriodRepo: billPeriodRepo,
		categoryRepo:   categoryRepo,
		accountRepo:    accountRepo,
		householdRepo:  householdRepo,
		transactionUC:  transactionUC,
		txManager:      txManager,
	}
}

// DueDateForPeriod hitung due_date = due_day di bulan `period` ("YYYY-MM"), clamp ke
// tanggal terakhir bulan itu kalau due_day melebihi jumlah hari bulan tsb (misal due_day=31
// di bulan Februari -> jadi tanggal 28/29).
func DueDateForPeriod(period string, dueDay int) (time.Time, error) {
	start, err := time.Parse("2006-01", period)
	if err != nil {
		return time.Time{}, err
	}
	lastDayOfMonth := start.AddDate(0, 1, -1).Day()
	day := dueDay
	if day > lastDayOfMonth {
		day = lastDayOfMonth
	}
	return time.Date(start.Year(), start.Month(), day, 0, 0, 0, 0, time.UTC), nil
}

// monthsBetween daftar period "YYYY-MM" inklusif dari start s.d. end.
func monthsBetween(start, end string) ([]string, error) {
	s, err := time.Parse("2006-01", start)
	if err != nil {
		return nil, err
	}
	e, err := time.Parse("2006-01", end)
	if err != nil {
		return nil, err
	}
	if e.Before(s) {
		return nil, apperror.ErrInvalidPeriodFormat
	}

	var periods []string
	for cur := s; !cur.After(e); cur = cur.AddDate(0, 1, 0) {
		periods = append(periods, cur.Format("2006-01"))
	}
	return periods, nil
}

func (u *BillUsecase) CreateBill(ctx context.Context, userID uuid.UUID, name string, categoryID uuid.UUID, amount float64, dueDay int, startPeriod string, endPeriod *string, reminderDaysBefore int) (*entity.Bill, []*entity.BillPeriod, error) {
	member, err := u.householdRepo.FindMemberByUserID(ctx, userID)
	if err != nil {
		return nil, nil, err
	}

	category, err := u.categoryRepo.FindByID(ctx, categoryID, member.HouseholdID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, apperror.ErrNotFound
		}
		return nil, nil, err
	}
	if category.Type != entity.CategoryExpense {
		return nil, nil, apperror.ErrCategoryNotExpense
	}
	hasChildren, err := u.categoryRepo.HasChildren(ctx, categoryID, member.HouseholdID)
	if err != nil {
		return nil, nil, err
	}
	if hasChildren {
		return nil, nil, apperror.ErrCategoryHasChildren
	}

	if !periodPattern.MatchString(startPeriod) {
		return nil, nil, apperror.ErrInvalidPeriodFormat
	}
	if endPeriod != nil && !periodPattern.MatchString(*endPeriod) {
		return nil, nil, apperror.ErrInvalidPeriodFormat
	}

	bill := &entity.Bill{
		HouseholdID:        member.HouseholdID,
		Name:               name,
		CategoryID:         categoryID,
		Amount:             amount,
		DueDay:             dueDay,
		StartPeriod:        startPeriod,
		EndPeriod:          endPeriod,
		ReminderDaysBefore: reminderDaysBefore,
		IsActive:           true,
		CreatedBy:          userID,
	}

	if err := u.billRepo.Create(ctx, bill); err != nil {
		return nil, nil, err
	}

	// Rentang tetap (poin 1): generate SEMUA periode sekarang, sinkron. Indefinite (poin 2):
	// cuma 1 periode awal, sisanya job bulanan bill-period-generator.
	periodsToGenerate := []string{startPeriod}
	if endPeriod != nil {
		periodsToGenerate, err = monthsBetween(startPeriod, *endPeriod)
		if err != nil {
			return nil, nil, err
		}
	}

	billPeriods := make([]*entity.BillPeriod, 0, len(periodsToGenerate))
	for _, period := range periodsToGenerate {
		dueDate, err := DueDateForPeriod(period, dueDay)
		if err != nil {
			return nil, nil, err
		}
		billPeriods = append(billPeriods, &entity.BillPeriod{
			BillID:  bill.ID,
			Period:  period,
			DueDate: dueDate,
			Status:  entity.BillPeriodUpcoming,
		})
	}

	if err := u.billPeriodRepo.CreateBulk(ctx, billPeriods); err != nil {
		return nil, nil, err
	}

	return bill, billPeriods, nil
}

type BillWithNextPeriod struct {
	*entity.Bill
	NextPeriod *entity.BillPeriod
}

func (u *BillUsecase) ListBills(ctx context.Context, userID uuid.UUID, isActive *bool) ([]*BillWithNextPeriod, error) {
	member, err := u.householdRepo.FindMemberByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	bills, err := u.billRepo.ListByHousehold(ctx, member.HouseholdID, isActive)
	if err != nil {
		return nil, err
	}

	result := make([]*BillWithNextPeriod, 0, len(bills))
	for _, b := range bills {
		next, err := u.billPeriodRepo.NextUpcomingByBillID(ctx, b.ID)
		if err != nil && !errors.Is(err, apperror.ErrNotFound) {
			return nil, err
		}
		if errors.Is(err, apperror.ErrNotFound) {
			next = nil
		}
		result = append(result, &BillWithNextPeriod{Bill: b, NextPeriod: next})
	}
	return result, nil
}

func (u *BillUsecase) GetBillPeriods(ctx context.Context, userID, billID uuid.UUID) ([]*entity.BillPeriod, error) {
	member, err := u.householdRepo.FindMemberByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if _, err := u.billRepo.FindByIDAndHousehold(ctx, billID, member.HouseholdID); err != nil {
		return nil, err
	}

	return u.billPeriodRepo.ListByBillID(ctx, billID)
}

type UpdateBillInput struct {
	IsActive           *bool
	Name               *string
	Amount             *float64
	CategoryID         *uuid.UUID
	ReminderDaysBefore *int
	DueDay             *int
	// EndPeriod: batas akhir bill (dipakai alur "stop tagihan" / menutup bill).
	EndPeriod *string
}

func (u *BillUsecase) UpdateBill(ctx context.Context, userID, billID uuid.UUID, input UpdateBillInput) (*entity.Bill, error) {
	member, err := u.householdRepo.FindMemberByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	bill, err := u.billRepo.FindByIDAndHousehold(ctx, billID, member.HouseholdID)
	if err != nil {
		return nil, err
	}

	if input.IsActive != nil {
		bill.IsActive = *input.IsActive
	}
	if input.Name != nil {
		if strings.TrimSpace(*input.Name) == "" {
			return nil, apperror.ErrInvalidInput
		}
		bill.Name = strings.TrimSpace(*input.Name)
	}
	if input.Amount != nil {
		if *input.Amount <= 0 {
			return nil, apperror.ErrInvalidInput
		}
		bill.Amount = *input.Amount
	}
	if input.CategoryID != nil {
		category, err := u.categoryRepo.FindByID(ctx, *input.CategoryID, member.HouseholdID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, apperror.ErrNotFound
			}
			return nil, err
		}
		if category.Type != entity.CategoryExpense {
			return nil, apperror.ErrCategoryNotExpense
		}
		bill.CategoryID = *input.CategoryID
	}
	if input.ReminderDaysBefore != nil {
		bill.ReminderDaysBefore = *input.ReminderDaysBefore
	}
	if input.DueDay != nil {
		bill.DueDay = *input.DueDay
	}
	if input.EndPeriod != nil {
		if !periodPattern.MatchString(*input.EndPeriod) {
			return nil, apperror.ErrInvalidPeriodFormat
		}
		if *input.EndPeriod < bill.StartPeriod {
			return nil, apperror.ErrInvalidPeriodFormat
		}
		bill.EndPeriod = input.EndPeriod
	}

	err = u.txManager.WithinTx(ctx, func(ctx context.Context) error {
		if err := u.billRepo.Update(ctx, bill); err != nil {
			return err
		}
		// Kalau due_day berubah, selaraskan due_date periode yang belum dibayar supaya
		// jatuh tempo periode ke depan mengikuti tanggal baru.
		if input.DueDay != nil {
			if err := u.billPeriodRepo.RecalcDueDatesFrom(ctx, bill.ID, *input.DueDay); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return bill, nil
}

// DeleteBill menghapus tagihan beserta seluruh periodenya (soft delete). Hanya bill milik
// household user yang bisa dihapus.
func (u *BillUsecase) DeleteBill(ctx context.Context, userID, billID uuid.UUID) error {
	member, err := u.householdRepo.FindMemberByUserID(ctx, userID)
	if err != nil {
		return err
	}

	if _, err := u.billRepo.FindByIDAndHousehold(ctx, billID, member.HouseholdID); err != nil {
		return err
	}

	return u.txManager.WithinTx(ctx, func(ctx context.Context) error {
		if err := u.billPeriodRepo.SoftDeleteByBillID(ctx, billID); err != nil {
			return err
		}
		return u.billRepo.SoftDelete(ctx, billID)
	})
}

// StopBill "menghentikan" tagihan berlangganan: menutup bill di bulan berjalan
// (is_active=false + end_period=bulan ini) dan menghapus periode masa depan yang
// belum dibayar supaya tidak menghasilkan tagihan baru lagi. Riwayat yang sudah
// dibayar tetap tersimpan.
func (u *BillUsecase) StopBill(ctx context.Context, userID, billID uuid.UUID) (*entity.Bill, error) {
	member, err := u.householdRepo.FindMemberByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	bill, err := u.billRepo.FindByIDAndHousehold(ctx, billID, member.HouseholdID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	stopPeriod := now.Format("2006-01")
	if stopPeriod < bill.StartPeriod {
		stopPeriod = bill.StartPeriod
	}
	if bill.EndPeriod != nil && bill.IsActive == false && stopPeriod >= *bill.EndPeriod {
		// sudah berhenti
		return bill, nil
	}

	err = u.txManager.WithinTx(ctx, func(ctx context.Context) error {
		bill.EndPeriod = &stopPeriod
		bill.IsActive = false
		if err := u.billRepo.Update(ctx, bill); err != nil {
			return err
		}
		return u.billPeriodRepo.DeleteUnpaidFrom(ctx, billID, stopPeriod)
	})
	if err != nil {
		return nil, err
	}

	return bill, nil
}

// PayBillPeriod menandai bill_period sebagai dibayar: buat transaksi expense lewat
// TransactionUsecase (reuse validasi & efek saldo Phase 06), lalu update status/transaction_id/paid_at.
func (u *BillUsecase) PayBillPeriod(ctx context.Context, userID, billPeriodID, accountID uuid.UUID, amount float64, transactionDate string) (*entity.BillPeriod, *repository.TransactionListItem, error) {
	member, err := u.householdRepo.FindMemberByUserID(ctx, userID)
	if err != nil {
		return nil, nil, err
	}

	period, err := u.billPeriodRepo.FindByID(ctx, billPeriodID)
	if err != nil {
		return nil, nil, err
	}
	if period.Status == entity.BillPeriodPaid {
		return nil, nil, apperror.ErrBillPeriodAlreadyPaid
	}

	bill, err := u.billRepo.FindByIDAndHousehold(ctx, period.BillID, member.HouseholdID)
	if err != nil {
		return nil, nil, err
	}

	tx, err := u.transactionUC.CreateTransaction(ctx, userID, TransactionInput{
		Type:            string(entity.TransactionExpense),
		AccountID:       accountID,
		CategoryID:      &bill.CategoryID,
		Amount:          amount,
		TransactionDate: transactionDate,
		BillPeriodID:    &billPeriodID,
	})
	if err != nil {
		return nil, nil, err
	}

	paidAt := time.Now()
	if err := u.billPeriodRepo.MarkPaid(ctx, billPeriodID, tx.ID, paidAt); err != nil {
		return nil, nil, err
	}

	period.Status = entity.BillPeriodPaid
	period.TransactionID = &tx.ID
	period.PaidAt = &paidAt

	return period, tx, nil
}
