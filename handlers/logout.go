package handlers

import (
	"jobFlow/database"
	"log"
	"net/http"
)

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		RenderError(
			w,
			http.StatusMethodNotAllowed,
			"Method Not Allowed",
			"The requested method is not allowed for this action.",
		)
		return
	}

	cookie, err := r.Cookie("session_token")
	if err != nil {
		// No valid session cookie means there is effectively
		// nothing to log out.
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	token := cookie.Value

	if err := database.DeleteSession(token); err != nil {
		log.Printf("LogoutHandler DeleteSession error: %v", err)

		RenderError(
			w,
			http.StatusInternalServerError,
			"Something Went Wrong",
			"We couldn't log you out right now.",
		)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
