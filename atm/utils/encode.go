package utils

import (
	"crypto/sha256"
	"fmt"
)

func EncodeBase64(tel *string) error {
	hash := sha256.New()
	hash.Write([]byte(*tel))
	hashedTel := hash.Sum(nil)
	*tel = fmt.Sprintf("%x", hashedTel)
	return nil
}

func HashSha256(tel *string) error {
	hash := sha256.New()
	hash.Write([]byte(*tel))
	hashedTel := hash.Sum(nil)
	*tel = fmt.Sprintf("%x", hashedTel)
	return nil
}
