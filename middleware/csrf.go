package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
)

const csrfLabel = "jobflow-csrf-v1"
const sessionTokenKey contextKey = "sessionToken"
const csrfMultipartLimit = 6 << 20

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
		contentType := r.Header.Get("Content-Type")

		if strings.HasPrefix(contentType, "multipart/form-data") {
			r.Body = http.MaxBytesReader(w, r.Body, csrfMultipartLimit)

			if err := r.ParseMultipartForm(csrfMultipartLimit); err != nil {
				var tooLarge *http.MaxBytesError

				if errors.As(err, &tooLarge) {
					http.Error(w, "request too large", http.StatusRequestEntityTooLarge)
					return false
				}
				http.Error(w, "bad request", http.StatusBadRequest)
				return false
			}
		}
		supplied = r.PostFormValue("csrf_token")
	}
	expected := CSRFToken(sessionToken)
	if !hmac.Equal([]byte(supplied), []byte(expected)) {
		http.Error(w, "invalid or missing CSRF token", http.StatusForbidden)
		return false
	}
	return true

}
