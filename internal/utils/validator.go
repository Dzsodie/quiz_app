package utils

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/Dzsodie/quiz_app/internal/models"
)

func ValidateUsername(username string) error {
	if len(username) == 0 {
		return errors.New("username cannot be empty")
	}
	return nil
}

func ValidateAnswerPayload(questionIndex, answer int, questions []models.Question) error {
	if questionIndex < 0 || questionIndex >= len(questions) {
		return errors.New("question index is out of range")
	}

	optionsCount := len(questions[questionIndex].Options)
	if answer < 0 || answer >= optionsCount {
		return fmt.Errorf("answer must be between 0 and %d", optionsCount-1)
	}

	return nil
}

func ValidatePassword(password, previousPassword string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}

	hasUppercase := regexp.MustCompile(`[A-Z]`).MatchString(password)
	if !hasUppercase {
		return errors.New("password must contain at least one uppercase letter")
	}

	hasNumber := regexp.MustCompile(`\d`).MatchString(password)
	if !hasNumber {
		return errors.New("password must contain at least one number")
	}

	hasSpecial := regexp.MustCompile(`[!@#$%^&*(),.?":{}|<>]`).MatchString(password)
	if !hasSpecial {
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
			return errors.New("password must differ from the previous password by at least 2 characters")
		}
	}

	return nil

}
