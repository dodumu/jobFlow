package handlers

import (
	"database/sql"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"
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

	// GET: load the existing profile and skills.
	if r.Method == http.MethodGet {
		profile, err := database.GetUserProfileByUserID(userID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			log.Printf(
				"EditProfileHandler GetUserProfileByUserID error: %v",
				err,
			)
			RenderError(
				w,
				r,
				http.StatusInternalServerError,
				"Something Went Wrong",
				"We couldn't load your profile information.",
			)
			return
		}

		skills, err := database.GetSkillsByUserID(userID)
		if err != nil {
			log.Printf(
				"EditProfileHandler GetSkillsByUserID error: %v",
				err,
			)
			RenderError(
				w,
				r,
				http.StatusInternalServerError,
				"Something Went Wrong",
				"We couldn't load your skills.",
			)
			return
		}

		data := models.EditProfilePageData{
			Profile: profile,
			Skills:  skills,
		}

		if err := utils.RenderTemplate(
			w,
			r,
			"edit-profile.html",
			data,
		); err != nil {
			log.Printf(
				"EditProfileHandler render edit-profile.html error: %v",
				err,
			)
		}
		return
	}

	// POST: limit and parse the submitted multipart form.
	const maxRequestSize = 6 << 20 // 6 MiB

	r.Body = http.MaxBytesReader(w, r.Body, maxRequestSize)

	if err := r.ParseMultipartForm(maxRequestSize); err != nil {
		var maxBytesErr *http.MaxBytesError

		if errors.As(err, &maxBytesErr) {
			RenderError(
				w,
				r,
				http.StatusRequestEntityTooLarge,
				"Upload Too Large",
				"Your request is too large. Please use a smaller image.",
			)
			return
		}

		log.Printf("EditProfileHandler ParseMultipartForm error: %v", err)
		RenderError(
			w,
			r,
			http.StatusBadRequest,
			"Invalid Form",
			"We couldn't process your submitted form.",
		)
		return
	}

	// Load the current profile so we can preserve its picture if no new one is uploaded.
	existingProfile, err := database.GetUserProfileByUserID(userID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf(
			"EditProfileHandler GetUserProfileByUserID error: %v",
			err,
		)
		RenderError(
			w,
			r,
			http.StatusInternalServerError,
			"Something Went Wrong",
			"We couldn't load your existing profile.",
		)
		return
	}

	oldProfilePicture := existingProfile.ProfilePicture

	// Load skills for ownership validation and for redisplaying the form if needed.
	existingSkills, err := database.GetSkillsByUserID(userID)
	if err != nil {
		log.Printf(
			"EditProfileHandler GetSkillsByUserID error: %v",
			err,
		)
		RenderError(
			w,
			r,
			http.StatusInternalServerError,
			"Something Went Wrong",
			"We couldn't verify your skills.",
		)
		return
	}

	// Collect the submitted profile fields.
	profile := models.UserProfile{
		UserID:         userID,
		ProfilePicture: oldProfilePicture,
		Headline:       strings.TrimSpace(r.FormValue("headline")),
		Bio:            strings.TrimSpace(r.FormValue("bio")),
		Location:       strings.TrimSpace(r.FormValue("location")),
	}

	// Read new skills separated by commas.
	newSkillsInput := strings.TrimSpace(r.FormValue("new_skills"))
	var newSkills []string

	seenNames := make(map[string]bool)

	for _, name := range strings.Split(newSkillsInput, ",") {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}

		// Avoid duplicate names within the same submission.
		normalizedName := strings.ToLower(name)
		if !seenNames[normalizedName] {
			newSkills = append(newSkills, name)
			seenNames[normalizedName] = true
		}
	}

	// Read skill IDs selected for removal.
	removeSkillValues := r.Form["remove_skill_ids"]
	var removeSkillIDs []int
	seenSkillIDs := make(map[int]bool)

	for _, value := range removeSkillValues {
		skillID, err := strconv.Atoi(value)
		if err != nil || skillID <= 0 {
			RenderError(
				w,
				r,
				http.StatusBadRequest,
				"Invalid Skill",
				"One or more submitted skill IDs are invalid.",
			)
			return
		}

		if !seenSkillIDs[skillID] {
			removeSkillIDs = append(removeSkillIDs, skillID)
			seenSkillIDs[skillID] = true
		}
	}

	// Only allow removal of skills assigned to the logged-in user.
	ownedSkills := make(map[int]bool)

	for _, skill := range existingSkills {
		ownedSkills[skill.ID] = true
	}

	for _, skillID := range removeSkillIDs {
		if !ownedSkills[skillID] {
			RenderError(
				w,
				r,
				http.StatusBadRequest,
				"Invalid Skill",
				"One or more selected skills do not belong to your profile.",
			)
			return
		}
	}

	// Helper to redisplay the edit form with the current skills.
	renderEditForm := func(status int, message string) {
		skills, err := database.GetSkillsByUserID(userID)
		if err != nil {
			log.Printf(
				"EditProfileHandler reload skills error: %v",
				err,
			)
			RenderError(
				w,
				r,
				http.StatusInternalServerError,
				"Something Went Wrong",
				"We couldn't reload your skills.",
			)
			return
		}

		w.WriteHeader(status)

		data := models.EditProfilePageData{
			Profile: profile,
			Skills:  skills,
			Error:   message,
		}

		if err := utils.RenderTemplate(
			w,
			r,
			"edit-profile.html",
			data,
		); err != nil {
			log.Printf(
				"EditProfileHandler render edit form error: %v",
				err,
			)
		}
	}

	// Retrieve the optional profile picture.
	file, fileHeader, err := r.FormFile("profile_picture")
	if err != nil && !errors.Is(err, http.ErrMissingFile) {
		var maxBytesErr *http.MaxBytesError

		if errors.As(err, &maxBytesErr) {
			renderEditForm(
				http.StatusRequestEntityTooLarge,
				"Your upload is too large. Please select a smaller image.",
			)
			return
		}

		log.Printf("EditProfileHandler FormFile error: %v", err)
		RenderError(
			w,
			r,
			http.StatusBadRequest,
			"Bad Request",
			"Unable to process your upload.",
		)
		return
	}

	var newProfilePicture string

	if file != nil {
		defer file.Close()

		if fileHeader.Size > constants.MaxProfilePictureSize {
			renderEditForm(
				http.StatusBadRequest,
				"Profile picture must be 5 MB or smaller.",
			)
			return
		}

		// Detect the actual file type from its contents.
		buffer := make([]byte, 512)
		_, err := file.Read(buffer)

		if err != nil && !errors.Is(err, io.EOF) {
			log.Printf("EditProfileHandler file read error: %v", err)
			renderEditForm(
				http.StatusBadRequest,
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
			renderEditForm(
				http.StatusBadRequest,
				"Profile picture must be a JPEG, PNG, or WebP image.",
			)
			return
		}

		// Rewind the file after MIME detection.
		if _, err := file.Seek(0, io.SeekStart); err != nil {
			log.Printf("EditProfileHandler file seek error: %v", err)
			renderEditForm(
				http.StatusInternalServerError,
				"We couldn't process your profile picture.",
			)
			return
		}

		filename, err := utils.SaveProfilePicture(
			file,
			userID,
			contentType,
		)
		if err != nil {
			log.Printf(
				"EditProfileHandler SaveProfilePicture error: %v",
				err,
			)
			renderEditForm(
				http.StatusInternalServerError,
				"We couldn't save your profile picture.",
			)
			return
		}

		newProfilePicture = filename
		profile.ProfilePicture = filename
	}

	// Update the profile fields and picture.
	if err := database.UpdateUserProfile(profile); err != nil {
		log.Printf(
			"EditProfileHandler UpdateUserProfile error: %v",
			err,
		)

		if newProfilePicture != "" {
			if cleanupErr := utils.DeleteProfilePicture(newProfilePicture); cleanupErr != nil {
				log.Printf(
					"EditProfileHandler new picture cleanup error: %v",
					cleanupErr,
				)
			}
		}

		RenderError(
			w,
			r,
			http.StatusInternalServerError,
			"Something Went Wrong",
			"We couldn't update your profile right now.",
		)
		return
	}

	// Save skill additions and removals together.
	if err := database.UpdateUserSkills(
		userID,
		newSkills,
		removeSkillIDs,
	); err != nil {
		log.Printf(
			"EditProfileHandler UpdateUserSkills error: %v",
			err,
		)

		// Do not delete the new picture here: the profile may already reference it.
		RenderError(
			w,
			r,
			http.StatusInternalServerError,
			"Profile Partially Updated",
			"Your profile information was saved, but we couldn't save your skill changes. Please review your skills and try again.",
		)
		return
	}

	// Delete the old picture only after the profile and skills have been saved.
	if newProfilePicture != "" &&
		oldProfilePicture != "" &&
		oldProfilePicture != newProfilePicture {
		if err := utils.DeleteProfilePicture(oldProfilePicture); err != nil {
			log.Printf(
				"EditProfileHandler old picture cleanup error: %v",
				err,
			)
		}
	}

	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}
func AddSkillHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	UserID, ok := utils.GetUserID(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	skillName := r.FormValue("skill_name")
	err := database.AddSkillsByName(UserID, skillName)
	if err != nil {
		log.Printf("Adding skill failed: %v", err)
		RenderError(w, r, http.StatusBadRequest, "Bad Request", "Could not add skill")
		return
	}
	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}
