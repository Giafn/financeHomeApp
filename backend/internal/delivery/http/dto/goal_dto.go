package dto

type CreateGoalRequest struct {
	Name            string  `json:"name" validate:"required,max=255"`
	Icon            *string `json:"icon" validate:"omitempty,max=50"`
	TargetAmount    float64 `json:"target_amount" validate:"required,gt=0"`
	LinkedAccountID string  `json:"linked_account_id" validate:"required,uuid"`
	TargetDate      *string `json:"target_date" validate:"omitempty,datetime=2006-01-02"`
}

type UpdateGoalRequest struct {
	Name         *string  `json:"name" validate:"omitempty,max=255"`
	Icon         *string  `json:"icon" validate:"omitempty,max=50"`
	TargetAmount *float64 `json:"target_amount" validate:"omitempty,gt=0"`
	TargetDate   *string  `json:"target_date" validate:"omitempty,datetime=2006-01-02"`
	Status       *string  `json:"status" validate:"omitempty,oneof=active achieved cancelled"`
}

type GoalResponse struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	Icon            *string `json:"icon,omitempty"`
	TargetAmount    float64 `json:"target_amount"`
	LinkedAccountID string  `json:"linked_account_id"`
	TargetDate      *string `json:"target_date,omitempty"`
	Status          string  `json:"status"`
	CurrentAmount   float64 `json:"current_amount"`
	Percentage      float64 `json:"percentage"`
}

type GoalDetailResponse struct {
	GoalResponse
	Contributions []TransactionResponse `json:"contributions"`
}
