package entity

import (
	"time"

	"github.com/google/uuid"
)

type NotificationType string

const (
	NotificationBillReminder NotificationType = "bill_reminder"
	NotificationBudgetAlert  NotificationType = "budget_alert"
	// NotificationTest dipakai job "test-notification-guard" (Phase 07) untuk
	// membuktikan mailer+dedup guard bekerja end-to-end tanpa menyentuh tipe produksi.
	NotificationTest NotificationType = "test"
)

// NotificationLog mencatat email yang sudah terkirim untuk mencegah duplikasi
// (unik per type+reference_id+period).
type NotificationLog struct {
	BaseModel
	HouseholdID uuid.UUID        `gorm:"type:uuid;not null;index" json:"household_id"`
	UserID      uuid.UUID        `gorm:"type:uuid;not null;index" json:"user_id"`
	Type        NotificationType `gorm:"type:varchar(30);not null;uniqueIndex:idx_notification_dedup" json:"type"`
	ReferenceID uuid.UUID        `gorm:"type:uuid;not null;uniqueIndex:idx_notification_dedup" json:"reference_id"`
	Period      string           `gorm:"type:varchar(7);not null;uniqueIndex:idx_notification_dedup" json:"period"`
	SentAt      time.Time        `json:"sent_at"`
}

func (NotificationLog) TableName() string { return "notification_logs" }
