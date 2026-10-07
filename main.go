package main

import (
	"jobFlow/database"
	"jobFlow/handlers"
	"jobFlow/middleware"
	"log"
	"net/http"
	"os"
)

func main() {
	
	databasePath := os.Getenv("DATABASE_PATH")
	if databasePath == "" {
		databasePath = "database.db"
	}

	err := database.InitDB(databasePath)
	if err != nil {
		log.Println(err)
	}

	uploadPath := os.Getenv("UPLOAD_PATH")
	if uploadPath == "" {
		uploadPath = "uploads"
	}

	profilePicturesPath := uploadPath + "/profile-pictures"

	http.Handle(
		"/uploads/profile-pictures/",
		http.StripPrefix(
			"/uploads/profile-pictures/",
			http.FileServer(http.Dir(profilePicturesPath)),
		),
	)

	http.Handle(
		"/home",
		middleware.AuthMiddleware(
			http.HandlerFunc(handlers.HomeHandler),
		),
	)
	http.Handle(
		"/dashboard",
		middleware.AuthMiddleware(
			http.HandlerFunc(handlers.DashboardHandler),
		),
	)
	http.Handle(
		"/profile",
		middleware.AuthMiddleware(
			http.HandlerFunc(handlers.ProfileHandler),
		),
	)
	http.Handle(
		"/profile/edit",
		middleware.AuthMiddleware(
			http.HandlerFunc(handlers.EditProfileHandler),
		),
	)
	http.Handle(
		"/posts/create",
		middleware.AuthMiddleware(
			http.HandlerFunc(handlers.CreatePostHandler),
		),
	)
	http.Handle(
		"/posts/like/",
		middleware.AuthMiddleware(
			http.HandlerFunc(handlers.LikePostHandler),
		),
	)

	http.Handle(
		"/posts/unlike/",
		middleware.AuthMiddleware(
			http.HandlerFunc(handlers.UnlikePostHandler),
		),
	)
	http.Handle(
		"/posts/comments/create",
		middleware.AuthMiddleware(
			http.HandlerFunc(handlers.CreateCommentHandler),
		),
	)
	http.Handle(
		"/posts/share",
		middleware.AuthMiddleware(
			http.HandlerFunc(handlers.SharePostHandler),
		),
	)

	http.Handle(
		"/posts/comments",
		middleware.AuthMiddleware(
			http.HandlerFunc(handlers.GetCommentsHandler),
		),
	)
	http.Handle(
		"/logout",
		middleware.AuthMiddleware(
			http.HandlerFunc(handlers.LogoutHandler),
		),
	)
	http.Handle(
		"/posts/comments/delete",
		middleware.AuthMiddleware(
			http.HandlerFunc(handlers.DeleteCommentHandler),
		),
	)
	http.Handle("/posts/delete", middleware.AuthMiddleware(
		http.HandlerFunc(handlers.DeletePostHandler),
	))
	http.Handle(
		"/jobs",
		middleware.AuthMiddleware(
			http.HandlerFunc(handlers.JobsHandler),
		),
	)
	http.Handle(
		"GET /jobs/{id}",
		middleware.AuthMiddleware(
			http.HandlerFunc(handlers.JobDetailsHandler),
		),
	)
	http.Handle(
		"GET /jobs/create",
		middleware.AuthMiddleware(
			http.HandlerFunc(handlers.CreateJobHandler),
		),
	)

	http.Handle(
		"POST /jobs/create",
		middleware.AuthMiddleware(
			http.HandlerFunc(handlers.CreateJobHandler),
		),
	)
	http.Handle(
		"GET /jobs/{id}/edit",
		middleware.AuthMiddleware(
			http.HandlerFunc(handlers.EditJobHandler),
		),
	)

	http.Handle(
		"POST /jobs/{id}/edit",
		middleware.AuthMiddleware(
			http.HandlerFunc(handlers.EditJobHandler),
		),
	)
	http.Handle(
		"POST /jobs/{id}/close",
		middleware.AuthMiddleware(
			http.HandlerFunc(handlers.CloseJobHandler),
		),
	)
	http.Handle(
		"/jobs/{id}/apply",
		middleware.AuthMiddleware(
			http.HandlerFunc(handlers.ApplyJobHandler),
		),
	)
	http.Handle(
		"GET /applications",
		middleware.AuthMiddleware(
			http.HandlerFunc(handlers.ApplicationHandler),
		),
	)
	http.Handle(
		"GET /applications/{id}",
		middleware.AuthMiddleware(
			http.HandlerFunc(handlers.ViewApplicationHandler),
		),
	)
	http.Handle(
		"POST /applications/{id}/accept",
		middleware.AuthMiddleware(
			http.HandlerFunc(handlers.AcceptApplicationHandler),
		),
	)
	http.Handle(
		"POST /applications/{id}/reject",
		middleware.AuthMiddleware(
			http.HandlerFunc(handlers.RejectApplicationHandler),
		),
	)
	http.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	})
	http.HandleFunc("/register", handlers.RegisterHandler)
	http.HandleFunc("/login", handlers.LoginHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	address := ":" + port
	log.Printf("server is running on %s", address)

	if err := http.ListenAndServe(address, nil); err != nil {
		log.Fatal("server failed: ", err)
	}
}
