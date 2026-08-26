package dto

type CreateBudgetRequest struct {
	CategoryID string  `json:"category_id" validate:"required,uuid"`
	Period     string  `json:"period" validate:"required,len=7"`
	Amount     float64 `json:"amount" validate:"required,gt=0"`
}

type UpdateBudgetRequest struct {
	Amount float64 `json:"amount" validate:"required,gt=0"`
}

type BudgetResponse struct {
	ID           string  `json:"id"`
	CategoryID   string  `json:"category_id"`
	CategoryName string  `json:"category_name"`
	Period       string  `json:"period"`
	Amount       float64 `json:"amount"`
	Spent        float64 `json:"spent"`
	Percentage   float64 `json:"percentage"`
}
