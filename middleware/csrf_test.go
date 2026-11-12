package middleware

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestCSRFTokenIsDeterministic(t *testing.T) {
	first := CSRFToken("session-abc")
	second := CSRFToken("session-abc")

	if first != second {
		t.Errorf("expected the same token twice, got %q and %q", first, second)
	}
}

func TestCSRFTokenDiffersPerSession(t *testing.T) {
	a := CSRFToken("session-abc")
	b := CSRFToken("session-xyz")

	if a == b {
		t.Errorf("different sessions produced the same token %q", a)
	}
}

func TestVerifyCSRF(t *testing.T) {
	const session = "session-abc"
	validToken := CSRFToken(session)

	tests := []struct {
		name     string
		method   string
		header   string // value for the X-CSRF-Token header ("" = not sent)
		formTok  string // value for the csrf_token form field ("" = not sent)
		wantOK   bool
		wantCode int // expected status when rejected (0 = don't check)
	}{
		{"GET needs no token", http.MethodGet, "", "", true, 0},
		{"POST with valid header", http.MethodPost, validToken, "", true, 0},
		{"POST with valid form field", http.MethodPost, "", validToken, true, 0},
		{"POST with no token", http.MethodPost, "", "", false, http.StatusForbidden},
		{"POST with wrong header", http.MethodPost, "forged", "", false, http.StatusForbidden},
		{"POST with wrong form field", http.MethodPost, "", "forged", false, http.StatusForbidden},
		{"POST with token from another session", http.MethodPost, CSRFToken("session-xyz"), "", false, http.StatusForbidden},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			form := url.Values{}
			if tc.formTok != "" {
				form.Set("csrf_token", tc.formTok)
			}

			req := httptest.NewRequest(tc.method, "/logout", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			if tc.header != "" {
				req.Header.Set("X-CSRF-Token", tc.header)
			}

			rec := httptest.NewRecorder()

			got := verifyCSRF(rec, req, session)

			if got != tc.wantOK {
				t.Errorf("verifyCSRF returned %v, want %v", got, tc.wantOK)
			}
			if tc.wantCode != 0 && rec.Code != tc.wantCode {
				t.Errorf("status = %d, want %d", rec.Code, tc.wantCode)
			}
		})
	}
}

func buildMultipartBody(t *testing.T, token string, fileSize int) (*bytes.Buffer, string) {
	t.Helper()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	if token != "" {
		if err := writer.WriteField("csrf_token", token); err != nil {
			t.Fatalf("writing token field: %v", err)
		}
	}

	if fileSize > 0 {
		part, err := writer.CreateFormFile("avatar", "avatar.png")
		if err != nil {
			t.Fatalf("creating file part: %v", err)
		}
		if _, err := part.Write(bytes.Repeat([]byte("a"), fileSize)); err != nil {
			t.Fatalf("writing file part: %v", err)
		}
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("closing multipart writer: %v", err)
	}

	return body, writer.FormDataContentType()
}

func TestVerifyCSRFMultipart(t *testing.T) {
	const session = "session-abc"
	valid := CSRFToken(session)

	tests := []struct {
		name     string
		token    string
		fileSize int
		wantOK   bool
		wantCode int
	}{
		{"valid token, small file", valid, 1024, true, 0},
		{"missing token", "", 1024, false, http.StatusForbidden},
		{"forged token", "forged", 1024, false, http.StatusForbidden},
		{"valid token but body over 6 MiB", valid, 7 << 20, false, http.StatusRequestEntityTooLarge},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			body, contentType := buildMultipartBody(t, tc.token, tc.fileSize)

			req := httptest.NewRequest(http.MethodPost, "/profile", body)
			req.Header.Set("Content-Type", contentType)

			rec := httptest.NewRecorder()

			got := verifyCSRF(rec, req, session)

			if got != tc.wantOK {
				t.Errorf("verifyCSRF returned %v, want %v", got, tc.wantOK)
			}
			if tc.wantCode != 0 && rec.Code != tc.wantCode {
				t.Errorf("status = %d, want %d", rec.Code, tc.wantCode)
			}
		})
	}
}
