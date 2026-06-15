package main

import (
	"log"
	"net/http"

	"dZev1/character-gallery/handlers"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "../openapi.yaml")
	})
	mux.HandleFunc("GET /docs", handlers.DocsHandler)

	// TODO: add the rest of the API routes here
	// mux.HandleFunc("POST /api/v1/characters", ...)
	// ...

	log.Println("Server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
