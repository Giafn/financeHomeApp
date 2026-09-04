package entity

import "github.com/google/uuid"

type AccountType string

const (
	AccountBank    AccountType = "bank"
	AccountEwallet AccountType = "ewallet"
	AccountCash    AccountType = "cash"
	AccountOther   AccountType = "other"
)

type AccountOwnerType string

const (
	AccountOwnerHousehold AccountOwnerType = "household"
	AccountOwnerPersonal  AccountOwnerType = "personal"
)

// Account merepresentasikan akun keuangan (bank/e-wallet/tunai/lainnya) milik sebuah Household.
// Kartu kredit sengaja belum didukung di v1.
//
// OwnerType membedakan akun bersama (household, default — semua anggota bisa transaksi, perilaku
// lama) dari akun personal (cuma OwnerUserID yang boleh transaksi, anggota lain read-only).
type Account struct {
	BaseModel
	HouseholdID    uuid.UUID        `gorm:"type:uuid;not null;index" json:"household_id"`
	Name           string           `gorm:"type:varchar(255);not null" json:"name"`
	Type           AccountType      `gorm:"type:varchar(20);not null" json:"type"`
	InitialBalance float64          `gorm:"type:numeric(18,2);not null;default:0" json:"initial_balance"`
	IsActive       bool             `gorm:"not null;default:true" json:"is_active"`
	OwnerType      AccountOwnerType `gorm:"type:varchar(20);not null;default:'household'" json:"owner_type"`
	OwnerUserID    *uuid.UUID       `gorm:"type:uuid" json:"owner_user_id,omitempty"`
	CreatedBy      uuid.UUID        `gorm:"type:uuid;not null" json:"created_by"`
}

func (Account) TableName() string { return "accounts" }
