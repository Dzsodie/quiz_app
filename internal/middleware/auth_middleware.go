package middleware

import (
	"net/http"

	"github.com/gorilla/sessions"
)

var sessionStore *sessions.CookieStore

// SetSessionStore sets the session store to be used by the middleware.
func SetSessionStore(store *sessions.CookieStore) {
	sessionStore = store
}

// AuthMiddleware ensures that a user is authenticated before accessing certain routes.
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := sessionStore.Get(r, "quiz-session")
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		username, ok := session.Values["username"].(string)
		if !ok || username == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
