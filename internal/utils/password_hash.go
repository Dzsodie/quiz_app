package utils

import (
	"errors"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// HashPassword hashes a password using bcrypt.
func HashPassword(password string) (string, error) {
	if Logger == nil {
		panic("Logger is not set for utils")
	}

	if len(password) < 8 {
		Logger.Warn("Password hashing failed: password too short")
		return "", errors.New("password must be at least 8 characters")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		Logger.Error("Error hashing password", zap.Error(err))
		return "", err
	}

	Logger.Info("Password hashed successfully")
	return string(hashed), nil
}

// ComparePassword compares a hashed password with a plain password.
func ComparePassword(hashedPassword, plainPassword string) bool {
	if Logger == nil {
		panic("Logger is not set for utils")
	}

	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
	if err != nil {
		Logger.Warn("Password comparison failed", zap.Error(err))
		return false
	}

	Logger.Info("Password comparison succeeded")
	return true
}
