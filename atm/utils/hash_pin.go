package utils

import "golang.org/x/crypto/bcrypt"

func ComparePins(hashedPin, Pin string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPin), []byte(Pin))
	return err
}

func HashPinBcrypt(input *string) error {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(*input), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	*input = string(hashedBytes)
	return nil
}
