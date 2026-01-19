package main

import (
	"log"
	"net/http"
	"os"

	"github.com/dZev1/character-gallery/handlers"
	"github.com/dZev1/character-gallery/internal/database"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("error loading .env file")
	}
	connectionString := os.Getenv("DATABASE_URL")

	err = godotenv.Overload("./config.env")
	dbType := os.Getenv("DATABASE_TYPE")

	gallery, err := database.NewCharacterGallery(dbType, connectionString)
	if err != nil {
		panic(err)
	}
	defer gallery.Close()

	gallery.SeedItems()

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
