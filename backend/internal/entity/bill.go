package entity

import "github.com/google/uuid"

// Bill adalah tagihan berulang dengan rentang periode eksplisit (StartPeriod..EndPeriod, format "YYYY-MM").
// EndPeriod nil berarti tagihan berulang tanpa batas akhir.
type Bill struct {
	BaseModel
	HouseholdID        uuid.UUID `gorm:"type:uuid;not null;index" json:"household_id"`
	Name               string    `gorm:"type:varchar(255);not null" json:"name"`
	CategoryID         uuid.UUID `gorm:"type:uuid;not null" json:"category_id"`
	Amount             float64   `gorm:"type:numeric(18,2);not null" json:"amount"`
	DueDay             int       `gorm:"not null" json:"due_day"`
	StartPeriod        string    `gorm:"type:varchar(7);not null" json:"start_period"`
	EndPeriod          *string   `gorm:"type:varchar(7)" json:"end_period,omitempty"`
	ReminderDaysBefore int       `gorm:"not null;default:5" json:"reminder_days_before"`
	IsActive           bool      `gorm:"not null;default:true" json:"is_active"`
	CreatedBy          uuid.UUID `gorm:"type:uuid;not null" json:"created_by"`
}

func (Bill) TableName() string { return "bills" }
