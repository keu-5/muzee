package util

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// GenerateVerificationCode generates a 6-digit verification code
func GenerateVerificationCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}
