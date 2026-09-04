package entity

import "github.com/google/uuid"

// Household adalah unit tenant utama (rumah tangga).
type Household struct {
	BaseModel
	Name      string    `gorm:"type:varchar(255);not null" json:"name"`
	CreatedBy uuid.UUID `gorm:"type:uuid;not null" json:"created_by"`
	// BudgetCycleStartDay 1-28, default 1 (= kalender biasa, tidak ada shift). >1 berarti
	// transaksi sebelum tanggal ini dianggap milik periode bulan berikutnya — lihat
	// internal/pkg/period untuk perhitungan window-nya.
	BudgetCycleStartDay int `gorm:"not null;default:1" json:"budget_cycle_start_day"`
}

func (Household) TableName() string { return "households" }
