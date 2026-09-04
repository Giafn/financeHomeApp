package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"homeapp/internal/entity"
)

// BillPeriodWithBill gabungan bill_period + data bill induk (nama, category_id, household_id)
// hasil join, dipakai job reminder/overdue & response list yang butuh konteks bill.
type BillPeriodWithBill struct {
	entity.BillPeriod
	BillName    string    `json:"bill_name"`
	HouseholdID uuid.UUID `json:"household_id"`
	CategoryID  uuid.UUID `json:"category_id"`
	Amount      float64   `json:"amount"`
}

type BillPeriodRepository interface {
	Create(ctx context.Context, period *entity.BillPeriod) error
	CreateBulk(ctx context.Context, periods []*entity.BillPeriod) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.BillPeriod, error)
	FindByBillAndPeriod(ctx context.Context, billID uuid.UUID, period string) (*entity.BillPeriod, error)
	ListByBillID(ctx context.Context, billID uuid.UUID) ([]*entity.BillPeriod, error)
	// LatestByBillID period bulan terakhir yang sudah ter-generate untuk bill ini (buat job generator).
	LatestByBillID(ctx context.Context, billID uuid.UUID) (*entity.BillPeriod, error)
	// NextUpcomingByBillID period upcoming/overdue terdekat — ringkasan di GET /bills.
	NextUpcomingByBillID(ctx context.Context, billID uuid.UUID) (*entity.BillPeriod, error)
	MarkPaid(ctx context.Context, id, transactionID uuid.UUID, paidAt time.Time) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status entity.BillPeriodStatus) error
	// RecalcDueDatesFrom memperbarui due_date periode yang belum dibayar (upcoming/overdue)
	// dengan due-day baru, dipanggil saat due_day bill diubah.
	RecalcDueDatesFrom(ctx context.Context, billID uuid.UUID, dueDay int) error
	// SoftDeleteByBillID menandai semua periode milik bill sebagai dihapus.
	SoftDeleteByBillID(ctx context.Context, billID uuid.UUID) error
	// DeleteUnpaidFrom menghapus periode belum dibayar (upcoming/overdue) mulai bulan `period`
	// ke depan — dipakai "stop tagihan" supaya periode masa depan tidak muncul lagi.
	DeleteUnpaidFrom(ctx context.Context, billID uuid.UUID, period string) error
	// ListDueForReminder upcoming period dengan due_date dalam <= reminder_days_before hari dari today,
	// belum lewat due_date (bukan overdue) — dipakai job bill-reminder-check.
	ListDueForReminder(ctx context.Context, today time.Time) ([]*BillPeriodWithBill, error)
	// ListOverdue upcoming period yang due_date-nya sudah lewat today — dipakai job bill-period-overdue-check.
	ListOverdue(ctx context.Context, today time.Time) ([]*entity.BillPeriod, error)
	// ListUpcomingForHousehold period upcoming dalam `days` hari ke depan untuk 1 household — dashboard (Phase 11).
	ListUpcomingForHousehold(ctx context.Context, householdID uuid.UUID, days int) ([]*BillPeriodWithBill, error)
	// ListUnpaidByHouseholdAndPeriod period (upcoming/overdue, belum 'paid') untuk 1 household di 1
	// periode tertentu — dipakai endpoint Rencana Anggaran (gabungan budget+bills).
	ListUnpaidByHouseholdAndPeriod(ctx context.Context, householdID uuid.UUID, period string) ([]*BillPeriodWithBill, error)
}
