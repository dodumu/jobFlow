package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

func GenerateVerificationCode() (string, error) {
	max := big.NewInt(100000000)

	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%08d", n.Int64()), nil
}
