package handlers

import (
	"encoding/json"
	"log"
	"net/http"
)

func writeJSONError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	err := json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
	if err != nil {
		log.Printf("writeJSONError encoding error: %v", err)
	}
}
