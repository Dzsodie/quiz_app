package utils

import (
	"os"
	"testing"

	"strings"

	"github.com/Dzsodie/quiz_app/internal/models"
)

func TestReadCSV(t *testing.T) {
	t.Run("Valid CSV file", func(t *testing.T) {
		testCSV := "test_valid.csv"
		content := `ID,Question,Option1,Option2,Option3,Answer
1,What is 2+2?,1,2,4,3
2,What is 3+3?,5,6,7,2`

		createTestFile(t, testCSV, content)
		defer os.Remove(testCSV)

		expectedQuestions := []models.Question{
			{QuestionID: 1, Question: "What is 2+2?", Options: []string{"1", "2", "4"}, Answer: 3},
			{QuestionID: 2, Question: "What is 3+3?", Options: []string{"5", "6", "7"}, Answer: 2},
		}

		questions, err := ReadCSV(testCSV)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if len(questions) != len(expectedQuestions) {
			t.Errorf("Expected %d questions, got %d", len(expectedQuestions), len(questions))
		}

		for i, q := range questions {
			if q.QuestionID != expectedQuestions[i].QuestionID ||
				q.Question != expectedQuestions[i].Question ||
				!equalOptions(q.Options, expectedQuestions[i].Options) ||
				q.Answer != expectedQuestions[i].Answer {
				t.Errorf("Mismatch in question %d: got %+v, expected %+v", i+1, q, expectedQuestions[i])
			}
		}
	})

	t.Run("Missing file", func(t *testing.T) {
		_, err := ReadCSV("nonexistent.csv")
		if err == nil || !contains(err.Error(), "failed to open file") {
			t.Errorf("Expected error containing 'failed to open file', got: %v", err)
		}
	})

	t.Run("Empty CSV file", func(t *testing.T) {
		testCSV := "test_empty.csv"
		createTestFile(t, testCSV, "")
		defer os.Remove(testCSV)

		_, err := ReadCSV(testCSV)
		if err == nil || !contains(err.Error(), "CSV file is empty") {
			t.Errorf("Expected error containing 'CSV file is empty', got: %v", err)
		}
	})
}

func createTestFile(t *testing.T, filename, content string) {
	t.Helper()
	file, err := os.Create(filename)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer file.Close()

	if _, err := file.WriteString(content); err != nil {
		t.Fatalf("Failed to write to test file: %v", err)
	}
}

func contains(haystack, needle string) bool {
	return strings.Contains(haystack, needle)
}

func equalOptions(options1, options2 []string) bool {
	if len(options1) != len(options2) {
		return false
	}
	for i := range options1 {
		if options1[i] != options2[i] {
			return false
		}
	}
	return true
}
