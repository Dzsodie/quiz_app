package middleware

import (
	"net/http"

	"github.com/Dzsodie/quiz_app/internal/utils"
	"go.uber.org/zap"
)

func AuthMiddleware(next http.Handler) http.Handler {
	logger := utils.GetLogger().Sugar()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Debug("Incoming cookies", zap.String("raw_cookies", r.Header.Get("Cookie")))

		session, err := utils.SessionStore.Get(r, "quiz-session")
		if err != nil {
			logger.Warn("Failed to retrieve session", zap.Error(err))
			http.Error(w, `{"message":"Invalid session"}`, http.StatusUnauthorized)
			return
		}

		logger.Debug("Session retrieved", zap.Any("session_values", session.Values))

		username, ok := session.Values["username"].(string)
		if !ok || username == "" {
			logger.Warn("Session missing username or invalid format")
			http.Error(w, `{"message":"Invalid session"}`, http.StatusUnauthorized)
			return
		}

		logger.Info("Session validated successfully", zap.String("username", username))
		next.ServeHTTP(w, r)
	})
}
