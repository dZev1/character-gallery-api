package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"dZev1/character-gallery/handlers"
	"dZev1/character-gallery/internal/database"
	"dZev1/character-gallery/internal/middleware"
	"dZev1/character-gallery/models/inventory"
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

	itemFile, err := os.Open("./item_pool.json")
	if err != nil {
		log.Print(fmt.Errorf("could not open seed file: %w", err))
	}
	defer itemFile.Close()
	
	var items []inventory.Item
	if err := json.NewDecoder(itemFile).Decode(&items); err != nil {
		log.Print(fmt.Errorf("could not decode items json: %w", err))
	}

	gallery.SeedItems(items)

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

	http.HandleFunc("GET "+baseRoute+"/items", handler.ShowPoolItems)
	http.HandleFunc("POST "+baseRoute+"/items", handler.CreateItem)
	http.HandleFunc("GET "+baseRoute+"/items/{item_id}", handler.ShowItem)

	log.Println("Server listening on http://localhost:8080/"+baseRoute)
	if err = http.ListenAndServe(":8080", middleware.EnableCors(middleware.RequireAPIKey(gallery.GetAuthStore())(http.DefaultServeMux))); err != nil {
		log.Fatalf("Could not start server: %v", err)
	}
}
