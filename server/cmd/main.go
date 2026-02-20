package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"dZev1/character-gallery/handlers"
	"dZev1/character-gallery/internal/database"
	"dZev1/character-gallery/internal/middleware"
	"dZev1/character-gallery/models/inventory"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: Could not load .env file, relying on environment variables")
	}

	connectionString := os.Getenv("DATABASE_URL")
	currentVersion := os.Getenv("API_VERSION")

	err = godotenv.Overload("./config.env")
	if err != nil {
		log.Println("Warning: Could not load config.env file, relying on environment variables")
	}

	dbType := os.Getenv("DATABASE_TYPE")

	gallery, err := database.NewCharacterGallery(dbType, connectionString)
	if err != nil {
		panic(err)
	}
	defer gallery.Close()

	itemFile, err := os.Open("./item_pool.json")
	if err != nil {
		log.Println(fmt.Errorf("could not open seed file: %w", err))
	}
	defer itemFile.Close()

	var items []inventory.Item
	if err := json.NewDecoder(itemFile).Decode(&items); err != nil {
		log.Println(fmt.Errorf("could not decode items json: %w", err))
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

	mux := http.NewServeMux()
	
	mux.HandleFunc("POST "+baseRoute+"/characters", handler.CreateCharacter)
	mux.HandleFunc("GET "+baseRoute+"/characters", handler.GetAllCharacters)
	mux.HandleFunc("GET "+baseRoute+"/characters/{id}", handler.GetCharacter)
	mux.HandleFunc("PUT "+baseRoute+"/characters/{id}", handler.EditCharacter)
	mux.HandleFunc("DELETE "+baseRoute+"/characters/{id}", handler.DeleteCharacter)

	mux.HandleFunc("POST "+baseRoute+"/characters/{character_id}/inventory/{item_id}", handler.AddItemToCharacter)
	mux.HandleFunc("DELETE "+baseRoute+"/characters/{character_id}/inventory/{item_id}", handler.RemoveItemFromCharacter)
	mux.HandleFunc("GET "+baseRoute+"/characters/{character_id}/inventory", handler.GetCharacterInventory)

	mux.HandleFunc("GET "+baseRoute+"/items", handler.ShowPoolItems)
	mux.HandleFunc("POST "+baseRoute+"/items", handler.CreateItem)
	mux.HandleFunc("GET "+baseRoute+"/items/{item_id}", handler.ShowItem)

	handler_with_middlewares := middleware.EnableCors(middleware.RequireAPIKey(gallery.GetAuthStore())(mux))

	server := &http.Server{
		Addr:         ":8080",
		Handler:      handler_with_middlewares,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Println("Server listening on http://localhost:8080" + baseRoute)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Could not start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	server.Shutdown(ctx)

}
