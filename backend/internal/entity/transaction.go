package entity

import (
	"time"

	"github.com/google/uuid"
)

type TransactionType string

const (
	TransactionIncome   TransactionType = "income"
	TransactionExpense  TransactionType = "expense"
	TransactionTransfer TransactionType = "transfer"
)

// Transaction mencatat pemasukan, pengeluaran, atau transfer internal antar akun.
// GoalID/BillPeriodID diisi otomatis kalau transaksi ini adalah kontribusi goal atau pembayaran tagihan.
type Transaction struct {
	BaseModel
	HouseholdID          uuid.UUID       `gorm:"type:uuid;not null;index" json:"household_id"`
	Type                 TransactionType `gorm:"type:varchar(20);not null" json:"type"`
	AccountID            uuid.UUID       `gorm:"type:uuid;not null;index" json:"account_id"`
	DestinationAccountID *uuid.UUID      `gorm:"type:uuid" json:"destination_account_id,omitempty"`
	CategoryID           *uuid.UUID      `gorm:"type:uuid" json:"category_id,omitempty"`
	Amount               float64         `gorm:"type:numeric(18,2);not null" json:"amount"`
	Description          *string         `gorm:"type:text" json:"description,omitempty"`
	TransactionDate      time.Time       `gorm:"type:date;not null" json:"transaction_date"`
	AttachmentURL        *string         `gorm:"type:varchar(500)" json:"attachment_url,omitempty"`
	GoalID               *uuid.UUID      `gorm:"type:uuid" json:"goal_id,omitempty"`
	BillPeriodID         *uuid.UUID      `gorm:"type:uuid" json:"bill_period_id,omitempty"`
	CreatedBy            uuid.UUID       `gorm:"type:uuid;not null" json:"created_by"`
}

func (Transaction) TableName() string { return "transactions" }
