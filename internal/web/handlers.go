package web

import (
	"crypto/subtle"
	"embed"
	"encoding/json"
	"html/template"
	"log/slog"
	"net/http"
	"time"

	"github.com/dangogh/silver-eureka/internal/database"
)

//go:embed templates/*.html
var templatesFS embed.FS

const sessionCookieName = "session_id"

// Handler manages web interface requests
type Handler struct {
	db           *database.DB
	sessions     *SessionStore
	templates    *template.Template
	authUsername string
	authPassword string
}

// NewHandler creates a new web interface handler
func NewHandler(db *database.DB, authUsername, authPassword string) *Handler {
	funcMap := template.FuncMap{
		"mul": func(a, b int) int { return a * b },
		"div": func(a, b int) int {
			if b == 0 {
				return 0
			}
			return a / b
		},
	}
	tmpl := template.Must(template.New("").Funcs(funcMap).ParseFS(templatesFS, "templates/*.html"))

	return &Handler{
		db:           db,
		sessions:     NewSessionStore(24 * time.Hour),
		templates:    tmpl,
		authUsername: authUsername,
		authPassword: authPassword,
	}
}

// HandleLoginPage displays the login form
func (h *Handler) HandleLoginPage(w http.ResponseWriter, r *http.Request) {
	// Check if already logged in
	if cookie, err := r.Cookie(sessionCookieName); err == nil {
		if _, ok := h.sessions.Get(cookie.Value); ok {
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
			return
		}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.templates.ExecuteTemplate(w, "login.html", nil); err != nil {
		slog.Error("Failed to render login template", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// HandleLoginSubmit processes login form submission
func (h *Handler) HandleLoginSubmit(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	// Validate credentials using constant-time comparison
	userMatch := subtle.ConstantTimeCompare([]byte(username), []byte(h.authUsername)) == 1
	passMatch := subtle.ConstantTimeCompare([]byte(password), []byte(h.authPassword)) == 1

	if !userMatch || !passMatch {
		time.Sleep(100 * time.Millisecond) // Prevent timing attacks
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("401 unauthorized\n"))
		return
	}

	// Create session
	sessionID, err := h.sessions.Create(username)
	if err != nil {
		slog.Error("Failed to create session", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Set cookie
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    sessionID,
		Path:     "/",
		MaxAge:   86400, // 24 hours
		HttpOnly: true,
		Secure:   false, // Set to true if using HTTPS
		SameSite: http.SameSiteStrictMode,
	})

	slog.Info("User logged in", "username", username)
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

// HandleLogout logs out the user
func (h *Handler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(sessionCookieName)
	if err == nil {
		h.sessions.Delete(cookie.Value)
	}

	// Clear cookie
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// HandleDashboard displays the main dashboard
func (h *Handler) HandleDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.templates.ExecuteTemplate(w, "dashboard.html", nil); err != nil {
		slog.Error("Failed to render dashboard template", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// HandleStatsView displays stats in HTML format
func (h *Handler) HandleStatsView(w http.ResponseWriter, r *http.Request) {
	statsType := r.PathValue("type")

	var data interface{}
	var err error
	var title string

	switch statsType {
	case "summary":
		title = "Summary Statistics"
		data, err = h.db.GetSummary()
	case "endpoints":
		title = "Endpoint Statistics"
		data, err = h.db.GetEndpointStats()
	case "sources":
		title = "Source IP Statistics"
		data, err = h.db.GetSourceStats()
	default:
		http.NotFound(w, r)
		return
	}

	if err != nil {
		slog.Error("Failed to retrieve stats", "type", statsType, "error", err)
		http.Error(w, "Failed to retrieve statistics", http.StatusInternalServerError)
		return
	}

	// Check if client wants JSON
	if r.Header.Get("Accept") == "application/json" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(data)
		return
	}

	// Calculate max values for bar chart scaling (endpoints only)
	var maxCount, maxUniqueIPs int64
	if statsType == "endpoints" {
		if endpoints, ok := data.([]database.EndpointStats); ok {
			for _, ep := range endpoints {
				if ep.Count > maxCount {
					maxCount = ep.Count
				}
				if ep.UniqueIPs > maxUniqueIPs {
					maxUniqueIPs = ep.UniqueIPs
				}
			}
		}
	}

	// Render HTML
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	templateData := map[string]interface{}{
		"Title":        title,
		"Type":         statsType,
		"Data":         data,
		"MaxCount":     maxCount,
		"MaxUniqueIPs": maxUniqueIPs,
	}

	if err := h.templates.ExecuteTemplate(w, "stats.html", templateData); err != nil {
		slog.Error("Failed to render stats template", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// getIPAddress extracts the client IP address from the request
func getIPAddress(r *http.Request) string {
	// Check X-Forwarded-For header first (for proxied requests)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For can contain multiple IPs; take the first one
		if idx := len(xff); idx > 0 {
			if commaIdx := 0; commaIdx < idx {
				for i, c := range xff {
					if c == ',' {
						return xff[:i]
					}
				}
			}
			return xff
		}
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}

// RequireAuth is middleware that ensures user is authenticated
func (h *Handler) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(sessionCookieName)
		if err != nil {
			// Log the request before returning 404
			if err := h.db.LogRequest(getIPAddress(r), r.URL.Path); err != nil {
				slog.Error("Failed to log request", "error", err)
			}
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("404 page not found\n"))
			return
		}

		_, ok := h.sessions.Get(cookie.Value)
		if !ok {
			// Log the request before returning 404
			if err := h.db.LogRequest(getIPAddress(r), r.URL.Path); err != nil {
				slog.Error("Failed to log request", "error", err)
			}
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("404 page not found\n"))
			return
		}

		next(w, r)
	}
}
