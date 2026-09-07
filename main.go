package main

import (
	"fmt"
	"jobFlow/database"
	"jobFlow/handlers"
	"jobFlow/middleware"
	"log"
	"net/http"
)

func DashboardHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	userDetails, err := database.GetUserByID(int(userID))
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "Welcome %v", userDetails.Username)
}
func main() {
	err := database.InitDB("database.db")
	if err != nil {
		log.Println(err)
	}

	http.Handle("/dashboard", middleware.AuthMiddleware(http.HandlerFunc(DashboardHandler)))
	http.HandleFunc("/register", handlers.RegisterHandler)
	http.HandleFunc("/login", handlers.LoginHandler)
	log.Println("server is running on :8081")
	http.ListenAndServe(":8081", nil)
}
