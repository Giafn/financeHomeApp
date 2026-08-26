// Binary worker TERPISAH dari cmd/api — menjalankan scheduler cron saja, tidak
// membuka port HTTP. Restart/scaling API tidak mengganggu jadwal job, dan sebaliknya.
//
// Penggunaan:
//
//	go run ./cmd/worker
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"homeapp/internal/config"
	"homeapp/internal/database"
	"homeapp/internal/job"
	"homeapp/internal/pkg/mailer"
	"homeapp/internal/pkg/notification"
	"homeapp/internal/pkg/scheduler"
	postgresRepo "homeapp/internal/repository/postgres"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("gagal load konfigurasi: %v", err)
	}

	db, err := database.NewPostgresConnection(cfg)
	if err != nil {
		log.Fatalf("gagal konek ke database: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("gagal ambil koneksi sql.DB: %v", err)
	}
	defer sqlDB.Close()

	notificationLogRepo := postgresRepo.NewNotificationLogRepository(db)
	notificationGuard := notification.NewGuard(notificationLogRepo)
	mailerClient := mailer.NewSMTPMailer(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPassword, cfg.SMTPFrom)
	budgetRepo := postgresRepo.NewBudgetRepository(db)
	householdRepo := postgresRepo.NewHouseholdRepository(db)
	billRepo := postgresRepo.NewBillRepository(db)
	billPeriodRepo := postgresRepo.NewBillPeriodRepository(db)

	registry := job.NewRegistry()
	job.RegisterJobs(registry, mailerClient, notificationGuard, cfg.SMTPUser, budgetRepo, householdRepo, billRepo, billPeriodRepo)

	sched := scheduler.New()

	runJob := func(name string) func() {
		return func() {
			if _, err := registry.Run(context.Background(), name); err != nil {
				log.Printf("%s error: %v", name, err)
			}
		}
	}

	// Interval pendek untuk verifikasi cepat sesuai DoD Phase 07.
	if err := sched.Register("dummy-heartbeat", "* * * * *", runJob("dummy-heartbeat")); err != nil {
		log.Fatalf("gagal daftar job dummy-heartbeat: %v", err)
	}

	// budget-auto-copy: tanggal 1 tiap bulan jam 00:05 (spec Phase 08 poin 3).
	if err := sched.Register("budget-auto-copy", "5 0 1 * *", runJob("budget-auto-copy")); err != nil {
		log.Fatalf("gagal daftar job budget-auto-copy: %v", err)
	}

	// budget-alert-check: harian jam 08:00 (spec Phase 08 poin 4).
	if err := sched.Register("budget-alert-check", "0 8 * * *", runJob("budget-alert-check")); err != nil {
		log.Fatalf("gagal daftar job budget-alert-check: %v", err)
	}

	// bill-period-generator: tanggal 1 tiap bulan jam 00:10 (spec Phase 10 poin 2).
	if err := sched.Register("bill-period-generator", "10 0 1 * *", runJob("bill-period-generator")); err != nil {
		log.Fatalf("gagal daftar job bill-period-generator: %v", err)
	}

	// bill-reminder-check: harian jam 08:00 (spec Phase 10 poin 3).
	if err := sched.Register("bill-reminder-check", "0 8 * * *", runJob("bill-reminder-check")); err != nil {
		log.Fatalf("gagal daftar job bill-reminder-check: %v", err)
	}

	// bill-period-overdue-check: harian jam 00:15 (spec Phase 10 poin 4).
	if err := sched.Register("bill-period-overdue-check", "15 0 * * *", runJob("bill-period-overdue-check")); err != nil {
		log.Fatalf("gagal daftar job bill-period-overdue-check: %v", err)
	}

	sched.Start()
	log.Println("🕒 worker berjalan, scheduler aktif")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("worker mematikan scheduler...")
	sched.Stop()
}
