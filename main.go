package main

import (
	"fmt"
	"github.com/ofirmad/task-manager/handlers"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	// Separate the routes into their own handlers package
	mux.HandleFunc("/tasks", handlers.HandleTasks)
	mux.HandleFunc("/tasks/", handlers.HandleTaskByID)

	// Wrap the mux with the CORS middleware
	handler := corsMiddleware(mux)

	fmt.Println("Server is running on http://localhost:8080")
	if err := http.ListenAndServe(":8080", handler); err != nil {
		fmt.Printf("server failed to start: %v", err)
	}
}

// corsMiddleware sets the CORS headers
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
