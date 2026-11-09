package utils

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"jobFlow/database"
	"jobFlow/middleware"
)

var templateFuncs = template.FuncMap{
	"formatDate": func(t time.Time) string {
		return t.Format("2006-01-02")
	},
}

var stubFuncs = template.FuncMap{
	"formatDate":        templateFuncs["formatDate"],
	"isAuthenticated":   func() bool { return false },
	"currentUserAvatar": func() string { return "" },
	"isActive":          func(path string) bool { return false },
	"showNavbar": func() bool {
		return false
	},
	"unreadNotificationCount": func() int { return 0 },
	"csrfToken":               func() string { return "" },
	"csrfField":               func() template.HTML { return "" },
}

var templateCache = map[string]*template.Template{}

func LoadTemplates() error {
	pages, err := filepath.Glob("templates/*.html")
	if err != nil {
		return fmt.Errorf("listing timplates: %w", err)
	}

	for _, pagePath := range pages {
		name := filepath.Base(pagePath)

		if name == "base.html" {
			continue
		}

		tmpl, err := template.New("base.html").
			Funcs(stubFuncs).ParseFiles("templates/base.html", pagePath)
		if err != nil {
			return fmt.Errorf("parsing template %s: %w", name, err)
		}

		templateCache[name] = tmpl
	}
	return nil
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
		"csrfToken": func() string {
			return middleware.CSRFTokenFromRequest(r)
		},
		"csrfField": func() template.HTML {
			token := middleware.CSRFTokenFromRequest(r)
			if token == "" {
				return template.HTML("")
			}
			return template.HTML(
				`<input type="hidden" name="csrf_token" value="` + token + `">`,
			)

		},
	}

	cached, ok := templateCache[page]
	if !ok {
		return fmt.Errorf("template %s not found in cache", page)
	}

	tmpl, err := cached.Clone()
	if err != nil {
		return fmt.Errorf("cloning template %s: %w", page, err)
	}

	tmpl.Funcs(funcs)

	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		return fmt.Errorf("executing template %s: %w", page, err)
	}

	return nil
}
