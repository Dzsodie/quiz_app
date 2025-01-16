package middleware

import (
	"net/http"

	"github.com/gorilla/sessions"
	"go.uber.org/zap"
)

var (
	sessionStore *sessions.CookieStore
	logger       *zap.Logger
)

func SetSessionStore(store *sessions.CookieStore) {
	sessionStore = store
}

func SetLogger(l *zap.Logger) {
	logger = l
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if logger == nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		session, err := sessionStore.Get(r, "quiz-session")
		if err != nil {
			logger.Warn("Failed to retrieve session", zap.Error(err))
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		username, ok := session.Values["username"].(string)
		if !ok || username == "" {
			logger.Warn("Unauthorized access attempt: no valid username in session")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		logger.Info("User authenticated successfully", zap.String("username", username))
		next.ServeHTTP(w, r)
	})
}
