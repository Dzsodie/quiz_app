package utils

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/Dzsodie/quiz_app/internal/models"
	"go.uber.org/zap"
)

func ValidateAnswerPayload(questionIndex, answer int, questions []models.Question) error {
	if Logger == nil {
		panic("Logger is not set for utils")
	}

	if questionIndex < 0 || questionIndex >= len(questions) {
		Logger.Warn("Validation failed: question index out of range", zap.Int("questionIndex", questionIndex))
		return errors.New("question index is out of range")
	}

	optionsCount := len(questions[questionIndex].Options)
	if answer < 0 || answer >= optionsCount {
		Logger.Warn("Validation failed: answer out of range", zap.Int("answer", answer), zap.Int("maxOptions", optionsCount-1))
		return fmt.Errorf("answer must be between 0 and %d", optionsCount-1)
	}

	Logger.Info("Answer payload validated successfully", zap.Int("questionIndex", questionIndex), zap.Int("answer", answer))
	return nil
}

func ValidateUsername(username string) error {
	if Logger == nil {
		panic("Logger is not set for utils")
	}

	if len(username) == 0 {
		Logger.Warn("Validation failed: username is empty")
		return errors.New("username cannot be empty")
	}

	Logger.Info("Username validated successfully", zap.String("username", username))
	return nil
}

func ValidatePassword(password, previousPassword string) error {
	if Logger == nil {
		panic("Logger is not set for utils")
	}

	if len(password) < 8 {
		Logger.Warn("Validation failed: password too short")
		return errors.New("password must be at least 8 characters long")
	}

	hasUppercase := regexp.MustCompile(`[A-Z]`).MatchString(password)
	if !hasUppercase {
		Logger.Warn("Validation failed: password missing uppercase letter")
		return errors.New("password must contain at least one uppercase letter")
	}

	hasNumber := regexp.MustCompile(`\d`).MatchString(password)
	if !hasNumber {
		Logger.Warn("Validation failed: password missing number")
		return errors.New("password must contain at least one number")
	}

	hasSpecial := regexp.MustCompile(`[!@#$%^&*(),.?":{}|<>]`).MatchString(password)
	if !hasSpecial {
		Logger.Warn("Validation failed: password missing special character")
		return errors.New("password must contain at least one special character")
	}

	if previousPassword != "" {
		differenceCount := 0
		maxLen := len(password)
		if len(previousPassword) > maxLen {
			maxLen = len(previousPassword)
		}
		for i := 0; i < maxLen; i++ {
			if i >= len(password) || i >= len(previousPassword) || password[i] != previousPassword[i] {
				differenceCount++
			}
		}
		if differenceCount < 2 {
			Logger.Warn("Validation failed: password too similar to previous password")
			return errors.New("password must differ from the previous password by at least 2 characters")
		}
	}

	Logger.Info("Password validated successfully")
	return nil
}
