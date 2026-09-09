package handlers

import (
	"database/sql"
	"errors"
	"html/template"
	"log"
	"net/http"

	"jobFlow/database"
	"jobFlow/middleware"
)

func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		log.Println("PROFILE ERROR: user ID missing from context")
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	log.Printf("Loading profile for user ID: %d\n", userID)

	user, err := database.GetUserByID(userID)
	if err != nil {
		log.Printf("PROFILE ERROR - GetUserByID: %v\n", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	profile, err := database.GetUserProfileByUserID(userID)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("PROFILE ERROR - GetUserProfileByUserID: %v\n", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	experiences, err := database.GetExperiencesByUserID(userID)
	if err != nil {
		log.Printf("PROFILE ERROR - GetExperiencesByUserID: %v\n", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	education, err := database.GetEducationByUserID(userID)
	if err != nil {
		log.Printf("PROFILE ERROR - GetEducationByUserID: %v\n", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	skills, err := database.GetSkillsByUserID(userID)
	if err != nil {
		log.Printf("PROFILE ERROR - GetSkillsByUserID: %v\n", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	preferences, err := database.GetUserPreferenceByUserID(userID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("PROFILE ERROR - GetUserPreferenceByUserID: %v\n", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	employmentTypes, err := database.GetEmploymentTypesByUserID(userID)
	if err != nil {
		log.Printf("PROFILE ERROR - GetEmploymentTypesByUserID: %v\n", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	workArrangements, err := database.GetWorkArrangementsByUserID(userID)
	if err != nil {
		log.Printf("PROFILE ERROR - GetWorkArrangementsByUserID: %v\n", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	posts, err := database.GetPostsByUserID(userID)
	if err != nil {
		log.Printf("PROFILE ERROR - GetPostsByUserID: %v\n", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	followerCount, err := database.GetFollowerCount(userID)
	if err != nil {
		log.Printf("PROFILE ERROR - GetFollowerCount: %v\n", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	followingCount, err := database.GetFollowingCount(userID)
	if err != nil {
		log.Printf("PROFILE ERROR - GetFollowingCount: %v\n", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	data := struct {
		User             any
		Profile          any
		Experiences      any
		Education        any
		Skills           any
		Preferences      any
		EmploymentTypes  any
		WorkArrangements any
		Posts            any
		FollowerCount    int
		FollowingCount   int
	}{
		User:             user,
		Profile:          profile,
		Experiences:      experiences,
		Education:        education,
		Skills:           skills,
		Preferences:      preferences,
		EmploymentTypes:  employmentTypes,
		WorkArrangements: workArrangements,
		Posts:            posts,
		FollowerCount:    followerCount,
		FollowingCount:   followingCount,
	}

	tmpl, err := template.ParseFiles(
		"templates/base.html",
		"templates/profile.html",
	)

	if err != nil {
		log.Printf("PROFILE ERROR - ParseFiles: %v\n", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		log.Printf("PROFILE ERROR - ExecuteTemplate: %v\n", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}
