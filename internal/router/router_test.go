package router

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/dangogh/silver-eureka/internal/database"
)

func setupTestDB(t *testing.T) *database.DB {
	t.Helper()

	tmpFile := t.TempDir() + "/test.db"
	db, err := database.New(tmpFile)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	return db
}

func TestHealthEndpoint_Healthy(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	router := New(db, "", "")

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}

	var response map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["status"] != "healthy" {
		t.Errorf("Expected status 'healthy', got '%s'", response["status"])
	}

	if response["database"] != "up" {
		t.Errorf("Expected database 'up', got '%s'", response["database"])
	}
}

func TestHealthEndpoint_Unhealthy(t *testing.T) {
	db := setupTestDB(t)
	db.Close()

	router := New(db, "", "")

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status 503, got %d", rec.Code)
	}

	var response map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["status"] != "unhealthy" {
		t.Errorf("Expected status 'unhealthy', got '%s'", response["status"])
	}

	if response["database"] != "down" {
		t.Errorf("Expected database 'down', got '%s'", response["database"])
	}
}

func TestStatsEndpointsRegistered(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Add some test data to avoid NULL timestamp errors
	db.LogRequest("192.168.1.1", "/test/path")

	router := New(db, "", "")

	tests := []struct {
		name     string
		path     string
		wantCode int
	}{
		{"Summary", "/stats/summary", http.StatusOK},
		{"Endpoints", "/stats/endpoints", http.StatusOK},
		{"Sources", "/stats/sources", http.StatusOK},
		{"Download", "/stats/download", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code != tt.wantCode {
				t.Errorf("Expected status %d for %s, got %d", tt.wantCode, tt.path, rec.Code)
			}

			contentType := rec.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("Expected Content-Type application/json for %s, got %s", tt.path, contentType)
			}
		})
	}
}

func TestDefaultHandlerLogsRequests(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	router := New(db, "", "")

	devNull, _ := os.Open(os.DevNull)
	defer devNull.Close()
	oldStdout := os.Stdout
	os.Stdout = devNull
	defer func() { os.Stdout = oldStdout }()

	req := httptest.NewRequest(http.MethodGet, "/some/random/path", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200 for non-stats path, got %d", rec.Code)
	}

	logs, err := db.GetLogs(10)
	if err != nil {
		t.Fatalf("Failed to get logs: %v", err)
	}

	if len(logs) != 1 {
		t.Errorf("Expected 1 logged request, got %d", len(logs))
	}

	if len(logs) > 0 && logs[0].URL != "/some/random/path" {
		t.Errorf("Expected URL '/some/random/path', got '%s'", logs[0].URL)
	}
}

func TestRouterHandlesMultipleRequests(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	router := New(db, "", "")

	devNull, _ := os.Open(os.DevNull)
	defer devNull.Close()
	oldStdout := os.Stdout
	os.Stdout = devNull
	defer func() { os.Stdout = oldStdout }()

	paths := []string{"/path1", "/path2", "/path3"}

	for _, path := range paths {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Expected status 200 for %s, got %d", path, rec.Code)
		}
	}

	logs, err := db.GetLogs(10)
	if err != nil {
		t.Fatalf("Failed to get logs: %v", err)
	}

	if len(logs) != len(paths) {
		t.Errorf("Expected %d logged requests, got %d", len(paths), len(logs))
	}
}

func TestBasicAuthProtectsStatsEndpoints(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Add test data
	db.LogRequest("192.168.1.1", "/test/path")

	// Create router with auth enabled
	router := New(db, "admin", "secret123")

	t.Run("stats endpoints require auth", func(t *testing.T) {
		endpoints := []string{"/stats/summary", "/stats/endpoints", "/stats/sources", "/stats/download"}

		for _, endpoint := range endpoints {
			req := httptest.NewRequest(http.MethodGet, endpoint, nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code != http.StatusNotFound {
				t.Errorf("Expected 404 for %s without auth, got %d", endpoint, rec.Code)
			}
		}
	})

	t.Run("stats endpoints accessible with valid credentials", func(t *testing.T) {
		endpoints := []string{"/stats/summary", "/stats/endpoints", "/stats/sources", "/stats/download"}
		auth := base64.StdEncoding.EncodeToString([]byte("admin:secret123"))

		for _, endpoint := range endpoints {
			req := httptest.NewRequest(http.MethodGet, endpoint, nil)
			req.Header.Set("Authorization", "Basic "+auth)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Errorf("Expected 200 for %s with valid auth, got %d", endpoint, rec.Code)
			}
		}
	})

	t.Run("health endpoint remains public", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Expected 200 for /health without auth, got %d", rec.Code)
		}
	})

	t.Run("logging endpoint remains public", func(t *testing.T) {
		devNull, _ := os.Open(os.DevNull)
		defer devNull.Close()
		oldStdout := os.Stdout
		os.Stdout = devNull
		defer func() { os.Stdout = oldStdout }()

		req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Expected 200 for /api/test without auth, got %d", rec.Code)
		}
	})
}
