package entity

import (
	"time"
)

// User merepresentasikan akun pengguna.
// PasswordHash dipakai untuk login email/password (JWT).
// GoogleID disiapkan untuk dukungan login Google di kemudian hari (opsional, nullable).
type User struct {
	BaseModel
	Email                string     `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	Name                 string     `gorm:"type:varchar(255);not null" json:"name"`
	PasswordHash         *string    `gorm:"type:varchar(255)" json:"-"`
	GoogleID             *string    `gorm:"type:varchar(255);uniqueIndex" json:"-"`
	AvatarURL            *string    `gorm:"type:varchar(500)" json:"avatar_url,omitempty"`
	EmailVerifiedAt      *time.Time `json:"email_verified_at"`
	VerificationToken    *string    `gorm:"type:varchar(255)" json:"-"`
	VerificationTokenExp *time.Time `json:"-"`
}

func (User) TableName() string { return "users" }

// Verified mengembalikan true jika email sudah diverifikasi.
func (u *User) Verified() bool {
	return u.EmailVerifiedAt != nil
}
