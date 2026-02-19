package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"dZev1/character-gallery/models/auth"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
)

func main() {
	name := flag.String("name", "CLI Generated Key", "Name/Description for the API key")
	flag.Parse()

	if err := godotenv.Load(); err != nil {
		log.Println("Warning: No .env file found")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	db, err := sqlx.Connect("pgx", dbURL)
	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}
	defer db.Close()

	keyHash, rawKey, err := auth.GenerateAPIKey()
	if err != nil {
		log.Fatalf("Error generating API key: %v", err)
	}

	query := `
		INSERT INTO api_keys (name, key_hash)
		VALUES ($1, $2)
		RETURNING id
	`

	var id int64
	err = db.QueryRow(query, *name, keyHash).Scan(&id)
	if err != nil {
		log.Fatalf("Error inserting API key into database: %v", err)
	}

	fmt.Println("API Key Generated Successfully!")
	fmt.Printf("ID:   %d\n", id)
	fmt.Printf("Name: %s\n", *name)
	fmt.Printf("Key:  %s\n", rawKey)
	fmt.Println("\nWARNING: This key will NOT be shown again. Save it securely!")
}
