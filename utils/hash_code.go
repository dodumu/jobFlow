package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
)

func HashVerificationCode(userID int, code string) (string, error) {
	secret := os.Getenv("RESET_CODE_SECRET")

	if secret == "" {
		return "", errors.New("RESET_CODE_SECRET is not configured")
	}

	h := hmac.New(sha256.New, []byte(secret))

	message := fmt.Sprintf("%d:%s", userID, code)

	if _, err := h.Write([]byte(message)); err != nil {
		return "", err
	}

	hash := h.Sum(nil)

	return hex.EncodeToString(hash), nil
}
