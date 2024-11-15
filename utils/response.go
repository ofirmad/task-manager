package utils

import (
	"encoding/json"
	"net/http"
)

// SendResponse sends a JSON response
func SendResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// SendError sends an error response
func SendError(w http.ResponseWriter, message string, statusCode int) {
	SendResponse(w, map[string]string{"error": message}, statusCode)
}
