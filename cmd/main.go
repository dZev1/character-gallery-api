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
	
	currentVersion := os.Getenv("API_VERSION")
	
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

	baseRoute := "/api/" + currentVersion

	http.HandleFunc("POST "+baseRoute+"/characters", handler.CreateCharacter)
	http.HandleFunc("GET "+baseRoute+"/characters", handler.GetAllCharacters)
	http.HandleFunc("GET "+baseRoute+"/characters/{id}", handler.GetCharacter)
	http.HandleFunc("PUT "+baseRoute+"/characters/{id}", handler.EditCharacter)
	http.HandleFunc("DELETE "+baseRoute+"/characters/{id}", handler.DeleteCharacter)

	http.HandleFunc("POST "+baseRoute+"/characters/{character_id}/inventory/{item_id}", handler.AddItemToCharacter)
	http.HandleFunc("DELETE "+baseRoute+"/characters/{character_id}/inventory/{item_id}", handler.RemoveItemFromCharacter)
	http.HandleFunc("GET "+baseRoute+"/characters/{character_id}/inventory", handler.GetCharacterInventory)

	log.Println("Server listening on http://localhost:8080")
	if err = http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Could not start server: %v", err)
	}
}
