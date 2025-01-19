package middleware

import (
	"context"
	"net/http"

	"github.com/Dzsodie/quiz_app/internal/utils"
	"go.uber.org/zap"
)

type contextKey string

const usernameKey contextKey = "username"

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := utils.GetLogger().Sugar()
		logger.Debug("Incoming cookies", zap.String("raw_cookies", r.Header.Get("Cookie")))

		session, err := utils.SessionStore.Get(r, "quiz-session")
		if err != nil {
			logger.Warn("Failed to retrieve session", zap.Error(err))
			http.Error(w, "Invalid session", http.StatusUnauthorized)
			return
		}

		username, ok := session.Values["username"].(string)
		sessionToken, tokenOk := session.Values["session_token"].(string)
		if !ok || !tokenOk || username == "" || sessionToken == "" {
			logger.Warn("Session missing username or token")
			http.Error(w, "Invalid session", http.StatusUnauthorized)
			return
		}

		storedUsername, exists := utils.SessionDB[sessionToken]
		if !exists || storedUsername != username {
			logger.Warn("Session token not found or mismatched")
			http.Error(w, "Invalid session", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), usernameKey, username)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
