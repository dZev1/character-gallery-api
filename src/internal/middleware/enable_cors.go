package middleware

import (
	"net/http"
	"os"
	"strings"
)

// defaultAllowedOrigin is used when CORS_ALLOW_ORIGIN is not set. The game
// frontend is a web app served at this address.
const defaultAllowedOrigin = "http://192.168.1.32:3000"

func allowOrigin() []string {
	raw := os.Getenv("CORS_ALLOW_ORIGIN")
	if strings.TrimSpace(raw) == "" {
		return []string{defaultAllowedOrigin}
	}

	var origins []string
	for _, o := range strings.Split(raw, ",") {
		if o = strings.TrimSpace(o); o != "" {
			origins = append(origins, o)
		}
	}
	if len(origins) == 0 {
		origins = []string{defaultAllowedOrigin}
	}
	return origins
}

func EnableCors(next http.Handler) http.Handler {
	allowed := allowOrigin()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// "*" is only safe for public/non-credentialed endpoints; the API uses
		// bearer tokens, but we still restrict by default to the game origin.
		if len(allowed) == 1 && allowed[0] == "*" {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		} else if origin != "" && containsOrigin(allowed, origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key")
		w.Header().Set("Vary", "Origin")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func containsOrigin(allowed []string, origin string) bool {
	for _, o := range allowed {
		if o == origin {
			return true
		}
	}
	return false
}
