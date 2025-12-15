package web

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/dangogh/silver-eureka/internal/database"
)

func setupTestDB(t *testing.T) *database.DB {
	dbPath := "/tmp/test_web_" + t.Name() + ".db"
	t.Cleanup(func() {
		if err := os.Remove(dbPath); err != nil {
			// Ignore remove errors in cleanup
		}
	})

	db, err := database.New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			// Ignore close errors in cleanup
		}
	})

	return db
}

func TestHandleLoginPage(t *testing.T) {
	db := setupTestDB(t)
	handler := NewHandler(db, "admin", "secret")

	tests := []struct {
		name           string
		existingCookie *http.Cookie
		wantStatus     int
		wantRedirect   bool
	}{
		{
			name:       "no session cookie",
			wantStatus: http.StatusOK,
		},
		{
			name: "invalid session cookie",
			existingCookie: &http.Cookie{
				Name:  "session_id",
				Value: "invalid-session-id",
			},
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/login", nil)
			if tt.existingCookie != nil {
				req.AddCookie(tt.existingCookie)
			}
			rec := httptest.NewRecorder()

			handler.HandleLoginPage(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("Status = %d, want %d", rec.Code, tt.wantStatus)
			}

			// Check for CSRF cookie
			cookies := rec.Result().Cookies()
			foundCSRF := false
			for _, c := range cookies {
				if c.Name == "csrf_token" {
					foundCSRF = true
					if c.Value == "" {
						t.Error("CSRF cookie has empty value")
					}
					break
				}
			}
			if !foundCSRF && rec.Code == http.StatusOK {
				t.Error("CSRF cookie not set")
			}
		})
	}
}

func TestHandleLoginSubmit(t *testing.T) {
	db := setupTestDB(t)
	handler := NewHandler(db, "admin", "secret")

	tests := []struct {
		name       string
		username   string
		password   string
		csrfToken  string
		csrfCookie string
		wantStatus int
	}{
		{
			name:       "valid credentials with CSRF",
			username:   "admin",
			password:   "secret",
			csrfToken:  "valid-token-123",
			csrfCookie: "valid-token-123",
			wantStatus: http.StatusSeeOther,
		},
		{
			name:       "invalid username",
			username:   "wronguser",
			password:   "secret",
			csrfToken:  "valid-token-123",
			csrfCookie: "valid-token-123",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "invalid password",
			username:   "admin",
			password:   "wrongpass",
			csrfToken:  "valid-token-123",
			csrfCookie: "valid-token-123",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "CSRF token mismatch",
			username:   "admin",
			password:   "secret",
			csrfToken:  "token-123",
			csrfCookie: "token-456",
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "missing CSRF cookie",
			username:   "admin",
			password:   "secret",
			csrfToken:  "token-123",
			csrfCookie: "",
			wantStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			form := url.Values{}
			form.Add("username", tt.username)
			form.Add("password", tt.password)
			form.Add("csrf_token", tt.csrfToken)

			req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			if tt.csrfCookie != "" {
				req.AddCookie(&http.Cookie{
					Name:  "csrf_token",
					Value: tt.csrfCookie,
				})
			}
			rec := httptest.NewRecorder()

			handler.HandleLoginSubmit(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("Status = %d, want %d", rec.Code, tt.wantStatus)
			}

			// For successful login, verify session cookie is set
			if tt.wantStatus == http.StatusSeeOther {
				cookies := rec.Result().Cookies()
				foundSession := false
				for _, c := range cookies {
					if c.Name == "session_id" {
						foundSession = true
						if c.Value == "" {
							t.Error("Session cookie has empty value")
						}
						if c.MaxAge != 86400 {
							t.Errorf("Session MaxAge = %d, want 86400", c.MaxAge)
						}
						break
					}
				}
				if !foundSession {
					t.Error("Session cookie not set after successful login")
				}
			}
		})
	}
}

func TestHandleLogout(t *testing.T) {
	db := setupTestDB(t)
	handler := NewHandler(db, "admin", "secret")

	tests := []struct {
		name       string
		csrfToken  string
		wantStatus int
	}{
		{
			name:       "valid logout",
			csrfToken:  "", // Will use actual CSRF token
			wantStatus: http.StatusSeeOther,
		},
		{
			name:       "invalid CSRF token",
			csrfToken:  "wrong-token",
			wantStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a fresh session for each test
			sessionID, err := handler.sessions.Create("admin")
			if err != nil {
				t.Fatalf("Failed to create session: %v", err)
			}

			session, _ := handler.sessions.Get(sessionID)
			csrfToken := tt.csrfToken
			if csrfToken == "" {
				csrfToken = session.CSRFToken
			}

			form := url.Values{}
			form.Add("csrf_token", csrfToken)

			req := httptest.NewRequest(http.MethodPost, "/logout", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req.AddCookie(&http.Cookie{
				Name:  "session_id",
				Value: sessionID,
			})
			rec := httptest.NewRecorder()

			handler.HandleLogout(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("Status = %d, want %d", rec.Code, tt.wantStatus)
			}

			// For successful logout, verify session is deleted
			if tt.wantStatus == http.StatusSeeOther {
				if _, ok := handler.sessions.Get(sessionID); ok {
					t.Error("Session still exists after logout")
				}

				// Verify cookie is cleared
				cookies := rec.Result().Cookies()
				for _, c := range cookies {
					if c.Name == "session_id" && c.MaxAge != -1 {
						t.Error("Session cookie not cleared")
					}
				}
			}
		})
	}
}

func TestHandleDashboard(t *testing.T) {
	db := setupTestDB(t)
	handler := NewHandler(db, "admin", "secret")

	// Create a session for authenticated access
	sessionID, err := handler.sessions.Create("admin")
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	req.AddCookie(&http.Cookie{
		Name:  "session_id",
		Value: sessionID,
	})
	rec := httptest.NewRecorder()

	handler.HandleDashboard(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", rec.Code, http.StatusOK)
	}

	if !strings.Contains(rec.Body.String(), "Dashboard") {
		t.Error("Response does not contain dashboard content")
	}
}

func TestHandleStatsView(t *testing.T) {
	db := setupTestDB(t)
	handler := NewHandler(db, "admin", "secret")

	// Add some test data
	if err := db.LogRequest("192.168.1.1", "/test1"); err != nil {
		t.Fatalf("Failed to log request: %v", err)
	}
	if err := db.LogRequest("192.168.1.2", "/test2"); err != nil {
		t.Fatalf("Failed to log request: %v", err)
	}

	tests := []struct {
		name       string
		statsType  string
		wantStatus int
	}{
		{"summary stats", "summary", http.StatusOK},
		{"endpoints stats", "endpoints", http.StatusOK},
		{"sources stats", "sources", http.StatusOK},
		{"invalid type", "invalid", http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/stats-view/"+tt.statsType, nil)
			req.SetPathValue("type", tt.statsType)
			rec := httptest.NewRecorder()

			handler.HandleStatsView(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("Status = %d, want %d", rec.Code, tt.wantStatus)
			}
		})
	}
}

func TestRequireAuth(t *testing.T) {
	db := setupTestDB(t)
	handler := NewHandler(db, "admin", "secret")

	// Create a valid session
	sessionID, err := handler.sessions.Create("admin")
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	tests := []struct {
		name       string
		sessionID  string
		wantStatus int
		wantLogged bool
	}{
		{
			name:       "valid session",
			sessionID:  sessionID,
			wantStatus: http.StatusOK,
			wantLogged: false,
		},
		{
			name:       "no session cookie",
			sessionID:  "",
			wantStatus: http.StatusNotFound,
			wantLogged: true,
		},
		{
			name:       "invalid session",
			sessionID:  "invalid-session-id",
			wantStatus: http.StatusNotFound,
			wantLogged: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Count existing logs
			logsBefore, err := db.GetLogs(100)
			if err != nil {
				t.Fatalf("Failed to get logs before: %v", err)
			}
			countBefore := len(logsBefore)

			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				if _, err := w.Write([]byte("success")); err != nil {
					t.Errorf("Failed to write response: %v", err)
				}
			})

			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			if tt.sessionID != "" {
				req.AddCookie(&http.Cookie{
					Name:  "session_id",
					Value: tt.sessionID,
				})
			}
			rec := httptest.NewRecorder()

			handler.RequireAuth(nextHandler)(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("Status = %d, want %d", rec.Code, tt.wantStatus)
			}

			// Check if request was logged
			logsAfter, err := db.GetLogs(100)
			if err != nil {
				t.Fatalf("Failed to get logs after: %v", err)
			}
			countAfter := len(logsAfter)

			if tt.wantLogged && countAfter != countBefore+1 {
				t.Error("Request should have been logged but wasn't")
			}
			if !tt.wantLogged && countAfter != countBefore {
				t.Error("Request was logged but shouldn't have been")
			}
		})
	}
}
