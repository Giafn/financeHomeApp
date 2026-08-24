package migration

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// NewMigrator membungkus gormigrate dengan daftar Migrations di atas.
// gormigrate otomatis membuat & mengelola tabel `migrations` untuk melacak versi yang sudah jalan.
func NewMigrator(db *gorm.DB) *gormigrate.Gormigrate {
	return gormigrate.New(db, gormigrate.DefaultOptions, Migrations)
}
