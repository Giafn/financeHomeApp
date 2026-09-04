package dto

type CreateBillRequest struct {
	Name               string  `json:"name" validate:"required,max=255"`
	CategoryID         string  `json:"category_id" validate:"required,uuid"`
	Amount             float64 `json:"amount" validate:"required,gt=0"`
	DueDay             int     `json:"due_day" validate:"required,min=1,max=31"`
	StartPeriod        string  `json:"start_period" validate:"required,len=7"`
	EndPeriod          *string `json:"end_period" validate:"omitempty,len=7"`
	ReminderDaysBefore int     `json:"reminder_days_before" validate:"omitempty,min=1,max=30"`
}

type UpdateBillRequest struct {
	IsActive           *bool    `json:"is_active"`
	Name               *string  `json:"name" validate:"omitempty,max=255"`
	Amount             *float64 `json:"amount" validate:"omitempty,gt=0"`
	CategoryID         *string  `json:"category_id" validate:"omitempty,uuid"`
	ReminderDaysBefore *int     `json:"reminder_days_before" validate:"omitempty,min=1,max=30"`
	DueDay             *int     `json:"due_day" validate:"omitempty,min=1,max=31"`
	// Perubahan amount/nama/kategori di sini langsung memengaruhi periode yang belum dibayar
	// (periode acuan selalu memakai amount bill saat dibayar). Periode yang sudah dibayar
	// tidak berubah karena transaksinya sudah tercatat dengan nominal lama.
	EndPeriod *string `json:"end_period" validate:"omitempty,len=7"`
}

type PayBillPeriodRequest struct {
	AccountID       string  `json:"account_id" validate:"required,uuid"`
	Amount          float64 `json:"amount" validate:"required,gt=0"`
	TransactionDate string  `json:"transaction_date" validate:"required,datetime=2006-01-02"`
}

type BillPeriodResponse struct {
	ID            string  `json:"id"`
	BillID        string  `json:"bill_id"`
	Period        string  `json:"period"`
	DueDate       string  `json:"due_date"`
	Status        string  `json:"status"`
	TransactionID *string `json:"transaction_id,omitempty"`
	PaidAt        *string `json:"paid_at,omitempty"`
}

type BillResponse struct {
	ID                 string              `json:"id"`
	Name               string              `json:"name"`
	CategoryID         string              `json:"category_id"`
	Amount             float64             `json:"amount"`
	DueDay             int                 `json:"due_day"`
	StartPeriod        string              `json:"start_period"`
	EndPeriod          *string             `json:"end_period,omitempty"`
	ReminderDaysBefore int                 `json:"reminder_days_before"`
	IsActive           bool                `json:"is_active"`
	NextPeriod         *BillPeriodResponse `json:"next_period,omitempty"`
}

type CreateBillResponse struct {
	BillResponse
	BillPeriods []BillPeriodResponse `json:"bill_periods"`
}
