package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

const csrfLabel = "jobflow-csrf-v1"

func CSRFToken(sessionToken string) string {
	mac := hmac.New(sha256.New, []byte(sessionToken))

	mac.Write([]byte(csrfLabel))
	sum := mac.Sum(nil)
	text := hex.EncodeToString(sum)

	return text
}
