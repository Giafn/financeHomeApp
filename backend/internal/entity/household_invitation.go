package entity

import (
	"time"

	"github.com/google/uuid"
)

type InvitationStatus string

const (
	InvitationActive  InvitationStatus = "active"
	InvitationUsed    InvitationStatus = "used"
	InvitationExpired InvitationStatus = "expired"
)

// HouseholdInvitation adalah kode undangan untuk bergabung ke sebuah Household.
// Default kadaluarsa 7 hari sejak dibuat (lihat usecase), sekali pakai.
type HouseholdInvitation struct {
	BaseModel
	HouseholdID uuid.UUID        `gorm:"type:uuid;not null;index" json:"household_id"`
	Code        string           `gorm:"type:varchar(16);uniqueIndex;not null" json:"code"`
	CreatedBy   uuid.UUID        `gorm:"type:uuid;not null" json:"created_by"`
	ExpiresAt   time.Time        `json:"expires_at"`
	UsedBy      *uuid.UUID       `gorm:"type:uuid" json:"used_by,omitempty"`
	UsedAt      *time.Time       `json:"used_at,omitempty"`
	Status      InvitationStatus `gorm:"type:varchar(20);not null;default:'active'" json:"status"`
}

func (HouseholdInvitation) TableName() string { return "household_invitations" }
