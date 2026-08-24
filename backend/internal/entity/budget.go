package entity

import "github.com/google/uuid"

// Budget adalah anggaran per kategori per bulan (format period: "YYYY-MM").
type Budget struct {
	BaseModel
	HouseholdID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_budget_period" json:"household_id"`
	CategoryID  uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_budget_period" json:"category_id"`
	Period      string    `gorm:"type:varchar(7);not null;uniqueIndex:idx_budget_period" json:"period"`
	Amount      float64   `gorm:"type:numeric(18,2);not null" json:"amount"`
	CreatedBy   uuid.UUID `gorm:"type:uuid;not null" json:"created_by"`
}

func (Budget) TableName() string { return "budgets" }
