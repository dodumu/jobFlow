package main

import (
	"log"
	"net/http"
	"os"

	"jobFlow/database"
	"jobFlow/handlers"
	"jobFlow/middleware"
)

func main() {
	// Initialize database.
	databasePath := os.Getenv("DATABASE_PATH")
	if databasePath == "" {
		databasePath = "database.db"
	}

	err := database.InitDB(databasePath)
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}

	// Configure uploads directory.
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

	// Home and profile routes.
	http.Handle(
		"GET /home",
		middleware.AuthMiddleware(http.HandlerFunc(handlers.HomeHandler)),
	)

	http.Handle(
		"GET /dashboard",
		middleware.AuthMiddleware(http.HandlerFunc(handlers.DashboardHandler)),
	)

	http.Handle(
		"GET /profile",
		middleware.AuthMiddleware(http.HandlerFunc(handlers.ProfileHandler)),
	)

	http.Handle(
		"GET /profile/edit",
		middleware.AuthMiddleware(http.HandlerFunc(handlers.EditProfileHandler)),
	)

	// Post routes.
	http.Handle(
		"POST /posts/create",
		middleware.AuthMiddleware(http.HandlerFunc(handlers.CreatePostHandler)),
	)

	http.Handle(
		"POST /posts/like/",
		middleware.AuthMiddleware(http.HandlerFunc(handlers.LikePostHandler)),
	)

	http.Handle(
		"POST /posts/unlike/",
		middleware.AuthMiddleware(http.HandlerFunc(handlers.UnlikePostHandler)),
	)

	http.Handle(
		"GET /posts/{id}",
		middleware.AuthMiddleware(http.HandlerFunc(handlers.PostDetailsHandler)),
	)

	http.Handle(
		"POST /posts/share",
		middleware.AuthMiddleware(http.HandlerFunc(handlers.SharePostHandler)),
	)

	http.Handle(
		"POST /posts/delete",
		middleware.AuthMiddleware(http.HandlerFunc(handlers.DeletePostHandler)),
	)

	// Comment routes.
	http.Handle(
		"POST /posts/comments/create",
		middleware.AuthMiddleware(http.HandlerFunc(handlers.CreateCommentHandler)),
	)

	http.Handle(
		"GET /posts/comments",
		middleware.AuthMiddleware(http.HandlerFunc(handlers.GetCommentsHandler)),
	)

	http.Handle(
		"POST /posts/comments/delete",
		middleware.AuthMiddleware(http.HandlerFunc(handlers.DeleteCommentHandler)),
	)

	// Authentication and session routes.
	http.Handle(
		"POST /logout",
		middleware.AuthMiddleware(http.HandlerFunc(handlers.LogoutHandler)),
	)

	http.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	})

	http.HandleFunc("/register", handlers.RegisterHandler)
	http.HandleFunc("/login", handlers.LoginHandler)
	http.HandleFunc("/forgot-password", handlers.ForgotPasswordHandler)
	http.HandleFunc("/verify-reset", handlers.VerifyResetHandler)
	http.HandleFunc("/reset-password", handlers.ResetPasswordHandler)

	// Follow routes.
	http.Handle(
		"POST /users/{id}/follow",
		middleware.AuthMiddleware(http.HandlerFunc(handlers.FollowHandler)),
	)

	http.Handle(
		"POST /users/{id}/unfollow",
		middleware.AuthMiddleware(http.HandlerFunc(handlers.UnfollowHandler)),
	)

	http.Handle(
		"GET /users/{id}/followers",
		middleware.AuthMiddleware(http.HandlerFunc(handlers.FollowersHandler)),
	)

	http.Handle(
		"GET /users/{id}/following",
		middleware.AuthMiddleware(http.HandlerFunc(handlers.FollowingHandler)),
	)

	http.Handle(
		"GET /users/{id}",
		middleware.AuthMiddleware(http.HandlerFunc(handlers.PublicProfileHandler)),
	)

	// Job routes.
	http.Handle(
		"/jobs",
		middleware.AuthMiddleware(http.HandlerFunc(handlers.JobsHandler)),
	)

	http.Handle(
		"GET /jobs/create",
		middleware.AuthMiddleware(http.HandlerFunc(handlers.CreateJobHandler)),
	)

	http.Handle(
		"POST /jobs/create",
		middleware.AuthMiddleware(http.HandlerFunc(handlers.CreateJobHandler)),
	)

	http.Handle(
		"GET /jobs/{id}/edit",
		middleware.AuthMiddleware(http.HandlerFunc(handlers.EditJobHandler)),
	)

	http.Handle(
		"POST /jobs/{id}/edit",
		middleware.AuthMiddleware(http.HandlerFunc(handlers.EditJobHandler)),
	)

	http.Handle(
		"POST /jobs/{id}/close",
		middleware.AuthMiddleware(http.HandlerFunc(handlers.CloseJobHandler)),
	)

	http.Handle(
		"/jobs/{id}/apply",
		middleware.AuthMiddleware(http.HandlerFunc(handlers.ApplyJobHandler)),
	)

	http.Handle(
		"GET /jobs/{id}",
		middleware.AuthMiddleware(http.HandlerFunc(handlers.JobDetailsHandler)),
	)

	http.Handle(
		"GET /applications",
		middleware.AuthMiddleware(http.HandlerFunc(handlers.ApplicationHandler)),
	)

	http.Handle(
		"GET /my-applications/{id}",
		middleware.AuthMiddleware(http.HandlerFunc(handlers.ViewMyApplicationHandler)),
	)

	http.Handle(
		"GET /applications/{id}",
		middleware.AuthMiddleware(http.HandlerFunc(handlers.ViewApplicationHandler)),
	)

	http.Handle(
		"POST /applications/{id}/accept",
		middleware.AuthMiddleware(http.HandlerFunc(handlers.AcceptApplicationHandler)),
	)

	http.Handle(
		"POST /applications/{id}/reject",
		middleware.AuthMiddleware(http.HandlerFunc(handlers.RejectApplicationHandler)),
	)

	// Notification routes.
	http.Handle(
		"GET /notifications",
		middleware.AuthMiddleware(http.HandlerFunc(handlers.NotificationsHandler)),
	)

	http.Handle(
		"GET /notifications/unread-count",
		middleware.AuthMiddleware(http.HandlerFunc(handlers.UnreadNotificationCountHandler)),
	)

	http.Handle(
		"POST /notifications/read",
		middleware.AuthMiddleware(http.HandlerFunc(handlers.MarkNotificationAsReadHandler)),
	)

	http.Handle(
		"GET /notifications/stream",
		middleware.AuthMiddleware(http.HandlerFunc(handlers.NotificationStreamHandler)),
	)

	// Start server.
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
