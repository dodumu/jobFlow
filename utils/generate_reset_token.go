package utils

import (
	"crypto/rand"
	"encoding/hex"
)

func GenerateResetAuthorizationToken() (string, error) {
	tokenBytes := make([]byte, 32)

	_, err := rand.Read(tokenBytes)
	if err != nil {
		return "", err
	}

	token := hex.EncodeToString(tokenBytes)

	return token, nil
}
