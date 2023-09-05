package domain

import "context"

type ExternalUsecase interface {
	SendSms(ctx context.Context)
}
