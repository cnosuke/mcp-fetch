package server

import (
	"crypto/subtle"
	"net/http"
	"strings"
)

func withAuthMiddleware(next http.Handler, authToken string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if authToken == "" {
			next.ServeHTTP(w, r)
			return
		}

		token := extractBearerToken(r)
		if subtle.ConstantTimeCompare([]byte(token), []byte(authToken)) != 1 {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func withOriginValidation(next http.Handler, allowedOrigins []string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(allowedOrigins) == 0 {
			next.ServeHTTP(w, r)
			return
		}

		origin := r.Header.Get("Origin")
		if origin == "" {
			next.ServeHTTP(w, r)
			return
		}

		for _, allowed := range allowedOrigins {
			if origin == allowed {
				next.ServeHTTP(w, r)
				return
			}
		}

		http.Error(w, "Forbidden", http.StatusForbidden)
	})
}

func extractBearerToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		return ""
	}
	return strings.TrimPrefix(auth, "Bearer ")
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}
