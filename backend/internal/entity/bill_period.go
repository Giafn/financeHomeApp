package entity

import (
	"time"

	"github.com/google/uuid"
)

type BillPeriodStatus string

const (
	BillPeriodUpcoming BillPeriodStatus = "upcoming"
	BillPeriodPaid     BillPeriodStatus = "paid"
	BillPeriodOverdue  BillPeriodStatus = "overdue"
)

// BillPeriod adalah instance bulanan dari sebuah Bill. Saat dibayar, TransactionID diisi
// dengan transaksi expense yang otomatis dibuat.
type BillPeriod struct {
	BaseModel
	BillID        uuid.UUID        `gorm:"type:uuid;not null;uniqueIndex:idx_bill_period" json:"bill_id"`
	Period        string           `gorm:"type:varchar(7);not null;uniqueIndex:idx_bill_period" json:"period"`
	DueDate       time.Time        `gorm:"type:date;not null" json:"due_date"`
	Status        BillPeriodStatus `gorm:"type:varchar(20);not null;default:'upcoming'" json:"status"`
	TransactionID *uuid.UUID       `gorm:"type:uuid" json:"transaction_id,omitempty"`
	PaidAt        *time.Time       `json:"paid_at,omitempty"`
}

func (BillPeriod) TableName() string { return "bill_periods" }
