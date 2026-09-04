package migration

import (
	"homeapp/internal/entity"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// Migrations berisi daftar migrasi terurut berdasarkan ID (versi).
//
// ATURAN PENTING:
//  1. JANGAN PERNAH mengubah migrasi yang sudah pernah dijalankan di environment manapun
//     (dev/staging/prod) — gormigrate mencatat ID yang sudah jalan di tabel `migrations`.
//  2. Selalu tambahkan migrasi BARU di bagian bawah list ini dengan ID baru
//     (format bebas asal urut & unik, disarankan: YYYYMMDDHHMMSS_deskripsi).
//  3. Migrasi TIDAK dijalankan otomatis saat `cmd/api` start. Jalankan manual via
//     `make migrate-up` atau `go run ./cmd/migrate up`.
var Migrations = []*gormigrate.Migration{
	{
		// Ekstensi pgcrypto dibutuhkan untuk default value gen_random_uuid() pada kolom UUID.
		ID: "20260101000001_enable_pgcrypto",
		Migrate: func(tx *gorm.DB) error {
			return tx.Exec(`CREATE EXTENSION IF NOT EXISTS pgcrypto`).Error
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Exec(`DROP EXTENSION IF EXISTS pgcrypto`).Error
		},
	},
	{
		ID: "20260101000002_create_users",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(&entity.User{})
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Migrator().DropTable(&entity.User{})
		},
	},
	{
		ID: "20260101000003_create_households",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(&entity.Household{})
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Migrator().DropTable(&entity.Household{})
		},
	},
	{
		ID: "20260101000004_create_household_members",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(&entity.HouseholdMember{})
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Migrator().DropTable(&entity.HouseholdMember{})
		},
	},
	{
		ID: "20260101000005_create_household_invitations",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(&entity.HouseholdInvitation{})
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Migrator().DropTable(&entity.HouseholdInvitation{})
		},
	},
	{
		ID: "20260101000006_create_accounts",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(&entity.Account{})
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Migrator().DropTable(&entity.Account{})
		},
	},
	{
		ID: "20260101000007_create_categories",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(&entity.Category{})
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Migrator().DropTable(&entity.Category{})
		},
	},
	{
		ID: "20260101000008_create_transactions",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(&entity.Transaction{})
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Migrator().DropTable(&entity.Transaction{})
		},
	},
	{
		ID: "20260101000009_create_budgets",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(&entity.Budget{})
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Migrator().DropTable(&entity.Budget{})
		},
	},
	{
		ID: "20260101000010_create_goals",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(&entity.Goal{})
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Migrator().DropTable(&entity.Goal{})
		},
	},
	{
		ID: "20260101000011_create_bills",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(&entity.Bill{})
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Migrator().DropTable(&entity.Bill{})
		},
	},
	{
		ID: "20260101000012_create_bill_periods",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(&entity.BillPeriod{})
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Migrator().DropTable(&entity.BillPeriod{})
		},
	},
	{
		ID: "20260101000013_create_notification_logs",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(&entity.NotificationLog{})
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Migrator().DropTable(&entity.NotificationLog{})
		},
	},
	{
		ID: "20260124000001_add_partial_unique_index_household_members",
		Migrate: func(tx *gorm.DB) error {
			// Drop old non-partial unique index if it exists
			_ = tx.Migrator().DropIndex(&entity.HouseholdMember{}, "idx_household_user")
			// Create new partial unique index: only active members (deleted_at IS NULL)
			// Allows same user to rejoin after being removed
			return tx.Exec(`
				CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS idx_uq_household_user_active
				ON household_members(household_id, user_id)
				WHERE deleted_at IS NULL
			`).Error
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Exec(`DROP INDEX IF EXISTS idx_uq_household_user_active`).Error
		},
	},
	{
		ID: "20260826000001_add_transaction_composite_indexes",
		Migrate: func(tx *gorm.DB) error {
			statements := []string{
				`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_transactions_household_date ON transactions(household_id, transaction_date)`,
				`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_transactions_household_category ON transactions(household_id, category_id)`,
				`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_transactions_household_created_by ON transactions(household_id, created_by)`,
			}
			for _, stmt := range statements {
				if err := tx.Exec(stmt).Error; err != nil {
					return err
				}
			}
			return nil
		},
		Rollback: func(tx *gorm.DB) error {
			statements := []string{
				`DROP INDEX IF EXISTS idx_transactions_household_date`,
				`DROP INDEX IF EXISTS idx_transactions_household_category`,
				`DROP INDEX IF EXISTS idx_transactions_household_created_by`,
			}
			for _, stmt := range statements {
				if err := tx.Exec(stmt).Error; err != nil {
					return err
				}
			}
			return nil
		},
	},
	{
		ID: "20260827000001_fix_notification_logs_dedup_constraint",
		Migrate: func(tx *gorm.DB) error {
			// idx_notification_dedup (type, reference_id, period) tidak menyertakan user_id —
			// job kirim per-anggota akan gagal insert baris kedua untuk anggota kedua. db.md §7.
			_ = tx.Migrator().DropIndex(&entity.NotificationLog{}, "idx_notification_dedup")
			return tx.Exec(`
				CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS idx_uq_notification_dedup_user
				ON notification_logs(type, reference_id, period, user_id)
			`).Error
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Exec(`DROP INDEX IF EXISTS idx_uq_notification_dedup_user`).Error
		},
	},
	{
		ID: "20260828000001_widen_transaction_category_index_for_budget",
		Migrate: func(tx *gorm.DB) error {
			// Budget spent-query (Phase 08) filter household_id + category_id + transaction_date
			// range — index 2 kolom dari Phase 06 tidak cover transaction_date, db.md §3.7 minta 3 kolom.
			if err := tx.Exec(`DROP INDEX IF EXISTS idx_transactions_household_category`).Error; err != nil {
				return err
			}
			return tx.Exec(`
				CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_transactions_household_category_date
				ON transactions(household_id, category_id, transaction_date)
			`).Error
		},
		Rollback: func(tx *gorm.DB) error {
			if err := tx.Exec(`DROP INDEX IF EXISTS idx_transactions_household_category_date`).Error; err != nil {
				return err
			}
			return tx.Exec(`
				CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_transactions_household_category
				ON transactions(household_id, category_id)
			`).Error
		},
	},
	{
		ID: "20260831000001_add_email_verification_to_users",
		Migrate: func(tx *gorm.DB) error {
			// Tambah kolom verifikasi email (email_verified_at, verification_token, verification_token_exp).
			// AutoMigrate hanya menambah kolom yang belum ada; aman untuk tabel users yang sudah ada.
			if err := tx.AutoMigrate(&entity.User{}); err != nil {
				return err
			}
			// Tambah index untuk lookup token saat verifikasi.
			return tx.Exec(`
				CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_verification_token
				ON users(verification_token)
			`).Error
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Exec(`DROP INDEX IF EXISTS idx_users_verification_token`).Error
		},
	},
	{
		ID: "20260831000002_add_admin_fee_to_transactions",
		Migrate: func(tx *gorm.DB) error {
			return tx.Exec(`
				ALTER TABLE transactions
				ADD COLUMN IF NOT EXISTS admin_fee numeric(18,2) NOT NULL DEFAULT 0
			`).Error
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Exec(`ALTER TABLE transactions DROP COLUMN IF EXISTS admin_fee`).Error
		},
	},
	{
		ID: "20260831000003_fix_budget_period_unique_index_soft_delete",
		Migrate: func(tx *gorm.DB) error {
			_ = tx.Migrator().DropIndex(&entity.Budget{}, "idx_budget_period")
			return tx.Exec(`
				CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS idx_uq_budget_period_active
				ON budgets(household_id, category_id, period)
				WHERE deleted_at IS NULL
			`).Error
		},
		Rollback: func(tx *gorm.DB) error {
			_ = tx.Exec(`DROP INDEX IF EXISTS idx_uq_budget_period_active`).Error
			return tx.Migrator().CreateIndex(&entity.Budget{}, "idx_budget_period")
		},
	},
	{
		ID: "20260901000001_add_category_parent_id",
		Migrate: func(tx *gorm.DB) error {
			// Sub-kategori (parent_id nullable, self-referencing) — kolom baru, default NULL,
			// kategori existing tidak berubah perilaku sama sekali.
			if err := tx.AutoMigrate(&entity.Category{}); err != nil {
				return err
			}
			return tx.Exec(`
				CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_categories_household_parent
				ON categories(household_id, parent_id)
			`).Error
		},
		Rollback: func(tx *gorm.DB) error {
			if err := tx.Exec(`DROP INDEX IF EXISTS idx_categories_household_parent`).Error; err != nil {
				return err
			}
			return tx.Exec(`ALTER TABLE categories DROP COLUMN IF EXISTS parent_id`).Error
		},
	},
	{
		ID: "20260901000002_add_account_owner_type",
		Migrate: func(tx *gorm.DB) error {
			// Akun personal vs rumah tangga. Default 'household' untuk semua baris existing —
			// backfill eksplisit sebagai safety net kalau default GORM tidak kejalan saat ALTER,
			// lalu CHECK constraint menegakkan owner_user_id konsisten dengan owner_type.
			if err := tx.AutoMigrate(&entity.Account{}); err != nil {
				return err
			}
			if err := tx.Exec(`UPDATE accounts SET owner_type = 'household' WHERE owner_type IS NULL OR owner_type = ''`).Error; err != nil {
				return err
			}
			return tx.Exec(`
				ALTER TABLE accounts
				ADD CONSTRAINT chk_account_owner_consistency
				CHECK (
					(owner_type = 'personal' AND owner_user_id IS NOT NULL)
					OR (owner_type = 'household' AND owner_user_id IS NULL)
				)
			`).Error
		},
		Rollback: func(tx *gorm.DB) error {
			if err := tx.Exec(`ALTER TABLE accounts DROP CONSTRAINT IF EXISTS chk_account_owner_consistency`).Error; err != nil {
				return err
			}
			if err := tx.Exec(`ALTER TABLE accounts DROP COLUMN IF EXISTS owner_user_id`).Error; err != nil {
				return err
			}
			return tx.Exec(`ALTER TABLE accounts DROP COLUMN IF EXISTS owner_type`).Error
		},
	},
	{
		ID: "20260901000003_add_household_budget_cycle",
		Migrate: func(tx *gorm.DB) error {
			// Siklus budget custom (gajian tanggal 25, dll). Default 1 = kalender biasa,
			// household existing tidak berubah perilaku sama sekali.
			if err := tx.AutoMigrate(&entity.Household{}); err != nil {
				return err
			}
			return tx.Exec(`
				ALTER TABLE households
				ADD CONSTRAINT chk_household_budget_cycle_start_day
				CHECK (budget_cycle_start_day BETWEEN 1 AND 28)
			`).Error
		},
		Rollback: func(tx *gorm.DB) error {
			if err := tx.Exec(`ALTER TABLE households DROP CONSTRAINT IF EXISTS chk_household_budget_cycle_start_day`).Error; err != nil {
				return err
			}
			return tx.Exec(`ALTER TABLE households DROP COLUMN IF EXISTS budget_cycle_start_day`).Error
		},
	},
}
