package middleware

import (
	"crypto/subtle"
	"net/http"
)

// BasicAuth returns a middleware that performs HTTP Basic Authentication
func BasicAuth(username, password string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// If no username/password configured, skip auth
			if username == "" || password == "" {
				next.ServeHTTP(w, r)
				return
			}

			user, pass, ok := r.BasicAuth()
			if !ok {
				notFound(w)
				return
			}

			// Use constant-time comparison to prevent timing attacks
			userMatch := subtle.ConstantTimeCompare([]byte(user), []byte(username)) == 1
			passMatch := subtle.ConstantTimeCompare([]byte(pass), []byte(password)) == 1

			if !userMatch || !passMatch {
				notFound(w)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func notFound(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusNotFound)
	if _, err := w.Write([]byte("404 page not found\n")); err != nil {
		// Response already started, can't do much here
	}
}
