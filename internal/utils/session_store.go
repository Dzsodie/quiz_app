package utils

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"sync"

	"github.com/Dzsodie/quiz_app/config"
	"github.com/gorilla/sessions"
)

var (
	SessionStore     *sessions.CookieStore
	sessionStoreOnce sync.Once
	SessionDB        = make(map[string]string) // Store session tokens
)

func InitializeSessionStore() {
	sessionStoreOnce.Do(func() {
		config := config.LoadConfig()
		if config.SessionSecret == "" {
			log.Fatal("Session secret is not set in the configuration")
		}

		SessionStore = sessions.NewCookieStore([]byte(config.SessionSecret))
		SessionStore.Options = &sessions.Options{
			Path:     "/",
			MaxAge:   3600, // 1 hour
			HttpOnly: true,
			Secure:   false, // Set to true if using HTTPS
			SameSite: http.SameSiteStrictMode,
		}
	})
}

func GenerateSessionToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	token := hex.EncodeToString(bytes)

	// Save session token in SessionDB
	SessionDB[token] = ""
	return token, nil
}
