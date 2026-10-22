package handlers

import (
	"jobFlow/database"
	"jobFlow/utils"
	"log"
	"net/http"
)

func ForgotPasswordHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		data := map[string]any{
			"Title": "Forgot Password",
		}

		if err := utils.RenderTemplate(w, r, "forgot_password.html", data); err != nil {
			log.Printf("Rendering forgot password page: %v", err)
			RenderError(w, r, http.StatusInternalServerError, "Internal server error", "Problem parsing the the reset passowrd page")
		}
		return
	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			RenderError(w, r, http.StatusBadRequest, "Bad request", "Invalid Form Submisstion")
			return
		}

		email := r.FormValue("email")
		if email == "" {
			RenderError(w, r, http.StatusBadRequest, "Bad request", "Email is requred")
			return
		}
		UserID, err := database.GetUserByEmail(email)
		if err != nil {
			http.Redirect(w, r, "/forgot-password?sent=1", http.StatusSeeOther)
			return
		}
		code, err := utils.GenerateVerificationCode()
		if err != nil {
			log.Printf("Generating reset code: %v", err)
			RenderError(w, r, http.StatusInternalServerError, "Internal sever error", "Unable to process request")
			return
		}
		codeHash, err := utils.HashVerificationCode(UserID.ID, code)
		if err != nil {
			log.Printf("Hashing reset code: %v", err)
			RenderError(w, r, http.StatusInternalServerError, "internal server error", "Unable to process request")
			return
		}
		if err := database.SavePasswordResetAuthorization(UserID.ID, codeHash); err != nil {
			log.Printf("Saving reset code: %v", err)
			RenderError(w, r, http.StatusInternalServerError, "internal server error", "Unable to process request")
			return
		}
		if err := utils.SendVerificationEmail(email, code); err != nil {
			log.Printf("Sending reset email: %v", err)
			RenderError(w, r, http.StatusInternalServerError, "internal server error", "Unable to process request")
			return
		}
	}

}
