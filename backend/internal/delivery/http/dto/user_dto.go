package dto

type UserProfileResponse struct {
	ID        string           `json:"id"`
	Name      string           `json:"name"`
	Email     string           `json:"email"`
	AvatarURL *string          `json:"avatar_url"`
	Household *UserHousehold   `json:"household"`
}

type UserHousehold struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

type UpdateProfileRequest struct {
	Name      *string `json:"name" validate:"omitempty"`
	AvatarURL *string `json:"avatar_url" validate:"omitempty,url"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

type MessageResponse struct {
	Message string `json:"message"`
}
