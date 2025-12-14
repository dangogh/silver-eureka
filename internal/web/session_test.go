package web

import (
	"testing"
	"time"
)

func TestSessionStore_CreateAndGet(t *testing.T) {
	store := NewSessionStore(1 * time.Hour)

	tests := []struct {
		name     string
		username string
	}{
		{"valid user", "testuser"},
		{"admin user", "admin"},
		{"user with spaces", "test user"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sessionID, err := store.Create(tt.username)
			if err != nil {
				t.Fatalf("Create() error = %v", err)
			}
			if sessionID == "" {
				t.Error("Create() returned empty session ID")
			}

			// Verify session can be retrieved
			session, ok := store.Get(sessionID)
			if !ok {
				t.Error("Get() failed to retrieve session")
			}
			if session.Username != tt.username {
				t.Errorf("Username = %q, want %q", session.Username, tt.username)
			}
			if session.CSRFToken == "" {
				t.Error("CSRF token is empty")
			}
			if session.ExpiresAt.Before(time.Now()) {
				t.Error("Session already expired")
			}
		})
	}
}

func TestSessionStore_GetExpired(t *testing.T) {
	store := NewSessionStore(10 * time.Millisecond)
	sessionID, err := store.Create("testuser")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// Wait for session to expire
	time.Sleep(50 * time.Millisecond)
	_, ok := store.Get(sessionID)
	if ok {
		t.Error("Get() returned expired session")
	}
}

func TestSessionStore_GetNonExistent(t *testing.T) {
	store := NewSessionStore(1 * time.Hour)
	_, ok := store.Get("nonexistent-session-id")
	if ok {
		t.Error("Get() returned true for non-existent session")
	}
}

func TestSessionStore_Delete(t *testing.T) {
	store := NewSessionStore(1 * time.Hour)
	sessionID, err := store.Create("testuser")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// Verify session exists
	_, ok := store.Get(sessionID)
	if !ok {
		t.Error("Session not found before delete")
	}

	// Delete session
	store.Delete(sessionID)

	// Verify session is gone
	_, ok = store.Get(sessionID)
	if ok {
		t.Error("Get() returned deleted session")
	}
}

func TestSessionStore_CleanupExpired(t *testing.T) {
	store := NewSessionStore(50 * time.Millisecond)

	// Create multiple sessions
	ids := make([]string, 3)
	for i := 0; i < 3; i++ {
		id, err := store.Create("user")
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		ids[i] = id
	}

	// Wait for sessions to expire
	time.Sleep(100 * time.Millisecond)

	// Verify all expired sessions are not returned by Get()
	for _, id := range ids {
		if _, ok := store.Get(id); ok {
			t.Error("Get() returned expired session")
		}
	}
}

func TestGenerateToken(t *testing.T) {
	tokens := make(map[string]bool)

	// Generate multiple tokens and verify uniqueness
	for i := 0; i < 100; i++ {
		token, err := generateToken()
		if err != nil {
			t.Fatalf("generateToken() error = %v", err)
		}
		if token == "" {
			t.Error("generateToken() returned empty string")
		}
		if len(token) != 64 { // 32 bytes -> 64 hex chars
			t.Errorf("Token length = %d, want 64", len(token))
		}
		if tokens[token] {
			t.Error("generateToken() returned duplicate token")
		}
		tokens[token] = true
	}
}
