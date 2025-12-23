package middleware

import (
	"log"
	"net/http"
	"os"
)

func APIKeyAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := os.Getenv("API_KEY")
		if apiKey == "" {
			// If not configured, warn but allow? Or deny?
			// Secure default: Deny unless configured? 
			// For this exercise, assume if not set, we might be in dev mode or it's a misconfig.
			// Let's log warning and deny to be safe.
			log.Println("Warning: API_KEY env var not set. Denying request.")
			http.Error(w, "Unauthorized: API Key not configured", http.StatusUnauthorized)
			return
		}

		clientKey := r.Header.Get("X-API-Key")
		if clientKey == "" {
			// Check query param as fallback? Requirement usually header.
			clientKey = r.URL.Query().Get("api_key")
		}

		if clientKey != apiKey {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
