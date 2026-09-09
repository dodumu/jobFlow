package main

import (
	"jobFlow/database"
	"jobFlow/handlers"
	"jobFlow/middleware"
	"log"
	"net/http"
)

func main() {
	err := database.InitDB("database.db")
	if err != nil {
		log.Println(err)
	}

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
		"/posts/comment",
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
	http.Handle("/posts/delete", middleware.AuthMiddleware(
		http.HandlerFunc(handlers.DeletePostHandler),
	))
	http.HandleFunc("/register", handlers.RegisterHandler)
	http.HandleFunc("/login", handlers.LoginHandler)

	log.Println("server is running on :8081")

	http.ListenAndServe(":8081", nil)
}
