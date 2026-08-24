package entity

import "github.com/google/uuid"

type CategoryType string

const (
	CategoryIncome  CategoryType = "income"
	CategoryExpense CategoryType = "expense"
)

// Category adalah kategori transaksi custom milik sebuah Household.
type Category struct {
	BaseModel
	HouseholdID uuid.UUID    `gorm:"type:uuid;not null;index" json:"household_id"`
	Name        string       `gorm:"type:varchar(255);not null" json:"name"`
	Type        CategoryType `gorm:"type:varchar(20);not null" json:"type"`
	Icon        *string      `gorm:"type:varchar(50)" json:"icon,omitempty"`
	Color       *string      `gorm:"type:varchar(20)" json:"color,omitempty"`
	IsArchived  bool         `gorm:"not null;default:false" json:"is_archived"`
	CreatedBy   uuid.UUID    `gorm:"type:uuid;not null" json:"created_by"`
}

func (Category) TableName() string { return "categories" }
