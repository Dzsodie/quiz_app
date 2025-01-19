package services

import (
	"sync"
	"testing"

	"github.com/Dzsodie/quiz_app/internal/database"
	"github.com/Dzsodie/quiz_app/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestQuizServiceGetQuestions(t *testing.T) {
	db := database.NewMemoryDB()
	s := NewQuizService(db)

	questions := []models.Question{
		{QuestionID: 1, Question: "What is 2+2?", Options: []string{"3", "4", "5"}, Answer: 1},
		{QuestionID: 2, Question: "What is the capital of France?", Options: []string{"Paris", "Berlin", "Madrid"}, Answer: 0},
	}
	s.LoadQuestions(questions)

	result, err := s.GetQuestions()
	assert.NoError(t, err, "expected no error when getting questions")
	assert.Equal(t, questions, result, "expected questions to match loaded questions")
}

func TestQuizServiceStartQuiz(t *testing.T) {
	db := database.NewMemoryDB()
	s := NewQuizService(db)

	// Add a user to the database
	db.AddUser(database.User{Username: "testuser"})

	err := s.StartQuiz("testuser")
	assert.NoError(t, err, "expected no error when starting a quiz")

	user, err := db.GetUser("testuser")
	assert.NoError(t, err, "expected no error when retrieving user")
	assert.Equal(t, 0, user.Score, "expected initial score to be 0")
	assert.Empty(t, user.Progress, "expected initial progress to be empty")
}

func TestQuizServiceGetNextQuestion(t *testing.T) {
	db := database.NewMemoryDB()
	s := NewQuizService(db)

	questions := []models.Question{
		{QuestionID: 1, Question: "What is 2+2?", Options: []string{"3", "4", "5"}, Answer: 1},
	}
	s.LoadQuestions(questions)

	db.AddUser(database.User{Username: "testuser"})

	err := s.StartQuiz("testuser")
	assert.NoError(t, err, "expected no error when starting a quiz")

	question, err := s.GetNextQuestion("testuser")
	assert.NoError(t, err, "expected no error when fetching the next question")
	assert.Equal(t, &questions[0], question, "expected question to match the first question")

	_, err = s.GetNextQuestion("testuser")
	assert.Error(t, err, "expected error when no more questions are available")
	assert.Equal(t, "quiz complete", err.Error(), "unexpected error message")
}

func TestQuizServiceSubmitAnswer(t *testing.T) {
	db := database.NewMemoryDB()
	s := NewQuizService(db)

	questions := []models.Question{
		{QuestionID: 1, Question: "What is 2+2?", Options: []string{"3", "4", "5"}, Answer: 1},
	}
	s.LoadQuestions(questions)

	db.AddUser(database.User{Username: "testuser"})

	err := s.StartQuiz("testuser")
	assert.NoError(t, err, "expected no error when starting a quiz")

	// Test valid answer
	correct, err := s.SubmitAnswer("testuser", 0, 1)
	assert.NoError(t, err, "expected no error when submitting a valid answer")
	assert.True(t, correct, "expected answer to be marked as correct")

	user, err := db.GetUser("testuser")
	assert.NoError(t, err, "expected no error when retrieving user")
	assert.Equal(t, 1, user.Score, "expected score to be updated after correct answer")

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
	db := database.NewMemoryDB()
	s := NewQuizService(db)

	db.AddUser(database.User{Username: "testuser"})

	err := s.StartQuiz("testuser")
	assert.NoError(t, err, "expected no error when starting a quiz")

	user, err := db.GetUser("testuser")
	assert.NoError(t, err, "expected no error when retrieving user")

	user.Score = 5
	assert.NoError(t, db.UpdateUser(user), "expected no error when updating user score")

	score, err := s.GetResults("testuser")
	assert.NoError(t, err, "expected no error when retrieving results")
	assert.Equal(t, 5, score, "expected score to match user's score")

	_, err = s.GetResults("nonexistent")
	assert.Error(t, err, "expected error when retrieving results for a non-existent user")
	assert.Equal(t, "user not found: user not found", err.Error(), "unexpected error message")
}

func TestQuizServiceConcurrency(t *testing.T) {
	db := database.NewMemoryDB()
	s := NewQuizService(db)

	questions := []models.Question{
		{QuestionID: 1, Question: "What is 2+2?", Options: []string{"3", "4", "5"}, Answer: 1},
	}
	s.LoadQuestions(questions)

	wg := sync.WaitGroup{}
	numRoutines := 50

	for i := 0; i < numRoutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			username := "user" + string(rune(i))
			db.AddUser(database.User{Username: username})

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
