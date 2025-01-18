package utils

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/gorilla/sessions"
)

var SessionStore *sessions.CookieStore

func GenerateSessionToken() (string, error) {

	bytes := make([]byte, 32)

	if _, err := rand.Read(bytes); err != nil {

		return "", err

	}

	return hex.EncodeToString(bytes), nil

}
