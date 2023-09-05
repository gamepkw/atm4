package domain

import (
	"context"
)

type NotificationUsecase interface {
	SendTransactionNotification(ctx context.Context)
}
