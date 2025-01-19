package utils

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"sync"

	"github.com/Dzsodie/quiz_app/config"
	"github.com/gorilla/sessions"
)

var (
	SessionStore     *sessions.CookieStore
	sessionStoreOnce sync.Once
)

func InitializeSessionStore() {
	sessionStoreOnce.Do(func() {
		// Load session secret from configuration
		config := config.LoadConfig()
		if config.SessionSecret == "" {
			log.Fatal("Session secret is not set in the configuration")
		}

		// Log the secret for debugging
		logger := GetLogger().Sugar()
		logger.Debug("Initializing SessionStore with secret: ", config.SessionSecret)

		// Initialize the session store with the secret key
		SessionStore = sessions.NewCookieStore([]byte(config.SessionSecret))
		SessionStore.Options = &sessions.Options{
			Path:     "/",
			MaxAge:   3600, // 1 hour
			HttpOnly: true,
		}
	})
}

func GenerateSessionToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
