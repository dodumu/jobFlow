package handlers

import (
	"jobFlow/database"
	"jobFlow/models"
	"jobFlow/utils"
	"net/http"
	"time"
)

var validRoles = map[string]bool{
	"company":    true,
	"individual": true,
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodGet {
		err := utils.RenderTemplate(w, "register.html", nil)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	username := r.FormValue("username")
	password := r.FormValue("password")
	firstName := r.FormValue("first_name")
	lastName := r.FormValue("last_name")
	email := r.FormValue("email")
	dob := r.FormValue("date_of_birth")
	role := r.FormValue("role")

	if role == "" || username == "" || password == "" || firstName == "" || lastName == "" || email == "" || dob == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if ok := validRoles[role]; !ok {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	hashPassword, err := utils.HashPassword(password)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	user := models.User{
		Username:     username,
		PasswordHash: hashPassword,
		FirstName:    firstName,
		LastName:     lastName,
		Email:        email,
		DOB:          dob,
		Role:         role,
	}
	id, err := database.CreateUser(user)
	if err != nil {
		http.Error(w, "create user: "+err.Error(), http.StatusInternalServerError)
		return
	}
	profile := models.UserProfile{
		UserID: id,
	}

	_, err = database.CreateUserProfile(profile)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if role == "company" {

		companyName := r.FormValue("company_name")
		description := r.FormValue("description")
		website := r.FormValue("website")
		location := r.FormValue("location")
		logo := r.FormValue("logo")
		if companyName == "" || description == "" || website == "" || location == "" {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		company := models.Company{
			UserID:      id,
			CompanyName: companyName,
			Description: description,
			Website:     website,
			Location:    location,
			Logo:        logo,
		}
		_, err := database.CreateCompany(company)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
	}
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		err := utils.RenderTemplate(w, "login.html", nil)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	username := r.FormValue("username")
	password := r.FormValue("password")

	if username == "" || password == "" {
		http.Error(w, "password or username cannot be empty", http.StatusBadRequest)
		return
	}
	user, err := database.GetUserByUsername(username)
	if err != nil {
		http.Error(w, "invalid username or password", http.StatusUnauthorized)
		return
	}
	ok := utils.CheckPassword(user.PasswordHash, password)
	if !ok {
		http.Error(w, "invalid username or password", http.StatusUnauthorized)
		return
	}
	token, err := utils.GenerateToken()
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	expires := time.Now().Add(24 * time.Hour)
	session := models.Session{
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: expires,
	}
	err = database.CreateSession(session)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    token,
		Expires:  expires,
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}
