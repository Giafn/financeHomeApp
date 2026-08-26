package job

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"homeapp/internal/entity"
	"homeapp/internal/pkg/mailer"
	"homeapp/internal/pkg/notification"
	"homeapp/internal/repository"
	"homeapp/internal/usecase"
)

// testGuardReferenceID tetap (fixed) supaya AlreadySent/MarkSent dedup bisa diverifikasi
// dengan memanggil job ini 2x berturut-turut — panggilan kedua harus terdeteksi sudah terkirim.
var testGuardReferenceID = uuid.MustParse("00000000-0000-0000-0000-000000000001")

// Period kolom varchar(7), konvensi "YYYY-MM" — dihitung saat job jalan supaya
// dedup tetap valid per bulan seperti job produksi (budget alert/bill reminder).
func testGuardPeriod() string {
	return time.Now().Format("2006-01")
}

// RegisterJobs mendaftarkan semua job aplikasi ke registry. Dipanggil dari cmd/worker
// (untuk dijadwalkan cron) dan cmd/api (untuk manual trigger lewat endpoint QA) supaya
// definisi job tidak dobel di dua tempat.
//
func RegisterJobs(
	registry *Registry,
	m mailer.Mailer,
	guard *notification.Guard,
	testRecipient string,
	budgetRepo repository.BudgetRepository,
	householdRepo repository.HouseholdRepository,
	billRepo repository.BillRepository,
	billPeriodRepo repository.BillPeriodRepository,
) {
	registry.Add("dummy-heartbeat", func(ctx context.Context) (map[string]interface{}, error) {
		now := time.Now().Format(time.RFC3339)
		log.Printf("dummy-heartbeat jalan pada %s", now)
		return map[string]interface{}{"ran_at": now}, nil
	})

	registry.Add("test-notification-guard", func(ctx context.Context) (map[string]interface{}, error) {
		if testRecipient == "" {
			return nil, fmt.Errorf("SMTP_USER belum diset, tidak ada alamat tujuan test email")
		}

		userID := testGuardReferenceID // pakai ID yang sama sebagai stand-in user untuk test infra

		alreadySent, err := guard.AlreadySent(ctx, entity.NotificationTest, testGuardReferenceID, testGuardPeriod(), userID)
		if err != nil {
			return nil, err
		}
		if alreadySent {
			log.Printf("test-notification-guard: sudah pernah terkirim, skip (dedup bekerja)")
			return map[string]interface{}{"sent": false, "reason": "already_sent"}, nil
		}

		if err := m.Send([]string{testRecipient}, "Test Notifikasi Infra Phase 07", "Ini email test dari job-infra Phase 07."); err != nil {
			return nil, err
		}

		if err := guard.MarkSent(ctx, uuid.Nil, entity.NotificationTest, testGuardReferenceID, testGuardPeriod(), userID); err != nil {
			return nil, err
		}

		log.Printf("test-notification-guard: email terkirim ke %s", testRecipient)
		return map[string]interface{}{"sent": true}, nil
	})

	// budget-auto-copy jalan tanggal 1 tiap bulan: untuk tiap household yang punya budget
	// bulan lalu, copy baris yang BELUM ada di bulan ini (kategori yang sudah di-set manual
	// bulan ini tidak ditimpa). Idempotent — cek "belum ada baris bulan ini" sebelum insert,
	// aman dijalankan berkali-kali kalau worker restart di tengah loop.
	registry.Add("budget-auto-copy", func(ctx context.Context) (map[string]interface{}, error) {
		now := time.Now()
		currentPeriod := now.Format("2006-01")
		previousPeriod := now.AddDate(0, -1, 0).Format("2006-01")

		householdIDs, err := budgetRepo.ListHouseholdIDsWithBudgetForPeriod(ctx, previousPeriod)
		if err != nil {
			return nil, err
		}

		copied := 0
		for _, hhID := range householdIDs {
			prevBudgets, err := budgetRepo.ListRawByPeriod(ctx, hhID, previousPeriod)
			if err != nil {
				return nil, err
			}
			currentBudgets, err := budgetRepo.ListRawByPeriod(ctx, hhID, currentPeriod)
			if err != nil {
				return nil, err
			}

			existing := make(map[uuid.UUID]bool, len(currentBudgets))
			for _, b := range currentBudgets {
				existing[b.CategoryID] = true
			}

			for _, pb := range prevBudgets {
				if existing[pb.CategoryID] {
					continue
				}
				newBudget := &entity.Budget{
					HouseholdID: hhID,
					CategoryID:  pb.CategoryID,
					Period:      currentPeriod,
					Amount:      pb.Amount,
					CreatedBy:   pb.CreatedBy,
				}
				if err := budgetRepo.Create(ctx, newBudget); err != nil {
					return nil, err
				}
				copied++
			}
		}

		log.Printf("budget-auto-copy: %d baris budget disalin ke periode %s", copied, currentPeriod)
		return map[string]interface{}{"copied": copied, "period": currentPeriod}, nil
	})

	// budget-alert-check jalan harian: untuk tiap household+kategori yang spent-nya sudah
	// >=80% dari budget bulan ini, kirim 1 email per anggota household — guard dedup
	// (per user, bukan per job run) memastikan alert cuma terkirim sekali per kategori/bulan
	// meski expense terus naik dan job jalan lagi besok.
	registry.Add("budget-alert-check", func(ctx context.Context) (map[string]interface{}, error) {
		period := time.Now().Format("2006-01")

		householdIDs, err := budgetRepo.ListHouseholdIDsWithBudgetForPeriod(ctx, period)
		if err != nil {
			return nil, err
		}

		alertsSent := 0
		for _, hhID := range householdIDs {
			budgets, err := budgetRepo.ListByPeriod(ctx, hhID, period)
			if err != nil {
				return nil, err
			}

			for _, b := range budgets {
				if b.Amount <= 0 {
					continue
				}
				pct := b.Spent / b.Amount * 100
				if pct < 80 {
					continue
				}

				members, err := householdRepo.FindMembersByHouseholdID(ctx, hhID)
				if err != nil {
					return nil, err
				}

				for _, member := range members {
					alreadySent, err := guard.AlreadySent(ctx, entity.NotificationBudgetAlert, b.CategoryID, period, member.UserID)
					if err != nil {
						return nil, err
					}
					if alreadySent {
						continue
					}

					subject := fmt.Sprintf("Peringatan Budget: %s sudah %.0f%%", b.CategoryName, pct)
					body := fmt.Sprintf("Budget kategori %s bulan %s sudah terpakai %.0f%% (Rp%.0f dari Rp%.0f).",
						b.CategoryName, period, pct, b.Spent, b.Amount)

					if err := m.Send([]string{member.User.Email}, subject, body); err != nil {
						log.Printf("budget-alert-check: gagal kirim ke %s: %v", member.User.Email, err)
						continue
					}

					if err := guard.MarkSent(ctx, hhID, entity.NotificationBudgetAlert, b.CategoryID, period, member.UserID); err != nil {
						return nil, err
					}
					alertsSent++
				}
			}
		}

		log.Printf("budget-alert-check: %d alert terkirim untuk periode %s", alertsSent, period)
		return map[string]interface{}{"alerts_sent": alertsSent, "period": period}, nil
	})

	// bill-period-generator jalan tanggal 1 tiap bulan: cuma untuk bill indefinite (tanpa
	// end_period, spec Phase 10 §2) — bill dengan rentang tetap sudah generate semua
	// periodenya sinkron saat dibuat. Idempotent: cek dulu periode target belum ada
	// sebelum insert (bukan cuma andalkan unique constraint).
	registry.Add("bill-period-generator", func(ctx context.Context) (map[string]interface{}, error) {
		bills, err := billRepo.ListIndefiniteActive(ctx)
		if err != nil {
			return nil, err
		}

		generated := 0
		for _, bill := range bills {
			latest, err := billPeriodRepo.LatestByBillID(ctx, bill.ID)
			if err != nil {
				return nil, err
			}

			latestTime, err := time.Parse("2006-01", latest.Period)
			if err != nil {
				return nil, err
			}
			nextPeriod := latestTime.AddDate(0, 1, 0).Format("2006-01")

			if _, err := billPeriodRepo.FindByBillAndPeriod(ctx, bill.ID, nextPeriod); err == nil {
				continue // sudah ada, idempotent skip
			}

			dueDate, err := usecase.DueDateForPeriod(nextPeriod, bill.DueDay)
			if err != nil {
				return nil, err
			}

			if err := billPeriodRepo.Create(ctx, &entity.BillPeriod{
				BillID:  bill.ID,
				Period:  nextPeriod,
				DueDate: dueDate,
				Status:  entity.BillPeriodUpcoming,
			}); err != nil {
				return nil, err
			}
			generated++
		}

		log.Printf("bill-period-generator: %d periode baru digenerate", generated)
		return map[string]interface{}{"generated": generated}, nil
	})

	// bill-reminder-check jalan harian: kirim 1 email per anggota household untuk tiap
	// bill_period upcoming yang due_date-nya sudah dalam rentang reminder_days_before —
	// dedup per user via notification_logs (type=bill_reminder, reference_id=bill_period.id).
	registry.Add("bill-reminder-check", func(ctx context.Context) (map[string]interface{}, error) {
		today := time.Now().Truncate(24 * time.Hour)

		due, err := billPeriodRepo.ListDueForReminder(ctx, today)
		if err != nil {
			return nil, err
		}

		remindersSent := 0
		for _, bp := range due {
			members, err := householdRepo.FindMembersByHouseholdID(ctx, bp.HouseholdID)
			if err != nil {
				return nil, err
			}

			for _, member := range members {
				alreadySent, err := guard.AlreadySent(ctx, entity.NotificationBillReminder, bp.ID, bp.Period, member.UserID)
				if err != nil {
					return nil, err
				}
				if alreadySent {
					continue
				}

				subject := fmt.Sprintf("Pengingat Tagihan: %s jatuh tempo %s", bp.BillName, bp.DueDate.Format("2 Jan 2006"))
				body := fmt.Sprintf("Tagihan %s periode %s jatuh tempo pada %s. Jangan sampai terlambat!",
					bp.BillName, bp.Period, bp.DueDate.Format("2 January 2006"))

				if err := m.Send([]string{member.User.Email}, subject, body); err != nil {
					log.Printf("bill-reminder-check: gagal kirim ke %s: %v", member.User.Email, err)
					continue
				}

				if err := guard.MarkSent(ctx, bp.HouseholdID, entity.NotificationBillReminder, bp.ID, bp.Period, member.UserID); err != nil {
					return nil, err
				}
				remindersSent++
			}
		}

		log.Printf("bill-reminder-check: %d reminder terkirim", remindersSent)
		return map[string]interface{}{"reminders_sent": remindersSent}, nil
	})

	// bill-period-overdue-check jalan harian: upcoming yang due_date-nya sudah lewat -> overdue.
	registry.Add("bill-period-overdue-check", func(ctx context.Context) (map[string]interface{}, error) {
		today := time.Now().Truncate(24 * time.Hour)

		overdue, err := billPeriodRepo.ListOverdue(ctx, today)
		if err != nil {
			return nil, err
		}

		for _, bp := range overdue {
			if err := billPeriodRepo.UpdateStatus(ctx, bp.ID, entity.BillPeriodOverdue); err != nil {
				return nil, err
			}
		}

		log.Printf("bill-period-overdue-check: %d periode diubah jadi overdue", len(overdue))
		return map[string]interface{}{"marked_overdue": len(overdue)}, nil
	})
}
