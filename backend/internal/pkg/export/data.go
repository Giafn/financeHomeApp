package export

// ReportData adalah bentuk data netral (bukan tergantung repository/usecase) yang dikonsumsi
// generator PDF/Excel — biar generator tidak coupled ke tipe internal usecase.
type ReportData struct {
	PeriodLabel     string
	Trend           []TrendRow
	CategoryBreakdown []CategoryRow
	MemberBreakdown []MemberRow
}

type TrendRow struct {
	Month   string
	Income  float64
	Expense float64
}

type CategoryRow struct {
	CategoryName string
	Total        float64
	Percentage   float64
}

type MemberRow struct {
	Name         string
	TotalExpense float64
	TotalIncome  float64
}
