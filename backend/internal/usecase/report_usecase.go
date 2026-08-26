package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"homeapp/internal/pkg/apperror"
	"homeapp/internal/pkg/export"
	"homeapp/internal/repository"
)

type ReportUsecase struct {
	transactionRepo repository.TransactionRepository
	householdRepo   repository.HouseholdRepository
}

func NewReportUsecase(transactionRepo repository.TransactionRepository, householdRepo repository.HouseholdRepository) *ReportUsecase {
	return &ReportUsecase{transactionRepo: transactionRepo, householdRepo: householdRepo}
}

const defaultTrendMonths = 6

// periodBounds menghitung [start, end) untuk period_type "month" (YYYY-MM) atau "year" (YYYY) —
// v1 sengaja cuma dua granularitas ini, bukan custom range (specs.md §5.10).
func periodBounds(period, periodType string) (start, end time.Time, err error) {
	switch periodType {
	case "month":
		start, err = time.Parse("2006-01", period)
		if err != nil {
			return time.Time{}, time.Time{}, apperror.ErrInvalidPeriodFormat
		}
		return start, start.AddDate(0, 1, 0), nil
	case "year":
		start, err = time.Parse("2006", period)
		if err != nil {
			return time.Time{}, time.Time{}, apperror.ErrInvalidPeriodFormat
		}
		return start, start.AddDate(1, 0, 0), nil
	default:
		return time.Time{}, time.Time{}, apperror.ErrInvalidPeriodType
	}
}

// previousBounds periode sebelumnya dengan granularitas sama (1 bulan atau 1 tahun ke belakang).
func previousBounds(start time.Time, periodType string) (prevStart, prevEnd time.Time) {
	if periodType == "year" {
		return start.AddDate(-1, 0, 0), start
	}
	return start.AddDate(0, -1, 0), start
}

func formatPeriodLabel(start time.Time, periodType string) string {
	if periodType == "year" {
		return start.Format("2006")
	}
	return start.Format("2006-01")
}

func (u *ReportUsecase) resolveHousehold(ctx context.Context, userID uuid.UUID) (uuid.UUID, error) {
	member, err := u.householdRepo.FindMemberByUserID(ctx, userID)
	if err != nil {
		return uuid.Nil, err
	}
	return member.HouseholdID, nil
}

func (u *ReportUsecase) GetTrend(ctx context.Context, userID uuid.UUID, months int) ([]repository.MonthlyTrendItem, error) {
	householdID, err := u.resolveHousehold(ctx, userID)
	if err != nil {
		return nil, err
	}
	if months <= 0 {
		months = defaultTrendMonths
	}
	return u.transactionRepo.MonthlyTrend(ctx, householdID, months)
}

type CategoryBreakdownResult struct {
	CategoryID   uuid.UUID `json:"category_id"`
	CategoryName string    `json:"category_name"`
	Total        float64   `json:"total"`
	Percentage   float64   `json:"percentage"`
}

func (u *ReportUsecase) GetCategoryBreakdown(ctx context.Context, userID uuid.UUID, period, periodType, txType string) ([]CategoryBreakdownResult, error) {
	householdID, err := u.resolveHousehold(ctx, userID)
	if err != nil {
		return nil, err
	}
	start, end, err := periodBounds(period, periodType)
	if err != nil {
		return nil, err
	}
	if txType == "" {
		txType = "expense"
	}

	items, err := u.transactionRepo.CategoryBreakdown(ctx, householdID, start, end, txType)
	if err != nil {
		return nil, err
	}

	var total float64
	for _, it := range items {
		total += it.Total
	}

	result := make([]CategoryBreakdownResult, len(items))
	for i, it := range items {
		pct := 0.0
		if total > 0 {
			pct = (it.Total / total) * 100
		}
		result[i] = CategoryBreakdownResult{
			CategoryID:   it.CategoryID,
			CategoryName: it.CategoryName,
			Total:        it.Total,
			Percentage:   pct,
		}
	}
	return result, nil
}

func (u *ReportUsecase) GetMemberBreakdown(ctx context.Context, userID uuid.UUID, period, periodType string) ([]repository.MemberBreakdownItem, error) {
	householdID, err := u.resolveHousehold(ctx, userID)
	if err != nil {
		return nil, err
	}
	start, end, err := periodBounds(period, periodType)
	if err != nil {
		return nil, err
	}
	return u.transactionRepo.MemberBreakdown(ctx, householdID, start, end)
}

type ComparisonPeriodTotal struct {
	Period       string  `json:"period"`
	TotalExpense float64 `json:"total_expense"`
}

type ComparisonCategoryItem struct {
	CategoryName   string  `json:"category_name"`
	Current        float64 `json:"current"`
	Previous       float64 `json:"previous"`
	DiffPercentage float64 `json:"diff_percentage"`
}

type ComparisonResult struct {
	Current       ComparisonPeriodTotal     `json:"current"`
	Previous      ComparisonPeriodTotal     `json:"previous"`
	DiffAmount    float64                   `json:"diff_amount"`
	DiffPercentage float64                  `json:"diff_percentage"`
	ByCategory    []ComparisonCategoryItem  `json:"by_category"`
}

func diffPercentage(current, previous float64) float64 {
	if previous == 0 {
		if current == 0 {
			return 0
		}
		return 100
	}
	return ((current - previous) / previous) * 100
}

func (u *ReportUsecase) GetComparison(ctx context.Context, userID uuid.UUID, period, periodType string) (*ComparisonResult, error) {
	householdID, err := u.resolveHousehold(ctx, userID)
	if err != nil {
		return nil, err
	}
	start, end, err := periodBounds(period, periodType)
	if err != nil {
		return nil, err
	}
	prevStart, prevEnd := previousBounds(start, periodType)

	currentTotal, err := u.transactionRepo.TotalByType(ctx, householdID, start, end, "expense")
	if err != nil {
		return nil, err
	}
	previousTotal, err := u.transactionRepo.TotalByType(ctx, householdID, prevStart, prevEnd, "expense")
	if err != nil {
		return nil, err
	}

	currentByCategory, err := u.transactionRepo.CategoryBreakdown(ctx, householdID, start, end, "expense")
	if err != nil {
		return nil, err
	}
	previousByCategory, err := u.transactionRepo.CategoryBreakdown(ctx, householdID, prevStart, prevEnd, "expense")
	if err != nil {
		return nil, err
	}

	prevByName := make(map[string]float64, len(previousByCategory))
	for _, c := range previousByCategory {
		prevByName[c.CategoryName] = c.Total
	}
	seen := make(map[string]bool, len(currentByCategory))

	byCategory := make([]ComparisonCategoryItem, 0, len(currentByCategory))
	for _, c := range currentByCategory {
		prev := prevByName[c.CategoryName]
		byCategory = append(byCategory, ComparisonCategoryItem{
			CategoryName:   c.CategoryName,
			Current:        c.Total,
			Previous:       prev,
			DiffPercentage: diffPercentage(c.Total, prev),
		})
		seen[c.CategoryName] = true
	}
	// Kategori yang ada di periode lalu tapi sudah tidak dipakai bulan/tahun ini — tetap tampil (turun 100%).
	for _, c := range previousByCategory {
		if seen[c.CategoryName] {
			continue
		}
		byCategory = append(byCategory, ComparisonCategoryItem{
			CategoryName:   c.CategoryName,
			Current:        0,
			Previous:       c.Total,
			DiffPercentage: diffPercentage(0, c.Total),
		})
	}

	return &ComparisonResult{
		Current:        ComparisonPeriodTotal{Period: formatPeriodLabel(start, periodType), TotalExpense: currentTotal},
		Previous:       ComparisonPeriodTotal{Period: formatPeriodLabel(prevStart, periodType), TotalExpense: previousTotal},
		DiffAmount:     currentTotal - previousTotal,
		DiffPercentage: diffPercentage(currentTotal, previousTotal),
		ByCategory:     byCategory,
	}, nil
}

// GenerateExport membuat file PDF/Excel berisi ringkasan yang sama dengan /reports (tren 6 bulan,
// breakdown kategori, breakdown anggota) untuk 1 periode terpilih — di memori, tidak ditulis ke disk.
func (u *ReportUsecase) GenerateExport(ctx context.Context, userID uuid.UUID, format, period, periodType string) (data []byte, filename, contentType string, err error) {
	if format != "pdf" && format != "excel" {
		return nil, "", "", apperror.ErrInvalidExportFormat
	}

	start, _, err := periodBounds(period, periodType)
	if err != nil {
		return nil, "", "", err
	}

	trend, err := u.GetTrend(ctx, userID, defaultTrendMonths)
	if err != nil {
		return nil, "", "", err
	}
	categories, err := u.GetCategoryBreakdown(ctx, userID, period, periodType, "expense")
	if err != nil {
		return nil, "", "", err
	}
	members, err := u.GetMemberBreakdown(ctx, userID, period, periodType)
	if err != nil {
		return nil, "", "", err
	}

	reportData := export.ReportData{
		PeriodLabel: formatPeriodLabel(start, periodType),
	}
	for _, t := range trend {
		reportData.Trend = append(reportData.Trend, export.TrendRow{Month: t.Month, Income: t.Income, Expense: t.Expense})
	}
	for _, c := range categories {
		reportData.CategoryBreakdown = append(reportData.CategoryBreakdown, export.CategoryRow{
			CategoryName: c.CategoryName, Total: c.Total, Percentage: c.Percentage,
		})
	}
	for _, m := range members {
		reportData.MemberBreakdown = append(reportData.MemberBreakdown, export.MemberRow{
			Name: m.Name, TotalExpense: m.TotalExpense, TotalIncome: m.TotalIncome,
		})
	}

	if format == "pdf" {
		data, err = export.GeneratePDF(reportData)
		return data, fmt.Sprintf("laporan-%s.pdf", period), "application/pdf", err
	}

	data, err = export.GenerateExcel(reportData)
	return data, fmt.Sprintf("laporan-%s.xlsx", period),
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", err
}
