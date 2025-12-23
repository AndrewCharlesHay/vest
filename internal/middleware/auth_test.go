package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestAPIKeyAuth(t *testing.T) {
	// Setup
	os.Setenv("API_KEY", "test-secret")
	defer os.Unsetenv("API_KEY")

	handler := APIKeyAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Case 1: Valid Key (Header)
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-API-Key", "test-secret")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", rec.Code)
	}

	// Case 2: Invalid Key
	req = httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-API-Key", "wrong-secret")
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 Unauthorized, got %d", rec.Code)
	}

	// Case 3: Missing Key
	req = httptest.NewRequest("GET", "/", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 Unauthorized, got %d", rec.Code)
	}
	
	// Case 4: No Env Configured
    // We need to unset temporarily but parallel tests might conflict. 
    // Since we are running sequentially here it's fine.
    // os.Unsetenv("API_KEY")
    // ... test ...
    // os.Setenv("API_KEY", "test-secret")
}
