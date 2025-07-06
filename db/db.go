package db

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func InitDB() {
	var err error
	DB, err = sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("Error connecting to the database: ", err)
	}
	if err = DB.Ping(); err != nil {
		log.Fatal("Error pinging DB: ", err)
	}
	log.Println("Connected to DB")

	log.Println("Initializing User Table")
	InitUserTable()
	log.Println("Initialized User Table Successfully")
	log.Println("Initializing Project Table")
	InitProjectTable()
	log.Println("Initialized Project Table Successfully")
}

func InitUserTable() {
	query := `
	CREATE TABLE IF NOT EXISTS users (
		id UUID PRIMARY KEY,
		github_id BIGINT UNIQUE NOT NULL,
		username TEXT NOT NULL,
		email TEXT UNIQUE NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`
	_, err := DB.Exec(query)
	if err != nil {
		log.Fatal("Failed to create users table: ", err)
	}
}

func InitProjectTable() {
	query := `
	CREATE TABLE IF NOT EXISTS projects (
		id UUID PRIMARY KEY,
		owner_id UUID REFERENCES users(id) ON DELETE CASCADE,
		name TEXT NOT NULL,
		description TEXT,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`

	_, err := DB.Exec(query)
	if err != nil {
		log.Fatal("Failed to create project table: ", err)
	}
}
