package database

import (
	"github.com/Dzsodie/quiz_app/internal/models"
)

type User models.User
type Question models.Question

type QuizDatabase interface {
	AddUser(user User) error
	GetUser(username string) (User, error)
	GetAllUsers() []User

	GetQuestion(id string) (Question, error)
	ListQuestions() ([]Question, error)
}
