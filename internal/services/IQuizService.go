package services

import "github.com/Dzsodie/quiz_app/internal/models"

// IQuizService defines the contract for quiz-related operations.
type IQuizService interface {
	// GetQuestions returns all loaded questions.
	GetQuestions() []models.Question

	// LoadQuestions initializes the quiz questions.
	LoadQuestions(qs []models.Question)

	// StartQuiz initializes a user's quiz session.
	StartQuiz(username string)

	// GetNextQuestion retrieves the next question for a user.
	// Returns an error if the quiz is not started or no more questions are available.
	GetNextQuestion(username string) (*models.Question, error)

	// SubmitAnswer evaluates a user's answer and updates the score.
	// Returns true if the answer is correct and an error if the question index is invalid.
	SubmitAnswer(username string, questionIndex, answer int) (bool, error)

	// GetResults retrieves the user's final score.
	// Returns an error if the quiz is not started.
	GetResults(username string) (int, error)
}
