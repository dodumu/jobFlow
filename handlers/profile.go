package handlers

import (
	"database/sql"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"

	"jobFlow/constants"
	"jobFlow/database"
	"jobFlow/models"
	"jobFlow/utils"
)

func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		RenderError(
			w,
			r,
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
			r,
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
			r,
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
			r,
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
			r,
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
			r,
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
			r,
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
			r,
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
			r,
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
			r,
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
			r,
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
			r,
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
			r,
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
				r,
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

	err = utils.RenderTemplate(w, r, "profile.html", data)
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
			r,
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
			r,
			http.StatusUnauthorized,
			"Unauthorized",
			"You must be logged in to edit your profile.",
		)
		return
	}

	// GET: display the existing profile.
	if r.Method == http.MethodGet {
		profile, err := database.GetUserProfileByUserID(userID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			log.Printf("EditProfileHandler GetUserProfileByUserID error: %v", err)

			RenderError(
				w,
				r,
				http.StatusInternalServerError,
				"Something Went Wrong",
				"We couldn't load your profile information.",
			)
			return
		}

		data := models.EditProfilePageData{
			Profile: profile,
		}

		if err := utils.RenderTemplate(w, r, "edit-profile.html", data); err != nil {
			log.Printf(
				"ProfileHandler RenderTemplate edit-profile.html error: %v",
				err,
			)
		}

		return
	}

	// POST: collect the submitted text fields first.
	// This lets us preserve them if file validation fails.
	profile := models.UserProfile{
		UserID:   userID,
		Headline: strings.TrimSpace(r.FormValue("headline")),
		Bio:      strings.TrimSpace(r.FormValue("bio")),
		Location: strings.TrimSpace(r.FormValue("location")),
	}

	// Retrieve the optional uploaded profile picture.
	file, fileHeader, err := r.FormFile("profile_picture")
	if err != nil && !errors.Is(err, http.ErrMissingFile) {
		log.Printf("EditProfileHandler FormFile error: %v", err)

		RenderError(
			w,
			r,
			http.StatusBadRequest,
			"Invalid File",
			"We couldn't process the selected profile picture.",
		)
		return
	}

	if file != nil {
		defer file.Close()

		// Reject profile pictures larger than 5 MB.
		if fileHeader.Size > constants.MaxProfilePictureSize {
			data := models.EditProfilePageData{
				Profile: profile,
				Error:   "Profile picture must be 5 MB or smaller.",
			}

			if err := utils.RenderTemplate(w, r, "edit-profile.html", data); err != nil {
				log.Printf("EditProfileHandler template error: %v", err)
			}

			return
		}

		// Read the first 512 bytes to detect the actual file type.
		buffer := make([]byte, 512)

		_, err = file.Read(buffer)
		if err != nil && !errors.Is(err, io.EOF) {
			log.Printf("EditProfileHandler file read error: %v", err)

			RenderError(
				w,
				r,
				http.StatusBadRequest,
				"Invalid File",
				"We couldn't read the selected profile picture.",
			)
			return
		}

		contentType := http.DetectContentType(buffer)

		allowedTypes := map[string]bool{
			"image/jpeg": true,
			"image/png":  true,
			"image/webp": true,
		}

		if !allowedTypes[contentType] {
			data := models.EditProfilePageData{
				Profile: profile,
				Error:   "Profile picture must be a JPEG, PNG, or WebP image.",
			}

			if err := utils.RenderTemplate(w, r, "edit-profile.html", data); err != nil {
				log.Printf("EditProfileHandler template error: %v", err)
			}

			return
		}
	}

	if err := database.UpdateUserProfile(profile); err != nil {
		log.Printf("EditProfileHandler UpdateUserProfile error: %v", err)

		RenderError(
			w,
			r,
			http.StatusInternalServerError,
			"Something Went Wrong",
			"We couldn't update your profile right now.",
		)
		return
	}

	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}
