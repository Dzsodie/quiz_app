package utils

import (
	"errors"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	logger := GetLogger().Sugar()

	if logger == nil {
		panic("Logger is not set for utils")
	}

	if len(password) < 8 {
		logger.Warn("Password hashing failed: password too short")
		return "", errors.New("password must be at least 8 characters")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("Error hashing password", zap.Error(err))
		return "", err
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
