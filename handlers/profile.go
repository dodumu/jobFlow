package handlers

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strings"

	"jobFlow/database"
	"jobFlow/models"
	"jobFlow/utils"
)

func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		RenderError(
			w,
			http.StatusMethodNotAllowed,
			"Method Not Allowed",
			"The requested method is not allowed on this page.",
		)
		return
	}

	userID, ok := utils.GetUserID(r)
	if !ok {
		RenderError(
			w,
			http.StatusUnauthorized,
			"Unauthorized",
			"You must be logged in to view your profile.",
		)
		return
	}

	user, err := database.GetUserByID(userID)
	if err != nil {
		log.Printf("ProfileHandler GetUserByID error: %v", err)
		RenderError(
			w,
			http.StatusInternalServerError,
			"Something Went Wrong",
			"We couldn't load your account information.",
		)
		return
	}

	profile, err := database.GetUserProfileByUserID(userID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("ProfileHandler GetUserProfileByUserID error: %v", err)
		RenderError(
			w,
			http.StatusInternalServerError,
			"Something Went Wrong",
			"We couldn't load your profile information.",
		)
		return
	}

	experiences, err := database.GetExperiencesByUserID(userID)
	if err != nil {
		log.Printf("ProfileHandler GetExperiencesByUserID error: %v", err)
		RenderError(
			w,
			http.StatusInternalServerError,
			"Something Went Wrong",
			"We couldn't load your work experience.",
		)
		return
	}

	education, err := database.GetEducationByUserID(userID)
	if err != nil {
		log.Printf("ProfileHandler GetEducationByUserID error: %v", err)
		RenderError(
			w,
			http.StatusInternalServerError,
			"Something Went Wrong",
			"We couldn't load your education information.",
		)
		return
	}

	skills, err := database.GetSkillsByUserID(userID)
	if err != nil {
		log.Printf("ProfileHandler GetSkillsByUserID error: %v", err)
		RenderError(
			w,
			http.StatusInternalServerError,
			"Something Went Wrong",
			"We couldn't load your skills.",
		)
		return
	}

	preferences, err := database.GetUserPreferenceByUserID(userID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("ProfileHandler GetUserPreferenceByUserID error: %v", err)
		RenderError(
			w,
			http.StatusInternalServerError,
			"Something Went Wrong",
			"We couldn't load your job preferences.",
		)
		return
	}

	employmentTypes, err := database.GetEmploymentTypesByUserID(userID)
	if err != nil {
		log.Printf("ProfileHandler GetEmploymentTypesByUserID error: %v", err)
		RenderError(
			w,
			http.StatusInternalServerError,
			"Something Went Wrong",
			"We couldn't load your employment preferences.",
		)
		return
	}

	workArrangements, err := database.GetWorkArrangementsByUserID(userID)
	if err != nil {
		log.Printf("ProfileHandler GetWorkArrangementsByUserID error: %v", err)
		RenderError(
			w,
			http.StatusInternalServerError,
			"Something Went Wrong",
			"We couldn't load your work preferences.",
		)
		return
	}

	posts, err := database.GetPostsByUserID(userID)
	if err != nil {
		log.Printf("ProfileHandler GetPostsByUserID error: %v", err)
		RenderError(
			w,
			http.StatusInternalServerError,
			"Something Went Wrong",
			"We couldn't load your posts.",
		)
		return
	}

	followerCount, err := database.GetFollowerCount(userID)
	if err != nil {
		log.Printf("ProfileHandler GetFollowerCount error: %v", err)
		RenderError(
			w,
			http.StatusInternalServerError,
			"Something Went Wrong",
			"We couldn't load your follower information.",
		)
		return
	}

	followingCount, err := database.GetFollowingCount(userID)
	if err != nil {
		log.Printf("ProfileHandler GetFollowingCount error: %v", err)
		RenderError(
			w,
			http.StatusInternalServerError,
			"Something Went Wrong",
			"We couldn't load your following information.",
		)
		return
	}

	var company *models.Company

	companyData, err := database.GetCompanyByUserID(userID)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			log.Printf("ProfileHandler GetCompanyByUserID error: %v", err)
			RenderError(
				w,
				http.StatusInternalServerError,
				"Something Went Wrong",
				"We couldn't load your company information.",
			)
			return
		}
	} else {
		company = &companyData
	}

	data := models.ProfilePageData{
		User:             user,
		Profile:          profile,
		Company:          company,
		Skills:           skills,
		Experiences:      experiences,
		Education:        education,
		Preferences:      preferences,
		EmploymentTypes:  employmentTypes,
		WorkArrangements: workArrangements,
		Posts:            posts,
		FollowerCount:    followerCount,
		FollowingCount:   followingCount,
	}

	err = utils.RenderTemplate(w, "profile.html", data)
	if err != nil {
		log.Printf(
			"ProfileHandler RenderTemplate profile.html error: %v",
			err)
	}
}
func EditProfileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		RenderError(
			w,
			http.StatusMethodNotAllowed,
			"Method Not Allowed",
			"The requested method is not allowed on this page.",
		)
		return
	}

	userID, ok := utils.GetUserID(r)
	if !ok {
		RenderError(
			w,
			http.StatusUnauthorized,
			"Unauthorized",
			"You must be logged in to edit your profile.",
		)
		return
	}

	if r.Method == http.MethodGet {
		profile, err := database.GetUserProfileByUserID(userID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			log.Printf("EditProfileHandler GetUserProfileByUserID error: %v", err)
			RenderError(
				w,
				http.StatusInternalServerError,
				"Something Went Wrong",
				"We couldn't load your profile information.",
			)
			return
		}

		data := models.EditProfilePageData{
			Profile: profile,
		}
		err = utils.RenderTemplate(w, "edit-profile.html", data)
		if err != nil {
			log.Printf(
				"ProfileHandler RenderTemplate edit-profile.html error: %v",
				err)
		}
		return
	}

	profile := models.UserProfile{
		UserID:         userID,
		ProfilePicture: strings.TrimSpace(r.FormValue("profile_picture")),
		Headline:       strings.TrimSpace(r.FormValue("headline")),
		Bio:            strings.TrimSpace(r.FormValue("bio")),
		Location:       strings.TrimSpace(r.FormValue("location")),
	}

	if err := database.UpdateUserProfile(profile); err != nil {
		log.Printf("EditProfileHandler UpdateUserProfile error: %v", err)

		RenderError(
			w,
			http.StatusInternalServerError,
			"Something Went Wrong",
			"We couldn't update your profile right now.",
		)
		return
	}

	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}
