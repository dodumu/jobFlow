package handlers

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"jobFlow/constants"
	"jobFlow/database"
)

func TestRegisterHandler(t *testing.T) {
	tests := []struct {
		name       string
		change     func(form url.Values)
		wantStatus int
	}{
		{"valid individual", func(form url.Values) {}, http.StatusSeeOther},
		{"passwords differ", func(form url.Values) {
			form.Set("confirm_password", "different7pass")
		}, http.StatusBadRequest},
		{"invalid email", func(form url.Values) {
			form.Set("email", "not-an-email")
		}, http.StatusBadRequest},
		{"weak password", func(form url.Values) {
			form.Set("password", "abc")
			form.Set("confirm_password", "abc")
		}, http.StatusBadRequest},
		{"missing username", func(form url.Values) {
			form.Del("username")
		}, http.StatusBadRequest},
		{"unknown role", func(form url.Values) {
			form.Set("role", "admin")
		}, http.StatusBadRequest},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			setupTestDB(t)

			form := validRegistration()
			tc.change(form)

			rec := postForm(RegisterHandler, "/register", form)

			if rec.Code != tc.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tc.wantStatus)
			}

			_, err := database.GetUserByUsername("david")
			accountExists := err == nil
			wantAccount := tc.wantStatus == http.StatusSeeOther

			if accountExists != wantAccount {
				t.Errorf("account exists = %v, want %v", accountExists, wantAccount)
			}
		})
	}
}

func TestLoginSuccess(t *testing.T) {
	setupTestDB(t)
	registerTestUser(t)

	rec := postForm(LoginHandler, "/login", url.Values{
		"username": {"david"},
		"password": {"correct7horse"},
	})

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusSeeOther)
	}
	if got := rec.Header().Get("Location"); got != "/dashboard" {
		t.Errorf("redirected to %q, want /dashboard", got)
	}

	var sessionValue string
	for _, cookie := range rec.Result().Cookies() {
		if cookie.Name == constants.SessionCookieName {
			sessionValue = cookie.Value
		}
	}
	if sessionValue == "" {
		t.Error("expected a session cookie, got none")
	}
}

func TestLoginFailureMessages(t *testing.T) {
	setupTestDB(t)
	registerTestUser(t)

	tests := []struct {
		name     string
		username string
		password string
	}{
		{"wrong password", "david", "wrong-password1"},
		{"unknown username", "nobody", "correct7horse"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rec := postForm(LoginHandler, "/login", url.Values{
				"username": {tc.username},
				"password": {tc.password},
			})

			body := rec.Body.String()

			// Both cases must give the same message, so attackers cannot
			// learn which usernames exist.
			if !strings.Contains(body, "Invalid username or password.") {
				t.Errorf("body does not contain the generic failure message")
			}
			for _, cookie := range rec.Result().Cookies() {
				if cookie.Name == constants.SessionCookieName {
					t.Errorf("a failed login must not set a session cookie")
				}
			}
		})
	}
}

func TestLoginLockout(t *testing.T) {
	setupTestDB(t)
	registerTestUser(t)

	wrong := url.Values{
		"username": {"david"},
		"password": {"wrong-password1"},
	}

	for i := 0; i < constants.MaxLoginAttempts; i++ {
		postForm(LoginHandler, "/login", wrong)
	}

	// Even the correct password is refused while the account is locked.
	rec := postForm(LoginHandler, "/login", url.Values{
		"username": {"david"},
		"password": {"correct7horse"},
	})

	if !strings.Contains(rec.Body.String(), "Too many login attempts") {
		t.Error("expected the lockout message after repeated failures")
	}
	for _, cookie := range rec.Result().Cookies() {
		if cookie.Name == constants.SessionCookieName {
			t.Error("a locked-out login must not set a session cookie")
		}
	}
}