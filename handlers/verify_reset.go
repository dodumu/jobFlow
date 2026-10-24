package handlers

import (
	"jobFlow/database"
	"jobFlow/utils"
	"log"
	"net/http"
	"net/url"
	"strings"
)

func VerifyResetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET, POST")
		RenderError(w, r, http.StatusMethodNotAllowed, "Unaccepted Method", "Method not allowed")
		return
	}

	w.Header().Set("Cache-Control", "no-store")

	if r.Method == http.MethodGet {
		data := map[string]any{
			"Title": "Verify Reset Code",
			"Sent":  r.URL.Query().Get("sent") == "1",
		}

		if err := utils.RenderTemplate(w, r, "verify_reset.html", data); err != nil {
			log.Printf("Rendering verification page: %v", err)
			RenderError(w, r, http.StatusInternalServerError, "Internal server error", "Something went wrong")
		}
		return
	}

	if err := r.ParseForm(); err != nil {
		RenderError(w, r, http.StatusBadRequest, "Bad request detected", "Invalid form submission")
		return
	}

	email := strings.TrimSpace(r.FormValue("email"))
	code := strings.TrimSpace(r.FormValue("code"))

	showError := func(message string, status int) {
		w.WriteHeader(status)

		data := map[string]any{
			"Title": "Verify Reset Code",
			"Error": message,
		}
		if err := utils.RenderTemplate(w, r, "verify_reset.html", data); err != nil {
			log.Printf("Rendereing verification err page: %v", err)
		}
	}
	if email == "" || len(code) != 8 {
		showError("Enter a valid eight-digit verification code.", http.StatusBadRequest)
		return
	}

	for _, digit := range code {
		if digit < '0' || digit > '9' {
			showError("Enter a valid eight-digit verification code.", http.StatusBadRequest)
			return
		}
	}
	user, err := database.GetUserByEmail(email)
	if err != nil {
		showError("The verification code is invalid or has expired.", http.StatusBadRequest)
		return
	}

	if err := database.VerifyPasswordResetCode(user.ID, code); err != nil {
		showError("The verification code is invalid or has expired.", http.StatusBadRequest)
		return
	}

	token, err := utils.GenerateResetAuthorizationToken()
	if err != nil {
		log.Printf("Generating reset authorization: %v", err)
		showError("We couldn't continue right now. Please try again later.", http.StatusInternalServerError)
		return
	}
	tokenHash := utils.HashResetAuthorizationToken(token)

	if err := database.SavePasswordResetAuthorization(user.ID, tokenHash); err != nil {
		log.Printf("Saving reset authorization: %v", err)
		showError("We couldn't continue right now. Please try again later.", http.StatusInternalServerError)
		return
	}

	// 5. Continue to the password-reset page.
	http.Redirect(
		w,
		r,
		"/reset-password?token="+url.QueryEscape(token),
		http.StatusSeeOther,
	)
}
