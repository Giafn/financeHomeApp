package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"homeapp/internal/entity"
)

type TransactionFilter struct {
	Type       string
	AccountID  *uuid.UUID
	CategoryID *uuid.UUID
	CreatedBy  *uuid.UUID
	GoalID     *uuid.UUID
	DateFrom   *time.Time
	DateTo     *time.Time
	Page       int
	Limit      int
}

// TransactionListItem gabungan transaksi + nama akun/kategori/pembuat hasil join.
type TransactionListItem struct {
	entity.Transaction
	AccountName   string  `json:"account_name"`
	CategoryName  *string `json:"category_name"`
	CreatedByName string  `json:"created_by_name"`
}

// MonthlyTrendItem agregat income vs expense per bulan — dipakai grafik tren dashboard.
type MonthlyTrendItem struct {
	Month   string  `json:"month"`
	Income  float64 `json:"income"`
	Expense float64 `json:"expense"`
}

// MemberBreakdownItem agregat expense/income per anggota household dalam suatu rentang tanggal.
type MemberBreakdownItem struct {
	UserID       uuid.UUID `json:"user_id"`
	Name         string    `json:"name"`
	TotalExpense float64   `json:"total_expense"`
	TotalIncome  float64   `json:"total_income"`
}

// CategoryBreakdownItem agregat total per kategori dalam suatu rentang tanggal + 1 tipe transaksi.
type CategoryBreakdownItem struct {
	CategoryID   uuid.UUID `json:"category_id"`
	CategoryName string    `json:"category_name"`
	Total        float64   `json:"total"`
}

type TransactionRepository interface {
	Create(ctx context.Context, transaction *entity.Transaction) error
	FindByID(ctx context.Context, id, householdID uuid.UUID) (*TransactionListItem, error)
	List(ctx context.Context, householdID uuid.UUID, filter TransactionFilter) ([]*TransactionListItem, int64, error)
	Update(ctx context.Context, transaction *entity.Transaction) error
	Delete(ctx context.Context, id, householdID uuid.UUID) error
	// LastUsedAccountAndCategory mengembalikan account_id & category_id dari transaksi
	// terakhir yang dibuat user ini di household tsb, dipakai untuk quick-select form.
	LastUsedAccountAndCategory(ctx context.Context, householdID, userID uuid.UUID) (*uuid.UUID, *uuid.UUID, error)
	// IsLinkedToPaidBillPeriod true kalau transaksi ini transaksi pembayaran bill_period berstatus 'paid'.
	// Query minimal langsung ke tabel bill_periods — modul Bill penuh baru dibangun di Phase 10.
	IsLinkedToPaidBillPeriod(ctx context.Context, transactionID uuid.UUID) (bool, error)
	// MonthlyTrend income vs expense per bulan, `months` bulan terakhir termasuk bulan berjalan.
	MonthlyTrend(ctx context.Context, householdID uuid.UUID, months int) ([]MonthlyTrendItem, error)
	// MemberBreakdown agregat expense/income per anggota dalam rentang [start, end).
	// Rentang eksplisit (bukan period string) supaya dipakai untuk periode bulan MAUPUN tahun (Phase 12).
	MemberBreakdown(ctx context.Context, householdID uuid.UUID, start, end time.Time) ([]MemberBreakdownItem, error)
	// CategoryBreakdown agregat per kategori untuk 1 tipe transaksi dalam rentang [start, end).
	CategoryBreakdown(ctx context.Context, householdID uuid.UUID, start, end time.Time, txType string) ([]CategoryBreakdownItem, error)
	// TotalByType jumlah total transaksi 1 tipe dalam rentang [start, end) — dipakai laporan perbandingan.
	TotalByType(ctx context.Context, householdID uuid.UUID, start, end time.Time, txType string) (float64, error)
}
