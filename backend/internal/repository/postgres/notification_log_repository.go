package postgres

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"homeapp/internal/entity"
)

type NotificationLogRepository struct {
	db *gorm.DB
}

func NewNotificationLogRepository(db *gorm.DB) *NotificationLogRepository {
	return &NotificationLogRepository{db: db}
}

func (r *NotificationLogRepository) Exists(ctx context.Context, notifType entity.NotificationType, referenceID uuid.UUID, period string, userID uuid.UUID) (bool, error) {
	var count int64
	err := dbOrTx(ctx, r.db).WithContext(ctx).
		Model(&entity.NotificationLog{}).
		Where("type = ? AND reference_id = ? AND period = ? AND user_id = ?", notifType, referenceID, period, userID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *NotificationLogRepository) Create(ctx context.Context, log *entity.NotificationLog) error {
	return dbOrTx(ctx, r.db).WithContext(ctx).Create(log).Error
}
