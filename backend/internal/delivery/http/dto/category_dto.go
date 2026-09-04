package dto

type CreateCategoryRequest struct {
	Name     string  `json:"name" validate:"required,max=255"`
	Type     string  `json:"type" validate:"required,oneof=income expense"`
	Icon     *string `json:"icon" validate:"omitempty,max=50"`
	Color    *string `json:"color" validate:"omitempty,max=20"`
	// ParentID opsional — isi buat bikin sub-kategori di bawah kategori leaf yang sudah ada.
	ParentID *string `json:"parent_id" validate:"omitempty,uuid"`
}

type UpdateCategoryRequest struct {
	Name       *string `json:"name" validate:"omitempty,max=255"`
	Icon       *string `json:"icon" validate:"omitempty,max=50"`
	Color      *string `json:"color" validate:"omitempty,max=20"`
	IsArchived *bool   `json:"is_archived"`
	ParentID   *string `json:"parent_id" validate:"omitempty,uuid"`
}

type CategoryResponse struct {
	ID          string  `json:"id"`
	HouseholdID string  `json:"household_id"`
	Name        string  `json:"name"`
	Type        string  `json:"type"`
	Icon        *string `json:"icon,omitempty"`
	Color       *string `json:"color,omitempty"`
	IsArchived  bool    `json:"is_archived"`
	ParentID    *string `json:"parent_id,omitempty"`
	CreatedBy   string  `json:"created_by"`
	CreatedAt   string  `json:"created_at"`
}
