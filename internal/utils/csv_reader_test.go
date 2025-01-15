package utils

import (
	"os"
	"testing"

	"github.com/Dzsodie/quiz_app/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestReadCSV(t *testing.T) {
	tests := []struct {
		name           string
		setupFile      func() string
		expectedResult []models.Question
		expectedError  string
	}{
		{
			name: "Valid CSV file",
			setupFile: func() string {
				tmpFile, err := os.CreateTemp("", "test_valid*.csv")
				assert.NoError(t, err)
				defer tmpFile.Close()

				_, err = tmpFile.WriteString(`Question,Option1,Option2,Option3,Answer
What is 2+2?,1,2,4,2
What is 3+3?,5,6,7,1
`)
				assert.NoError(t, err)
				return tmpFile.Name()
			},
			expectedResult: []models.Question{
				{Question: "What is 2+2?", Options: []string{"1", "2", "4"}, Answer: 2},
				{Question: "What is 3+3?", Options: []string{"5", "6", "7"}, Answer: 1},
			},
			expectedError: "",
		},
		{
			name: "Missing file",
			setupFile: func() string {
				return "nonexistent.csv"
			},
			expectedResult: nil,
			expectedError:  "failed to open file",
		},
		{
			name: "Malformed CSV file",
			setupFile: func() string {
				tmpFile, err := os.CreateTemp("", "test_invalid*.csv")
				assert.NoError(t, err)
				defer tmpFile.Close()

				_, err = tmpFile.WriteString(`Question,Option1,Option2,Option3,Answer
What is 2+2?,1,2,4,invalid_answer
`)
				assert.NoError(t, err)
				return tmpFile.Name()
			},
			expectedResult: nil,
			expectedError:  "invalid answer format in record",
		},
		{
			name: "Empty CSV file",
			setupFile: func() string {
				tmpFile, err := os.CreateTemp("", "test_empty*.csv")
				assert.NoError(t, err)
				defer tmpFile.Close() // Ensure the file is created but empty
				return tmpFile.Name()
			},
			expectedResult: nil,
			expectedError:  "CSV file is empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create the file using setupFile
			filename := tt.setupFile()
			defer os.Remove(filename) // Cleanup the file after test

			// Call ReadCSV
			result, err := ReadCSV(filename)

			// Validate results
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}
		})
	}
}
