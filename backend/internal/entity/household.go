package entity

import "github.com/google/uuid"

// Household adalah unit tenant utama (rumah tangga).
type Household struct {
	BaseModel
	Name      string    `gorm:"type:varchar(255);not null" json:"name"`
	CreatedBy uuid.UUID `gorm:"type:uuid;not null" json:"created_by"`
}

func (Household) TableName() string { return "households" }
