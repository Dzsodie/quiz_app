package services

import (
	"errors"
	"sync"
	"time"

	"github.com/Dzsodie/quiz_app/internal/models"
)

type QuizService struct{}

var (
	questions    []models.Question
	userScores   = make(map[string]int)
	userProgress = make(map[string]int)
	quizTimers   = make(map[string]*time.Timer)
	quizMu       sync.Mutex
)

// GetQuestions returns all loaded questions.
func (s *QuizService) GetQuestions() []models.Question {
	quizMu.Lock()
	defer quizMu.Unlock()
	return questions
}

// LoadQuestions initializes the quiz questions.
func (s *QuizService) LoadQuestions(qs []models.Question) {
	quizMu.Lock()
	defer quizMu.Unlock()
	questions = qs
}

// StartQuiz initializes a user's quiz session.
func (s *QuizService) StartQuiz(username string) {
	quizMu.Lock()
	defer quizMu.Unlock()

	userScores[username] = 0
	userProgress[username] = 0

	if timer, exists := quizTimers[username]; exists {
		timer.Stop()
	}

	quizTimers[username] = time.AfterFunc(10*time.Minute, func() {
		quizMu.Lock()
		delete(userProgress, username)
		quizMu.Unlock()
	})
}

// GetNextQuestion retrieves the next question for a user.
func (s *QuizService) GetNextQuestion(username string) (*models.Question, error) {
	quizMu.Lock()
	defer quizMu.Unlock()

	progress, exists := userProgress[username]
	if !exists {
		return nil, errors.New("quiz not started")
	}

	if progress >= len(questions) {
		return nil, errors.New("no more questions")
	}

	question := questions[progress]
	userProgress[username]++
	return &question, nil
}

func (s *QuizService) SubmitAnswer(username string, questionIndex, answer int) (bool, error) {
	quizMu.Lock()
	defer quizMu.Unlock()

	if questionIndex < 0 || questionIndex >= len(questions) {
		return false, errors.New("invalid question index")
	}

	if answer == questions[questionIndex].Answer {
		userScores[username]++
		return true, nil
	}

	return false, nil
}

// GetResults retrieves the user's final score.
func (s *QuizService) GetResults(username string) (int, error) {
	quizMu.Lock()
	defer quizMu.Unlock()

	score, exists := userScores[username]
	if !exists {
		return 0, errors.New("quiz not started")
	}

	return score, nil
}

var _ IQuizService = &QuizService{}
