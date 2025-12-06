package main

import (
	"log"
	"net/http"
	"os"

	"github.com/dZev1/character-gallery/internal/database"
	"github.com/dZev1/character-gallery/handlers"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("error loading .env file")
	}
	connStr := os.Getenv("DATABASE_URL")

	schemaPath := os.Getenv("SCHEMA_PATH")

	gallery, err := database.NewPostgresCharacterGallery(connStr, schemaPath)
	if err != nil {
		panic(err)
	}

	if g, ok := gallery.(*database.PostgresCharacterGallery); ok {
		defer g.Close()
	}

	handler := &handlers.CharacterHandler{
		Gallery: gallery,
	}

	http.HandleFunc("POST /characters", handler.CreateCharacter)
	http.HandleFunc("GET /characters", handler.GetAllCharacters)
	http.HandleFunc("GET /characters/{id}", handler.GetCharacter)
	http.HandleFunc("PUT /characters/{id}", handler.EditCharacter)
	http.HandleFunc("DELETE /characters/{id}", handler.DeleteCharacter)

	log.Println("Server listening on http://localhost:8080")
	if err = http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Could not start server: %v", err)
	}
}
