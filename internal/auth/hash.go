package auth

import (
	"crypto/sha256"
	"fmt"
)

func Encrypt(password string) string {
	sum := sha256.Sum256([]byte(password))
	return fmt.Sprintf("%x", sum)
}
