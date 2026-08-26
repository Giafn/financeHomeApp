package database

import (
	"homeapp/internal/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewPostgresConnection membuka koneksi GORM ke Postgres.
// CATATAN: fungsi ini TIDAK menjalankan migrasi apapun — hanya membuka koneksi.
func NewPostgresConnection(cfg *config.Config) (*gorm.DB, error) {
	logLevel := logger.Warn
	if cfg.AppEnv == "development" {
		logLevel = logger.Info
	}

	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, err
	}

	return db, nil
}
