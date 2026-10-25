package utils

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	"jobFlow/database"
)

var templateFuncs = template.FuncMap{
	"formatDate": func(t time.Time) string {
		return t.Format("2006-01-02")
	},
}

func RenderTemplate(
	w http.ResponseWriter,
	r *http.Request,
	page string,
	data any,
) error {

	UserID, isAuthenticated := GetUserID(r)

	avatarURL := ""

	if isAuthenticated {
		profile, err := database.GetUserProfileByUserID(UserID)
		if err != nil {
			log.Printf("Unable to retrieve profile picture: %v", err)
		} else {
			avatarURL = "/uploads/profile-pictures/" + profile.ProfilePicture
		}
	}

	funcs := template.FuncMap{
		"formatDate": templateFuncs["formatDate"],

		"isAuthenticated": func() bool {
			return isAuthenticated
		},
		"currentUserAvatar": func() string {
			return avatarURL
		},
		"isActive": func(path string) bool {
			currentPath := r.URL.Path

			if path == "/" {
				return currentPath == "/"
			}

			return currentPath == path ||
				strings.HasPrefix(currentPath, path+"/")
		},
		"showNavbar": func() bool {
			return page != "forgot_password.html" &&
				page != "verify_reset.html" &&
				page != "reset_password.html"
		},

		"unreadNotificationCount": func() int {
			if !isAuthenticated {
				return 0
			}

			count, err := database.GetUnreadNotificationCount(int64(UserID))
			if err != nil {
				log.Printf("Unable to retrieve unread notification count: %v", err)
				return 0
			}

			return count
		},
	}

	tmpl, err := template.New("base.html").
		Funcs(funcs).
		ParseFiles(
			"templates/base.html",
			"templates/"+page,
		)

	if err != nil {
		return fmt.Errorf("parsing template %s: %w", page, err)
	}

	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		return fmt.Errorf("executing template %s: %w", page, err)
	}

	return nil
}
