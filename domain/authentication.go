package domain

import (
	"context"
	"time"
)

type AuthenticationUsecase interface {
	GenerateOtp(c context.Context, tel string) (string, error)
	SendOtp(c context.Context, tel string) error
	ValidateOtp(c context.Context, userOTP string, secretKey string) bool
}

type AuthenticationRepository interface {
	SaveOtpSecret(ctx context.Context, uuid string, secretKey string) (err error)
	GetOtpSecret(ctx context.Context, uuid string) (secretKey string, expiredAt time.Time, err error)
}
