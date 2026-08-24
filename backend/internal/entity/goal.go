package entity

import (
	"time"

	"github.com/google/uuid"
)

type GoalStatus string

const (
	GoalActive    GoalStatus = "active"
	GoalAchieved  GoalStatus = "achieved"
	GoalCancelled GoalStatus = "cancelled"
)

// Goal adalah target tabungan yang terhubung ke sebuah Account tujuan (LinkedAccountID).
// Progress dihitung dari akumulasi transaksi transfer yang ditandai GoalID, bukan dari saldo akun langsung.
type Goal struct {
	BaseModel
	HouseholdID     uuid.UUID  `gorm:"type:uuid;not null;index" json:"household_id"`
	Name            string     `gorm:"type:varchar(255);not null" json:"name"`
	Icon            *string    `gorm:"type:varchar(50)" json:"icon,omitempty"`
	TargetAmount    float64    `gorm:"type:numeric(18,2);not null" json:"target_amount"`
	LinkedAccountID uuid.UUID  `gorm:"type:uuid;not null" json:"linked_account_id"`
	TargetDate      *time.Time `gorm:"type:date" json:"target_date,omitempty"`
	Status          GoalStatus `gorm:"type:varchar(20);not null;default:'active'" json:"status"`
	CreatedBy       uuid.UUID  `gorm:"type:uuid;not null" json:"created_by"`
}

func (Goal) TableName() string { return "goals" }
