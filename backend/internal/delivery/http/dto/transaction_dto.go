package dto

type CreateTransactionRequest struct {
	Type                 string  `json:"type" validate:"required,oneof=income expense transfer"`
	AccountID            string  `json:"account_id" validate:"required,uuid"`
	DestinationAccountID *string `json:"destination_account_id" validate:"omitempty,uuid"`
	CategoryID           *string `json:"category_id" validate:"omitempty,uuid"`
	Amount               float64 `json:"amount" validate:"required,gt=0"`
	AdminFee             float64 `json:"admin_fee" validate:"min=0"`
	Description          *string `json:"description"`
	TransactionDate      string  `json:"transaction_date" validate:"required,datetime=2006-01-02"`
	AttachmentURL        *string `json:"attachment_url"`
	// GoalID opsional — diisi flow "Nabung" di halaman /goals (Phase 09), transaksi generik
	// tidak butuh tahu logic goal, cukup terima & simpan FK ini.
	GoalID *string `json:"goal_id" validate:"omitempty,uuid"`
}

type UpdateTransactionRequest struct {
	Type                 *string  `json:"type" validate:"omitempty,oneof=income expense transfer"`
	AccountID            *string  `json:"account_id" validate:"omitempty,uuid"`
	DestinationAccountID *string  `json:"destination_account_id" validate:"omitempty,uuid"`
	CategoryID           *string  `json:"category_id" validate:"omitempty,uuid"`
	Amount               *float64 `json:"amount" validate:"omitempty,gt=0"`
	AdminFee             *float64 `json:"admin_fee" validate:"omitempty,min=0"`
	Description          *string  `json:"description"`
	TransactionDate      *string  `json:"transaction_date" validate:"omitempty,datetime=2006-01-02"`
	AttachmentURL        *string  `json:"attachment_url"`
	GoalID               *string  `json:"goal_id" validate:"omitempty,uuid"`
}

type TransactionResponse struct {
	ID                   string  `json:"id"`
	HouseholdID          string  `json:"household_id"`
	Type                 string  `json:"type"`
	AccountID            string  `json:"account_id"`
	AccountName          string  `json:"account_name"`
	DestinationAccountID *string `json:"destination_account_id,omitempty"`
	CategoryID           *string `json:"category_id,omitempty"`
	CategoryName         *string `json:"category_name,omitempty"`
	Amount               float64 `json:"amount"`
	AdminFee             float64 `json:"admin_fee"`
	Description          *string `json:"description,omitempty"`
	TransactionDate      string  `json:"transaction_date"`
	AttachmentURL        *string `json:"attachment_url,omitempty"`
	GoalID               *string `json:"goal_id,omitempty"`
	BillPeriodID         *string `json:"bill_period_id,omitempty"`
	CreatedBy            string  `json:"created_by"`
	CreatedByName        string  `json:"created_by_name"`
	CreatedAt            string  `json:"created_at"`
}

type TransactionListResponse struct {
	Items      []TransactionResponse `json:"items"`
	Pagination PaginationResponse    `json:"pagination"`
}

type PaginationResponse struct {
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
	Total int64 `json:"total"`
}

type QuickSelectResponse struct {
	LastAccountID  *string `json:"last_account_id,omitempty"`
	LastCategoryID *string `json:"last_category_id,omitempty"`
}

type PresignUploadRequest struct {
	Filename    string `json:"filename" validate:"required"`
	ContentType string `json:"content_type" validate:"required"`
}

type PresignUploadResponse struct {
	UploadURL string `json:"upload_url"`
	FileURL   string `json:"file_url"`
}
