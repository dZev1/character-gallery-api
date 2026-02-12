package middleware

import (
	"dZev1/character-gallery/models/auth"
	"net/http"
)

func RequireAPIKey(authStore auth.AuthStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apiKey := r.Header.Get("X-API-Key")
			if apiKey == "" {
				http.Error(w, "Missing API key", http.StatusUnauthorized)
				return
			}

			keyHash := auth.HashAPIKey(apiKey)
			valid, err := authStore.ValidateAPIKey(keyHash)
			if err != nil {
				http.Error(w, "Error validating API key", http.StatusInternalServerError)
				return
			}

			if !valid {
				http.Error(w, "Invalid API key", http.StatusUnauthorized)
				return
			}

			authStore.UpdateLastUsed(keyHash)

			next.ServeHTTP(w, r)
		})
	}
}