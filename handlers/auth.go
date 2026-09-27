package handlers

import (
	"jobFlow/constants"
	"jobFlow/database"
	"jobFlow/models"
	"jobFlow/utils"
	"log"
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
			log.Printf("RegisterHandler template error: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	if r.Method != http.MethodPost {
		RenderError(
			w,
			http.StatusMethodNotAllowed,
			"Method Not Allowed",
			"The requested method is not allowed on this page.",
		)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")
	firstName := r.FormValue("first_name")
	lastName := r.FormValue("last_name")
	email := r.FormValue("email")
	dob := r.FormValue("date_of_birth")
	role := r.FormValue("role")

	if role == "" ||
		username == "" ||
		password == "" ||
		firstName == "" ||
		lastName == "" ||
		email == "" ||
		dob == "" {

		RenderError(
			w,
			http.StatusBadRequest,
			"Missing Information",
			"Please complete all required registration fields.",
		)
		return
	}

	if !validRoles[role] {
		RenderError(
			w,
			http.StatusBadRequest,
			"Invalid Account Type",
			"Please select a valid account type.",
		)
		return
	}

	hashPassword, err := utils.HashPassword(password)
	if err != nil {
		log.Printf("RegisterHandler HashPassword error: %v", err)
		RenderError(
			w,
			http.StatusInternalServerError,
			"Something Went Wrong",
			"We couldn't create your account right now.",
		)
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
		log.Printf("RegisterHandler CreateUser error: %v", err)
		RenderError(
			w,
			http.StatusInternalServerError,
			"Registration Failed",
			"We couldn't create your account right now.",
		)
		return
	}

	profile := models.UserProfile{
		UserID: id,
	}

	_, err = database.CreateUserProfile(profile)
	if err != nil {
		log.Printf("RegisterHandler CreateUserProfile error: %v", err)
		RenderError(
			w,
			http.StatusInternalServerError,
			"Something Went Wrong",
			"We couldn't finish setting up your profile.",
		)
		return
	}

	if role == constants.RoleCompany {
		industry := r.FormValue("industry")
		companySize := r.FormValue("company_size")
		companyName := r.FormValue("company_name")
		description := r.FormValue("description")
		website := r.FormValue("website")
		location := r.FormValue("location")
		logo := r.FormValue("logo")

		if companyName == "" ||
			description == "" ||
			website == "" ||
			location == "" ||
			industry == "" ||
			companySize == "" {

			RenderError(
				w,
				http.StatusBadRequest,
				"Missing Company Information",
				"Please complete all required company fields.",
			)
			return
		}

		company := models.Company{
			UserID:      id,
			CompanyName: companyName,
			Description: description,
			Website:     website,
			Location:    location,
			Logo:        logo,
			Industry:    industry,
			FoundedYear: 0,
			CompanySize: companySize,
		}

		_, err := database.CreateCompany(company)
		if err != nil {
			log.Printf("RegisterHandler CreateCompany error: %v", err)
			RenderError(
				w,
				http.StatusInternalServerError,
				"Something Went Wrong",
				"We couldn't finish setting up your company account.",
			)
			return
		}
	}

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		err := utils.RenderTemplate(w, "login.html", nil)
		if err != nil {
			log.Printf("LoginHandler template error: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	if r.Method != http.MethodPost {
		RenderError(
			w,
			http.StatusMethodNotAllowed,
			"Method Not Allowed",
			"The requested method is not allowed on this page.",
		)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	if username == "" || password == "" {
		RenderError(
			w,
			http.StatusBadRequest,
			"Missing Credentials",
			"Username and password are required.",
		)
		return
	}

	user, err := database.GetUserByUsername(username)
	if err != nil {
		RenderError(
			w,
			http.StatusUnauthorized,
			"Login Failed",
			"Invalid username or password.",
		)
		return
	}

	if !utils.CheckPassword(user.PasswordHash, password) {
		RenderError(
			w,
			http.StatusUnauthorized,
			"Login Failed",
			"Invalid username or password.",
		)
		return
	}

	token, err := utils.GenerateToken()
	if err != nil {
		log.Printf("LoginHandler GenerateToken error: %v", err)
		RenderError(
			w,
			http.StatusInternalServerError,
			"Something Went Wrong",
			"We couldn't log you in right now.",
		)
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
		log.Printf("LoginHandler CreateSession error: %v", err)
		RenderError(
			w,
			http.StatusInternalServerError,
			"Something Went Wrong",
			"We couldn't create your login session.",
		)
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
