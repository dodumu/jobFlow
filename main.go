package main

import (
	"fmt"
	"jobFlow/database"
	"jobFlow/middleware"
	"log"
	"net/http"
)

func DashboardHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey)
	fmt.Fprintf(w, "Welcome %v", userID)
}
func main() {
	err := database.InitDB("database.db")
	if err != nil {
		log.Println(err)
	}

	http.Handle("/dashboard", middleware.AuthMiddleware(http.HandlerFunc(DashboardHandler)))
	log.Println("server is running on :8081")
	http.ListenAndServe(":8081", nil)
}
