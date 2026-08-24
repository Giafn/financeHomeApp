package entity

// User merepresentasikan akun pengguna.
// PasswordHash dipakai untuk login email/password (JWT).
// GoogleID disiapkan untuk dukungan login Google di kemudian hari (opsional, nullable).
type User struct {
	BaseModel
	Email        string  `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	Name         string  `gorm:"type:varchar(255);not null" json:"name"`
	PasswordHash *string `gorm:"type:varchar(255)" json:"-"`
	GoogleID     *string `gorm:"type:varchar(255);uniqueIndex" json:"-"`
	AvatarURL    *string `gorm:"type:varchar(500)" json:"avatar_url,omitempty"`
}

func (User) TableName() string { return "users" }
