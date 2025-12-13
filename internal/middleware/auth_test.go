package middleware

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBasicAuth(t *testing.T) {
	// Create a simple test handler that always succeeds
	successHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	t.Run("no auth configured allows all requests", func(t *testing.T) {
		middleware := BasicAuth("", "")
		handler := middleware(successHandler)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", rec.Code)
		}
		if rec.Body.String() != "success" {
			t.Errorf("Expected 'success', got %s", rec.Body.String())
		}
	})

	t.Run("valid credentials allow access", func(t *testing.T) {
		middleware := BasicAuth("testuser", "testpass")
		handler := middleware(successHandler)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		auth := base64.StdEncoding.EncodeToString([]byte("testuser:testpass"))
		req.Header.Set("Authorization", "Basic "+auth)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", rec.Code)
		}
		if rec.Body.String() != "success" {
			t.Errorf("Expected 'success', got %s", rec.Body.String())
		}
	})

	t.Run("missing credentials are rejected", func(t *testing.T) {
		middleware := BasicAuth("testuser", "testpass")
		handler := middleware(successHandler)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", rec.Code)
		}

		if rec.Header().Get("WWW-Authenticate") != `Basic realm="Restricted"` {
			t.Errorf("Expected WWW-Authenticate header, got %s", rec.Header().Get("WWW-Authenticate"))
		}

		if rec.Header().Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", rec.Header().Get("Content-Type"))
		}
	})

	t.Run("invalid username is rejected", func(t *testing.T) {
		middleware := BasicAuth("testuser", "testpass")
		handler := middleware(successHandler)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		auth := base64.StdEncoding.EncodeToString([]byte("wronguser:testpass"))
		req.Header.Set("Authorization", "Basic "+auth)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", rec.Code)
		}
	})

	t.Run("invalid password is rejected", func(t *testing.T) {
		middleware := BasicAuth("testuser", "testpass")
		handler := middleware(successHandler)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		auth := base64.StdEncoding.EncodeToString([]byte("testuser:wrongpass"))
		req.Header.Set("Authorization", "Basic "+auth)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", rec.Code)
		}
	})

	t.Run("malformed Authorization header is rejected", func(t *testing.T) {
		middleware := BasicAuth("testuser", "testpass")
		handler := middleware(successHandler)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "InvalidFormat")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", rec.Code)
		}
	})
}
