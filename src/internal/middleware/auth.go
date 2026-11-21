package middleware

import (
	"dZev1/character-gallery/internal/auth"
	"encoding/json"
	"net/http"
	"strings"
)

func AuthMiddleware(jwtSecret string, skippedPaths ...string) func(http.Handler) http.Handler {
	skip := make(map[string]struct{})
	for _, p := range skippedPaths {
		skip[p] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := r.Method + " " + r.URL.Path
			if _, ok := skip[key]; ok {
				next.ServeHTTP(w, r)
				return
			}

			header := r.Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				writeUnauthorized(w)
				return
			}

			token := strings.TrimPrefix(header, "Bearer ")
			claims, err := auth.ValidateJWT(token, jwtSecret)
			if err != nil {
				writeUnauthorized(w)
				return
			}

			ctx := auth.ContextWithClaims(r.Context(), claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func writeUnauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(map[string]string{
		"error": "unauthorized",
		"code":  "UNAUTHORIZED",
	})
}
