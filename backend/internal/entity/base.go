package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BaseModel adalah field dasar yang dipakai semua entity: UUID primary key + timestamp + soft delete.
// UUID digenerate oleh Postgres lewat fungsi gen_random_uuid() (ekstensi pgcrypto, lihat migrasi pertama).
type BaseModel struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
