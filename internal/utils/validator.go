package utils

import (
	"errors"
)

func ValidateUsername(username string) error {
	if len(username) == 0 {
		return errors.New("username cannot be empty")
	}
	return nil
}

func ValidateAnswerPayload(questionIndex, answer int) error {
	if questionIndex < 0 {
		return errors.New("question index cannot be negative")
	}
	if answer < 0 {
		return errors.New("answer cannot be negative")
	}
	return nil
}

func ValidatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}
	return nil
}
