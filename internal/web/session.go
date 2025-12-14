package web

import (
	"sync"
	"time"
)

// Session represents an authenticated user session
type Session struct {
	Username  string
	CSRFToken string
	ExpiresAt time.Time
}

// SessionStore manages user sessions in memory
type SessionStore struct {
	sessions sync.Map
	timeout  time.Duration
}

// NewSessionStore creates a new session store with the given timeout
func NewSessionStore(timeout time.Duration) *SessionStore {
	store := &SessionStore{
		timeout: timeout,
	}

	// Start cleanup goroutine
	go store.cleanupExpired()

	return store
}

// Create creates a new session for the given username
func (s *SessionStore) Create(username string) (string, error) {
	sessionID, err := generateToken()
	if err != nil {
		return "", err
	}

	csrfToken, err := generateToken()
	if err != nil {
		return "", err
	}

	session := Session{
		Username:  username,
		CSRFToken: csrfToken,
		ExpiresAt: time.Now().Add(s.timeout),
	}

	s.sessions.Store(sessionID, session)
	return sessionID, nil
}

// Get retrieves a session by ID
func (s *SessionStore) Get(sessionID string) (Session, bool) {
	val, ok := s.sessions.Load(sessionID)
	if !ok {
		return Session{}, false
	}

	session := val.(Session)

	// Check if expired
	if time.Now().After(session.ExpiresAt) {
		s.sessions.Delete(sessionID)
		return Session{}, false
	}

	return session, true
}

// Delete removes a session
func (s *SessionStore) Delete(sessionID string) {
	s.sessions.Delete(sessionID)
}

// cleanupExpired periodically removes expired sessions
func (s *SessionStore) cleanupExpired() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		s.sessions.Range(func(key, value interface{}) bool {
			session := value.(Session)
			if now.After(session.ExpiresAt) {
				s.sessions.Delete(key)
			}
			return true
		})
	}
}
