package dto

type CreateHouseholdRequest struct {
	Name string `json:"name" validate:"required,min=2"`
}

type JoinHouseholdRequest struct {
	Code string `json:"code" validate:"required"`
}

type HouseholdResponse struct {
	ID                  string `json:"id"`
	Name                string `json:"name"`
	BudgetCycleStartDay int    `json:"budget_cycle_start_day"`
}

type InvitationResponse struct {
	Code      string `json:"code"`
	ExpiresAt string `json:"expires_at"`
}

type UpdateHouseholdRequest struct {
	Name *string `json:"name" validate:"omitempty,min=2"`
	// BudgetCycleStartDay opsional, 1-28. >1 berarti transaksi sebelum tanggal ini dianggap
	// milik periode bulan berikutnya (misal gajian tanggal 25).
	BudgetCycleStartDay *int `json:"budget_cycle_start_day" validate:"omitempty,min=1,max=28"`
}
