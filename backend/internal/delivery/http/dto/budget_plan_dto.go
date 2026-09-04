package dto

type BudgetPlanCategoryResponse struct {
	CategoryID   string  `json:"category_id"`
	CategoryName string  `json:"category_name"`
	Amount       float64 `json:"amount"`
	Spent        float64 `json:"spent"`
	Percentage   float64 `json:"percentage"`
}

type BudgetPlanUnpaidBillResponse struct {
	BillPeriodID string  `json:"bill_period_id"`
	BillName     string  `json:"bill_name"`
	Amount       float64 `json:"amount"`
	DueDate      string  `json:"due_date"`
	Status       string  `json:"status"`
}

type BudgetPlanResponse struct {
	Period               string                         `json:"period"`
	TotalIncome          float64                        `json:"total_income"`
	TotalBudgeted        float64                        `json:"total_budgeted"`
	RemainingUnallocated float64                        `json:"remaining_unallocated"`
	Categories           []BudgetPlanCategoryResponse   `json:"categories"`
	UnpaidBills          []BudgetPlanUnpaidBillResponse `json:"unpaid_bills"`

	// CurrentHouseholdBalance dst. jawab pertanyaan inti app: cukup gak uang bersama SEKARANG
	// buat nutup sisa budget + tagihan belum lunas (bukan soal pemasukan periode ini).
	CurrentHouseholdBalance float64 `json:"current_household_balance"`
	TotalNeeded             float64 `json:"total_needed"`
	Surplus                 float64 `json:"surplus"`
	IsSufficient            bool    `json:"is_sufficient"`
}
