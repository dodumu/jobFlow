package utils

import (
	"net/http"
)

func RedirectBack(w http.ResponseWriter, r *http.Request, fallback string) {
	referer := r.Header.Get("Referer")

	if referer != "" {
		http.Redirect(w, r, referer, http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, fallback, http.StatusSeeOther)
}
