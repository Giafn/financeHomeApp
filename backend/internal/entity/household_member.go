package entity

import (
	"time"

	"github.com/google/uuid"
)

type HouseholdRole string

const (
	RoleOwner  HouseholdRole = "owner"
	RoleMember HouseholdRole = "member"
)

// HouseholdMember menghubungkan User ke Household beserta perannya.
// Satu user hanya boleh tergabung di satu household (uniqueIndex household_id+user_id).
type HouseholdMember struct {
	BaseModel
	HouseholdID uuid.UUID     `gorm:"type:uuid;not null;uniqueIndex:idx_household_user" json:"household_id"`
	UserID      uuid.UUID     `gorm:"type:uuid;not null;uniqueIndex:idx_household_user" json:"user_id"`
	Role        HouseholdRole `gorm:"type:varchar(20);not null" json:"role"`
	JoinedAt    time.Time     `json:"joined_at"`

	Household Household `gorm:"foreignKey:HouseholdID" json:"-"`
	User      User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (HouseholdMember) TableName() string { return "household_members" }
