package handlers

import (
	"log"
	"net/http"
	"strings"

	"jobFlow/database"
	"jobFlow/utils"
)

func ResetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		w.Header().Set("Allow", "GET, POST")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Cache-Control", "no-store")

	var token string

	if r.Method == http.MethodGet {
		token = r.URL.Query().Get("token")
	} else {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Invalid form submission", http.StatusBadRequest)
			return
		}
		token = r.FormValue("token")
	}

	token = strings.TrimSpace(token)

	renderPage := func(message string, status int) {
		data := map[string]any{
			"Title": "Reset Password",
			"Token": token,
			"Error": message,
		}

		if status != http.StatusOK {
			w.WriteHeader(status)
		}

		if err := utils.RenderTemplate(w, r, "reset_password.html", data); err != nil {
			log.Printf("Rendering reset-password page: %v", err)
		}
	}

	if token == "" {
		renderPage("Your reset link is invalid. Please request a new code.", http.StatusBadRequest)
		return
	}

	if err := database.ValidatePasswordResetAuthorization(token); err != nil {
		renderPage("Your reset link is invalid or has expired. Please start again.", http.StatusBadRequest)
		return
	}

	// GET requests only display the form.
	if r.Method == http.MethodGet {
		renderPage("", http.StatusOK)
		return
	}

	password := r.FormValue("password")
	confirmPassword := r.FormValue("confirm_password")

	if password != confirmPassword {
		renderPage("The passwords do not match.", http.StatusBadRequest)
		return
	}

	if len(password) < 8 {
		renderPage("Your password must contain at least 8 characters.", http.StatusBadRequest)
		return
	}

	// Use JobFlow's existing bcrypt password-hashing function.
	passwordHash, err := utils.HashPassword(password)
	if err != nil {
		log.Printf("Hashing new password: %v", err)
		renderPage("We couldn't reset your password right now. Please try again.", http.StatusInternalServerError)
		return
	}

	// The database function atomically consumes the authorization,
	// updates the password, marks the code used, and revokes sessions.
	if err := database.ResetPassword(token, passwordHash); err != nil {
		log.Printf("Resetting password: %v", err)
		renderPage("Your reset link is no longer valid. Please request a new code.", http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, "/login?password_reset=1", http.StatusSeeOther)
}
