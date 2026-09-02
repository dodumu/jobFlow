package handlers

import (
	"jobFlow/database"
	"jobFlow/models"
	"jobFlow/utils"
	"net/http"
)

var validRoles = map[string]bool{
	"company":    true,
	"individual": true,
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	username := r.FormValue("username")
	password := r.FormValue("password")
	name := r.FormValue("name")
	email := r.FormValue("email")
	dob := r.FormValue("date_of_birth")
	role := r.FormValue("role")

	if role == "" || username == "" || password == "" || name == "" || email == "" || dob == "" {
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
		Name:         name,
		Email:        email,
		DOB:          dob,
		Role:         role,
	}
	_, err = database.CreateUser(user)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}
