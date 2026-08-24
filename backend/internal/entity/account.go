package entity

import "github.com/google/uuid"

type AccountType string

const (
	AccountBank    AccountType = "bank"
	AccountEwallet AccountType = "ewallet"
	AccountCash    AccountType = "cash"
	AccountOther   AccountType = "other"
)

// Account merepresentasikan akun keuangan (bank/e-wallet/tunai/lainnya) milik sebuah Household.
// Kartu kredit sengaja belum didukung di v1.
type Account struct {
	BaseModel
	HouseholdID    uuid.UUID   `gorm:"type:uuid;not null;index" json:"household_id"`
	Name           string      `gorm:"type:varchar(255);not null" json:"name"`
	Type           AccountType `gorm:"type:varchar(20);not null" json:"type"`
	InitialBalance float64     `gorm:"type:numeric(18,2);not null;default:0" json:"initial_balance"`
	IsActive       bool        `gorm:"not null;default:true" json:"is_active"`
	CreatedBy      uuid.UUID   `gorm:"type:uuid;not null" json:"created_by"`
}

func (Account) TableName() string { return "accounts" }
