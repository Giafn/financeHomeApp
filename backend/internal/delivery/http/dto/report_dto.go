package dto

type CategoryBreakdownResponse struct {
	CategoryID   string  `json:"category_id"`
	CategoryName string  `json:"category_name"`
	Total        float64 `json:"total"`
	Percentage   float64 `json:"percentage"`
}

type MemberBreakdownResponse struct {
	UserID       string  `json:"user_id"`
	Name         string  `json:"name"`
	TotalExpense float64 `json:"total_expense"`
	TotalIncome  float64 `json:"total_income"`
}

type ComparisonPeriodResponse struct {
	Period       string  `json:"period"`
	TotalExpense float64 `json:"total_expense"`
}

type ComparisonCategoryResponse struct {
	CategoryName   string  `json:"category_name"`
	Current        float64 `json:"current"`
	Previous       float64 `json:"previous"`
	DiffPercentage float64 `json:"diff_percentage"`
}

type ComparisonResponse struct {
	Current        ComparisonPeriodResponse     `json:"current"`
	Previous       ComparisonPeriodResponse     `json:"previous"`
	DiffAmount     float64                      `json:"diff_amount"`
	DiffPercentage float64                      `json:"diff_percentage"`
	ByCategory     []ComparisonCategoryResponse `json:"by_category"`
}
