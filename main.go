package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"jobFlow/ai"
	"jobFlow/database"
	"jobFlow/handlers"
	"jobFlow/middleware"
	"jobFlow/utils"
	"jobFlow/worker"
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

	if err := utils.LoadTemplates(); err != nil {
		log.Fatalf("loading tmeplates: %v", err)
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

	server := &http.Server{
		Addr:              address,
		Handler:           nil,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go cleanupSessions(ctx)

	// Any AI job that was mid-run when the app last stopped goes back in the queue.
	requeued, err := database.RequeueRunningAIJobs()
	if err != nil {
		log.Fatalf("requeueing ai jobs: %v", err)
	}
	if requeued > 0 {
		log.Printf("requeued %d interrupted ai jobs", requeued)
	}

	aiWorker := worker.New()
	aiWorker.Register("ping", worker.PingHandler)

	aiClient, err := ai.NewFromEnv()
	switch {
	case errors.Is(err, ai.ErrNoAPIKey):
		log.Println("AI_API_KEY is not set: AI features are disabled")
	case err != nil:
		log.Fatalf("setting up the AI client: %v", err)
	default:
		aiWorker.Register("ai_ping", worker.NewAIPingHandler(aiClient))
	}

	workerDone := make(chan struct{})
	go func() {
		aiWorker.Run(ctx)
		close(workerDone)
	}()

	go func() {
		log.Printf("server is running on %s", address)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server failed: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown did not finish: %v", err)
	}
	// Let a running AI job finish (up to 10 seconds) before closing the database.
	select {
	case <-workerDone:
	case <-time.After(10 * time.Second):
		log.Println("ai worker did not stop in time")
	}
	database.DB.Close()
	log.Println("bye")

}

func cleanupSessions(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			deleted, err := database.DeleteExpiredSessions()
			if err != nil {
				log.Printf("session cleanup failed: %v", err)
				continue
			}
			log.Printf("session cleanup: removed %d expired sessions", deleted)
		case <-ctx.Done():
			return
		}
	}
}
