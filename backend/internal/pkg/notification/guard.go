package notification

import (
	"context"
	"time"

	"github.com/google/uuid"
	"homeapp/internal/entity"
	"homeapp/internal/repository"
)

// Guard menyediakan dedup untuk job pengirim notifikasi (Phase 08/10), supaya job
// spesifik tidak perlu tahu detail SQL. Restart-safe: dedup dicek per anggota, bukan
// per job run, jadi loop yang terhenti di tengah jalan aman dilanjutkan.
type Guard struct {
	repo repository.NotificationLogRepository
}

func NewGuard(repo repository.NotificationLogRepository) *Guard {
	return &Guard{repo: repo}
}

func (g *Guard) AlreadySent(ctx context.Context, notifType entity.NotificationType, referenceID uuid.UUID, period string, userID uuid.UUID) (bool, error) {
	return g.repo.Exists(ctx, notifType, referenceID, period, userID)
}

// MarkSent dipanggil SETELAH email sukses terkirim, tidak sebelumnya — kalau kirim
// gagal, job berikutnya akan retry karena belum ada log (lebih aman daripada gagal
// permanen; risiko kirim dobel diterima untuk notifikasi info).
func (g *Guard) MarkSent(ctx context.Context, householdID uuid.UUID, notifType entity.NotificationType, referenceID uuid.UUID, period string, userID uuid.UUID) error {
	return g.repo.Create(ctx, &entity.NotificationLog{
		HouseholdID: householdID,
		UserID:      userID,
		Type:        notifType,
		ReferenceID: referenceID,
		Period:      period,
		SentAt:      time.Now(),
	})
}
