package repository

import (
	"context"

	"github.com/google/uuid"
	"homeapp/internal/entity"
)

type NotificationLogRepository interface {
	Exists(ctx context.Context, notifType entity.NotificationType, referenceID uuid.UUID, period string, userID uuid.UUID) (bool, error)
	Create(ctx context.Context, log *entity.NotificationLog) error
}
