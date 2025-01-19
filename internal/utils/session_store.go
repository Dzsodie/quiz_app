package utils

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
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

func InitializeSessionStore(config.Config) {
	sessionStoreOnce.Do(func() {

		if config.LoadConfig().SessionSecret == "" {
			log.Fatal("Session secret is not set in the configuration")
		}

		SessionStore = sessions.NewCookieStore([]byte(config.LoadConfig().SessionSecret))
		if SessionStore == nil {
			log.Fatal("Failed to initialize session store")
		}

		SessionStore.Options = &sessions.Options{
			Path:     "/",
			MaxAge:   3600, // 1 hour
			HttpOnly: true,
			Secure:   false,
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

func ValidateSessionStore() error {
	if SessionStore == nil {
		return errors.New("session store is not initialized")
	}
	return nil
}
