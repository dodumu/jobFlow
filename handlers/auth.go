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
		if err := utils.RenderTemplate(w, r, "register.html", nil); err != nil {
			log.Printf("RegisterHandler template error: %v", err)
			return
		}
		return
	}

	if r.Method != http.MethodPost {
		RenderError(
			w,
			r,
			http.StatusMethodNotAllowed,
			"Method Not Allowed",
			"The requested method is not allowed on this page.",
		)
		return
	}

	// Get user registration fields.
	username := r.FormValue("username")
	password := r.FormValue("password")
	firstName := r.FormValue("first_name")
	lastName := r.FormValue("last_name")
	email := r.FormValue("email")
	dob := r.FormValue("date_of_birth")
	role := r.FormValue("role")

	// Validate required user fields.
	if role == "" ||
		username == "" ||
		password == "" ||
		firstName == "" ||
		lastName == "" ||
		email == "" ||
		dob == "" {

		RenderError(
			w,
			r,
			http.StatusBadRequest,
			"Missing Information",
			"Please complete all required registration fields.",
		)
		return
	}

	if !validRoles[role] {
		RenderError(
			w,
			r,
			http.StatusBadRequest,
			"Invalid Account Type",
			"Please select a valid account type.",
		)
		return
	}

	// Validate company information BEFORE writing anything to the database.
	var company *models.Company

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
				r,
				http.StatusBadRequest,
				"Missing Company Information",
				"Please complete all required company fields.",
			)
			return
		}

		company = &models.Company{
			CompanyName: companyName,
			Description: description,
			Website:     website,
			Location:    location,
			Logo:        logo,
			Industry:    industry,
			FoundedYear: 0,
			CompanySize: companySize,
		}
	}

	// Hash the password only after validation succeeds.
	hashPassword, err := utils.HashPassword(password)
	if err != nil {
		log.Printf("RegisterHandler HashPassword error: %v", err)

		RenderError(
			w,
			r,
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

	// UserID will be assigned inside CreateAccount after the user is created.
	profile := models.UserProfile{}

	// Create the complete account in a single database transaction.
	_, err = database.CreateAccount(
		user,
		profile,
		company,
	)
	if err != nil {
		log.Printf("RegisterHandler CreateAccount error: %v", err)

		RenderError(
			w,
			r,
			http.StatusInternalServerError,
			"Registration Failed",
			"We couldn't create your account right now.",
		)
		return
	}

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		err := utils.RenderTemplate(w, r, "login.html", nil)
		if err != nil {
			log.Printf("LoginHandler template error: %v", err)
			return
		}

	}

	if r.Method != http.MethodPost {
		RenderError(
			w,
			r,
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
			r,
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
			r,
			http.StatusUnauthorized,
			"Login Failed",
			"Invalid username or password.",
		)
		return
	}

	if !utils.CheckPassword(user.PasswordHash, password) {
		RenderError(
			w,
			r,
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
			r,
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
			r,
			http.StatusInternalServerError,
			"Something Went Wrong",
			"We couldn't create your login session.",
		)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     constants.SessionCookieName,
		Value:    token,
		Expires:  expires,
		HttpOnly: true,
		Secure:   utils.IsProduction(),
		Path:     "/",

		SameSite: http.SameSiteLaxMode,
	})

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}
