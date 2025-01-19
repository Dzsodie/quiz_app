package utils

import (
	"log"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	logger := GetLogger().Sugar()

	if logger == nil {
		panic("Logger is not set for utils")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	logger.Info("Password hashed successfully")
	return string(hashed), nil
}

func ComparePassword(hashedPassword, plainPassword string) bool {
	logger := GetLogger().Sugar()
	if logger == nil {
		panic("Logger is not set for utils")
	}

	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
	if err != nil {
		logger.Warn("Password comparison failed", zap.Error(err))
		return false
	}

	logger.Info("Password comparison succeeded")
	return true
}
