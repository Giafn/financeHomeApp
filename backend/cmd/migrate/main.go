// Binary migrasi TERPISAH dari aplikasi utama (cmd/api).
// Migrasi TIDAK PERNAH dijalankan otomatis saat server start — harus dijalankan manual.
//
// Penggunaan:
//   go run ./cmd/migrate up       # menjalankan semua migrasi yang belum jalan
//   go run ./cmd/migrate down     # rollback 1 migrasi terakhir
//   go run ./cmd/migrate status   # info cara cek status migrasi
//
// Atau lewat Makefile:
//   make migrate-up
//   make migrate-down
package main

import (
	"fmt"
	"log"
	"os"

	"family-finance-api/internal/config"
	"family-finance-api/internal/database"
	"family-finance-api/internal/database/migration"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Penggunaan: go run ./cmd/migrate [up|down|status]")
		os.Exit(1)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("gagal load config: %v", err)
	}

	db, err := database.NewPostgresConnection(cfg)
	if err != nil {
		log.Fatalf("gagal konek database: %v", err)
	}

	m := migration.NewMigrator(db)

	switch os.Args[1] {
	case "up":
		if err := m.Migrate(); err != nil {
			log.Fatalf("migrasi gagal: %v", err)
		}
		fmt.Println("✅ Migrasi berhasil dijalankan.")
	case "down":
		if err := m.RollbackLast(); err != nil {
			log.Fatalf("rollback gagal: %v", err)
		}
		fmt.Println("✅ Rollback migrasi terakhir berhasil.")
	case "status":
		fmt.Println("gormigrate mencatat migrasi yang sudah jalan di tabel `migrations`.")
		fmt.Println("Cek langsung: SELECT * FROM migrations ORDER BY id;")
	default:
		fmt.Println("Perintah tidak dikenal. Gunakan: up | down | status")
		os.Exit(1)
	}
}
