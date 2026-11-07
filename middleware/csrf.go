package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
)

const csrfLabel = "jobflow-csrf-v1"
const sessionTokenKey contextKey = "sessionToken"

func CSRFToken(sessionToken string) string {
	mac := hmac.New(sha256.New, []byte(sessionToken))

	mac.Write([]byte(csrfLabel))
	sum := mac.Sum(nil)
	text := hex.EncodeToString(sum)

	return text
}

func CSRFTokenFromRequest(r *http.Request) string {
	sessionToken, ok := r.Context().Value(sessionTokenKey).(string)
	if !ok || sessionToken == "" {
		return ""
	}
	return CSRFToken(sessionToken)
}

func verifyCSRF(w http.ResponseWriter, r *http.Request, sessionToken string) bool {
	switch r.Method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return true
	}

	supplied := r.Header.Get("X-CSRF-Token")
	if supplied == "" {
		supplied = r.PostFormValue("csrf_token")
	}
	expected := CSRFToken(sessionToken)
	if !hmac.Equal([]byte(supplied), []byte(expected)) {
		http.Error(w, "invalid or missing CSRF token", http.StatusForbidden)
		return false
	}
	return true

}
