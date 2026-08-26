package dto

type CreateAccountRequest struct {
	Name           string  `json:"name" validate:"required,min=2"`
	Type           string  `json:"type" validate:"required,oneof=bank ewallet cash other"`
	InitialBalance float64 `json:"initial_balance" validate:"gte=0"`
}

type UpdateAccountRequest struct {
	Name     *string `json:"name"`
	Type     *string `json:"type" validate:"omitempty,oneof=bank ewallet cash other"`
	IsActive *bool   `json:"is_active"`
}

type AccountResponse struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	Type            string  `json:"type"`
	InitialBalance  float64 `json:"initial_balance"`
	CurrentBalance  float64 `json:"current_balance"`
	IsActive        bool    `json:"is_active"`
}
