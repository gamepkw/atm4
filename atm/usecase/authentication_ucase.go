package usecase

import (
	"context"
	"encoding/base32"
	"fmt"
	"math/rand"
	"time"

	"main/atm/utils"
	"main/domain"
	producer "main/kafka/producer"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"github.com/spf13/viper"
)

type authenticationUsecase struct {
	authenticationRepo domain.AuthenticationRepository
	contextTimeout     time.Duration
}

// NewAccountUsecase will create new an accountUsecase object representation of domain.AccountUsecase interface
func NewAuthenticationUsecase(auth domain.AuthenticationRepository, timeout time.Duration) domain.AuthenticationUsecase {
	return &authenticationUsecase{
		authenticationRepo: auth,
		contextTimeout:     timeout,
	}
}

func (auth *authenticationUsecase) GenerateOtp(c context.Context, tel string) (string, error) {
	ctx, cancel := context.WithTimeout(c, auth.contextTimeout)
	defer cancel()
	secretKey, err := generateRandomSecretKey()
	if err != nil {
		return "", err
	}

	validateOpts := totp.ValidateOpts{
		Period:    180,
		Skew:      1,
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
	}

	otp, err := totp.GenerateCodeCustom(secretKey, time.Now(), validateOpts)
	if err != nil {
		return "", err
	}

	auth.saveOtpSecret(ctx, tel, secretKey)

	return otp, nil
}

func (auth *authenticationUsecase) SendOtp(c context.Context, tel string) error {
	topic := "sms"
	brokerAddress := viper.GetString("kafka.broker_address")
	ctx, cancel := context.WithTimeout(c, auth.contextTimeout)
	defer cancel()
	otp, err := auth.GenerateOtp(ctx, tel)
	if err != nil {
		return err
	}

	producer.RunKafkaProducer(brokerAddress, topic, otp)
	return nil
}

func (auth *authenticationUsecase) ValidateOtp(c context.Context, tel string, otpUser string) bool {
	ctx, cancel := context.WithTimeout(c, auth.contextTimeout)
	defer cancel()

	secretKey, expiredAt, err := auth.getSecretKeyByUUID(ctx, tel)

	if secretKey == "" {
		fmt.Println("key error")
		return false
	}

	if err != nil {
		fmt.Println("OTP error")
		return false
	}

	if expiredAt.Before(time.Now()) {
		fmt.Println("OTP expired")
		return false
	}

	validateOpts := totp.ValidateOpts{
		Period:    180,
		Skew:      1,
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
	}

	valid, err := totp.ValidateCustom(otpUser, secretKey, time.Now(), validateOpts)
	if err != nil {
		fmt.Println("OTP error")
		return false
	}

	return valid
}

func (auth *authenticationUsecase) getSecretKeyByUUID(c context.Context, tel string) (string, time.Time, error) {
	ctx, cancel := context.WithTimeout(c, auth.contextTimeout)
	defer cancel()

	utils.EncodeBase64(&tel)

	secretKey, expiredAt, _ := auth.authenticationRepo.GetOtpSecret(ctx, tel)

	return secretKey, expiredAt, nil
}

func generateRandomSecretKey() (string, error) {
	key := make([]byte, 16) // Generate a 16-byte random key
	_, err := rand.Read(key)
	if err != nil {
		return "", err
	}
	return base32.StdEncoding.EncodeToString(key), nil
}

func (auth *authenticationUsecase) saveOtpSecret(c context.Context, uuid string, secretKey string) (err error) {
	ctx, cancel := context.WithTimeout(c, auth.contextTimeout)
	defer cancel()

	if err = utils.EncodeBase64(&uuid); err != nil {
		return
	}

	if err = auth.authenticationRepo.SaveOtpSecret(ctx, uuid, secretKey); err != nil {
		return err
	}

	return nil
}
