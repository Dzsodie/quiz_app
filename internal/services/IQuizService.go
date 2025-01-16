package services

import "github.com/Dzsodie/quiz_app/internal/models"

type IQuizService interface {
	GetQuestions() ([]models.Question, error)

	LoadQuestions(qs []models.Question)

	StartQuiz(username string) error

	GetNextQuestion(username string) (*models.Question, error)

	SubmitAnswer(username string, questionIndex, answer int) (bool, error)

	GetResults(username string) (int, error)
}
