package main

import (
	"app/urtc/db"
	"app/urtc/middleware"
	"app/urtc/routers"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
)

func main() {
	// Load environment variables
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Initialize database
	log.Println("Initializing database...")
	db.InitDB()
	log.Println("Database initialized successfully")

	// Setup routes
	log.Println("Setting up routes...")
	router := routers.SetupRoutes()

	// Create rate limiter (60 requests per minute, burst of 10)
	rateLimiter := middleware.NewRateLimiter(60, 10)

	// Apply middleware stack
	handler := middleware.Recovery(
		middleware.RequestLogger(
			middleware.SecurityHeaders(
				middleware.CORS(
					rateLimiter.Limit(
						middleware.UserContext(
							middleware.ProjectContext(
								router,
							),
						),
					),
				),
			),
		),
	)

	// Optional: Add CORS with more options using gorilla/handlers
	corsHandler := handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}), // Configure for production
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization", "X-User-ID", "X-Project-ID"}),
		handlers.AllowCredentials(),
	)(handler)

	// Start server
	log.Printf("Server starting on port %s...", port)
	log.Printf("WebSocket endpoint: ws://localhost:%s/ws", port)
	log.Printf("API base URL: http://localhost:%s", port)

	if err := http.ListenAndServe(":"+port, corsHandler); err != nil {
		log.Fatal("Server failed to start: ", err)
	}
}
