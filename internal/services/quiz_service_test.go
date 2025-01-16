package services

import (
	"sync"
	"testing"

	"github.com/Dzsodie/quiz_app/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestQuizServiceGetQuestions(t *testing.T) {
	s := &QuizService{}
	questions := []models.Question{
		{Question: "What is 2+2?", Options: []string{"3", "4", "5"}, Answer: 1},
		{Question: "What is the capital of France?", Options: []string{"Paris", "Berlin", "Madrid"}, Answer: 0},
	}
	s.LoadQuestions(questions)

	result, err := s.GetQuestions()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assert.NoError(t, err, "expected no error when getting questions")
	assert.Equal(t, questions, result, "expected questions to match loaded questions")
}

func TestQuizServiceStartQuiz(t *testing.T) {
	s := &QuizService{}
	if err := s.StartQuiz("testuser"); err != nil {
		t.Errorf("Failed to start quiz for user 'testuser': %v", err)
	}

	assert.Equal(t, 0, userScores["testuser"], "expected initial score to be 0")
	assert.Equal(t, 0, userProgress["testuser"], "expected initial progress to be 0")
}

func TestQuizServiceGetNextQuestion(t *testing.T) {
	s := &QuizService{}
	questions := []models.Question{
		{Question: "What is 2+2?", Options: []string{"3", "4", "5"}, Answer: 1},
	}
	s.LoadQuestions(questions)
	if err := s.StartQuiz("testuser"); err != nil {
		t.Errorf("Failed to start quiz for user 'testuser': %v", err)
	}

	question, err := s.GetNextQuestion("testuser")
	assert.NoError(t, err, "expected no error when fetching the next question")
	assert.Equal(t, &questions[0], question, "expected question to match the first question")

	_, err = s.GetNextQuestion("testuser")
	assert.Error(t, err, "expected error when no more questions are available")
	assert.Equal(t, "no more questions", err.Error(), "unexpected error message")
}

func TestQuizServiceSubmitAnswer(t *testing.T) {
	s := &QuizService{}
	questions := []models.Question{
		{Question: "What is 2+2?", Options: []string{"3", "4", "5"}, Answer: 1},
	}
	s.LoadQuestions(questions)
	if err := s.StartQuiz("testuser"); err != nil {
		t.Fatalf("Failed to start quiz for user 'testuser': %v", err)
	}

	// Test valid answer
	correct, err := s.SubmitAnswer("testuser", 0, 1)
	assert.NoError(t, err, "expected no error when submitting a valid answer")
	assert.True(t, correct, "expected answer to be marked as correct")

	// Test invalid answer
	correct, err = s.SubmitAnswer("testuser", 0, 0)
	assert.NoError(t, err, "expected no error when submitting an incorrect answer")
	assert.False(t, correct, "expected answer to be marked as incorrect")

	// Test invalid question index
	_, err = s.SubmitAnswer("testuser", 10, 0)
	assert.Error(t, err, "expected error when submitting for an invalid question index")
	assert.Equal(t, "question index is out of range", err.Error(), "unexpected error message")
}

func TestQuizServiceGetResults(t *testing.T) {
	s := &QuizService{}
	if err := s.StartQuiz("testuser"); err != nil {
		t.Errorf("Failed to start quiz for user 'testuser': %v", err)
	}

	userScores["testuser"] = 5

	score, err := s.GetResults("testuser")
	assert.NoError(t, err, "expected no error when retrieving results")
	assert.Equal(t, 5, score, "expected score to match user's score")

	_, err = s.GetResults("nonexistent")
	assert.Error(t, err, "expected error when retrieving results for a non-existent user")
	assert.Equal(t, "quiz not started", err.Error(), "unexpected error message")
}

func TestQuizServiceConcurrency(t *testing.T) {
	s := &QuizService{}
	questions := []models.Question{
		{Question: "What is 2+2?", Options: []string{"3", "4", "5"}, Answer: 1},
	}
	s.LoadQuestions(questions)

	wg := sync.WaitGroup{}
	numRoutines := 50

	for i := 0; i < numRoutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			username := "user" + string(rune(i))
			if err := s.StartQuiz(username); err != nil {
				t.Errorf("Failed to start quiz for user '%s': %v", username, err)
			}

			if _, err := s.SubmitAnswer(username, 0, 1); err != nil {
				t.Errorf("Failed to submit answer for user '%s': %v", username, err)
			}

		}(i)
	}

	wg.Wait()

	for i := 0; i < numRoutines; i++ {
		username := "user" + string(rune(i))
		score, err := s.GetResults(username)
		assert.NoError(t, err, "expected no error for concurrent user")
		assert.Equal(t, 1, score, "expected correct score for concurrent user")
	}
}
