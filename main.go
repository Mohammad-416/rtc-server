package main

import (
	"app/urtc/db"
	"app/urtc/routers"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	// Initialize DB
	db.InitDB()

	// Setup router
	router := routers.SetupRoutes()

	port := os.Getenv("PORT")
	log.Println("Server running on port", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
