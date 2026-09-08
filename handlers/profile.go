package handlers

import (
	"jobFlow/database"
	"jobFlow/models"
	"jobFlow/utils"
	"net/http"
)

func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, err := utils.GetUserIDFromSession(r)
	if userID == 0 {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	user, err := database.GetUserByID(userID)
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	profile, err := database.GetUserProfileByUserID(userID)
	if err != nil {

		http.Error(w, "failed to load profile", http.StatusInternalServerError)
		return

		profile = models.UserProfile{
			UserID: userID,
		}
	}

	skills, err := database.GetSkillsByUserID(userID)
	if err != nil {
		http.Error(w, "failed to load skills", http.StatusInternalServerError)
		return
	}

	experiences, err := database.GetExperiencesByUserID(userID)
	if err != nil {
		http.Error(w, "failed to load experience", http.StatusInternalServerError)
		return
	}

	education, err := database.GetEducationByUserID(userID)
	if err != nil {
		http.Error(w, "failed to load education", http.StatusInternalServerError)
		return
	}

	data := models.ProfilePageData{
		User:        user,
		Profile:     profile,
		Skills:      skills,
		Experiences: experiences,
		Education:   education,
	}

	if user.Role == "company" {
		company, err := database.GetCompanyByUserID(userID)
		if err == nil {
			data.Company = &company
		}
	}

	utils.RenderTemplate(w, "profile.html", data)
}
