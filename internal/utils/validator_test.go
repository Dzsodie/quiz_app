package utils

import (
	"testing"

	"github.com/Dzsodie/quiz_app/internal/models"
)

func TestValidateAnswerPayload(t *testing.T) {
	questions := []models.Question{
		{Question: "What is 2+2?", Options: []string{"1", "2", "4"}, Answer: 2},
	}

	tests := []struct {
		name          string
		questionIndex int
		answer        int
		expectedErr   bool
	}{
		{"ValidPayload", 0, 2, false},                                   // Valid case
		{"InvalidQuestionIndexNegative", -1, 2, true},                   // Negative question index
		{"InvalidQuestionIndexOutOfRange", 1, 2, true},                  // Out-of-range question index
		{"InvalidAnswerNegative", 0, -1, true},                          // Negative answer
		{"InvalidAnswerOutOfRange", 0, len(questions[0].Options), true}, // Out-of-range answer
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAnswerPayload(tt.questionIndex, tt.answer, questions)
			if (err != nil) != tt.expectedErr {
				t.Errorf("ValidateAnswerPayload() error = %v, expectedErr = %v", err, tt.expectedErr)
			}
		})
	}
}

func TestValidateUsername(t *testing.T) {

	tests := []struct {
		name        string
		username    string
		expectedErr bool
	}{
		{"ValidUsername", "testuser", false},
		{"EmptyUsername", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUsername(tt.username)
			if (err != nil) != tt.expectedErr {
				t.Errorf("ValidateUsername() error = %v, expectedErr = %v", err, tt.expectedErr)
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {

	tests := []struct {
		name             string
		password         string
		previousPassword string
		expectedErr      bool
	}{
		{"ValidPassword", "P@ssw0rd", "", false},
		{"ShortPassword", "short", "", true},
		{"NoUppercase", "p@ssw0rd", "", true},
		{"NoNumber", "Password!", "", true},
		{"NoSpecialChar", "Password1", "", true},
		{"TooSimilarToPrevious", "P@ssw0rd", "P@ssw1rd", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.password, tt.previousPassword)
			if (err != nil) != tt.expectedErr {
				t.Errorf("ValidatePassword() error = %v, expectedErr = %v", err, tt.expectedErr)
			}
		})
	}
}
