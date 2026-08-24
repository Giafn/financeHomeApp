package dto

type CreateHouseholdRequest struct {
	Name string `json:"name" validate:"required,min=2"`
}

type JoinHouseholdRequest struct {
	Code string `json:"code" validate:"required"`
}

type HouseholdResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type InvitationResponse struct {
	Code      string `json:"code"`
	ExpiresAt string `json:"expires_at"`
}

type UpdateHouseholdRequest struct {
	Name string `json:"name" validate:"required,min=2"`
}
