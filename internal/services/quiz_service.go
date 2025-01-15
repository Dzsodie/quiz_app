package services

import (
	"errors"
	"sync"
	"time"

	"github.com/Dzsodie/quiz_app/internal/models"
	"github.com/Dzsodie/quiz_app/internal/utils"
	"go.uber.org/zap"
)

type QuizService struct {
	Logger *zap.Logger
}

var (
	questions    []models.Question
	userScores   = make(map[string]int)
	userProgress = make(map[string]int)
	quizTimers   = make(map[string]*time.Timer)
	quizMu       sync.Mutex
)

func NewQuizService(logger *zap.Logger) *QuizService {
	return &QuizService{Logger: logger}
}

// GetQuestions returns all loaded questions.
func (s *QuizService) GetQuestions() ([]models.Question, error) {
	quizMu.Lock()
	defer quizMu.Unlock()

	if questions == nil {
		s.Logger.Warn("Attempted to get questions but none are available")
		return nil, errors.New("no questions available")
	}
	s.Logger.Info("Questions retrieved successfully", zap.Int("count", len(questions)))
	return questions, nil
}

// LoadQuestions initializes the quiz questions.
func (s *QuizService) LoadQuestions(qs []models.Question) {
	if s.Logger == nil {
		panic("QuizService logger is not set")
	}
	s.Logger.Info("Loading questions into QuizService", zap.Int("question_count", len(questions)))

	quizMu.Lock()
	defer quizMu.Unlock()

	questions = qs
	s.Logger.Info("Questions loaded successfully", zap.Int("count", len(qs)))
}

// StartQuiz initializes a user's quiz session.
func (s *QuizService) StartQuiz(username string) error {
	quizMu.Lock()
	defer quizMu.Unlock()

	userScores[username] = 0
	userProgress[username] = 0

	if timer, exists := quizTimers[username]; exists {
		timer.Stop()
		s.Logger.Warn("Existing quiz session timer stopped", zap.String("username", username))
	}

	quizTimers[username] = time.AfterFunc(10*time.Minute, func() {
		quizMu.Lock()
		delete(userProgress, username)
		delete(userScores, username)
		delete(quizTimers, username)
		quizMu.Unlock()
		s.Logger.Info("Quiz session expired", zap.String("username", username))
	})
	s.Logger.Info("Quiz session started", zap.String("username", username))
	return nil
}

// GetNextQuestion retrieves the next question for a user.
func (s *QuizService) GetNextQuestion(username string) (*models.Question, error) {
	quizMu.Lock()
	defer quizMu.Unlock()

	progress, exists := userProgress[username]
	if !exists {
		s.Logger.Error("Quiz not started for user", zap.String("username", username))
		return nil, errors.New("quiz not started")
	}

	if progress >= len(questions) {
		s.Logger.Warn("No more questions available for user", zap.String("username", username))
		return nil, errors.New("no more questions")
	}

	question := questions[progress]
	userProgress[username]++
	s.Logger.Info("Next question retrieved", zap.String("username", username), zap.Int("progress", userProgress[username]))
	return &question, nil
}

func (s *QuizService) SubmitAnswer(username string, questionIndex, answer int) (bool, error) {
	quizMu.Lock()
	defer quizMu.Unlock()

	if err := utils.ValidateAnswerPayload(questionIndex, answer, questions); err != nil {
		s.Logger.Error("Invalid answer payload", zap.String("username", username), zap.Error(err))
		return false, err
	}

	if answer == questions[questionIndex].Answer {
		userScores[username]++
		s.Logger.Info("Correct answer submitted", zap.String("username", username), zap.Int("score", userScores[username]))
		return true, nil
	}

	s.Logger.Info("Incorrect answer submitted", zap.String("username", username), zap.Int("score", userScores[username]))
	return false, nil
}

// GetResults retrieves the user's final score.
func (s *QuizService) GetResults(username string) (int, error) {
	quizMu.Lock()
	defer quizMu.Unlock()

	score, exists := userScores[username]
	if !exists {
		s.Logger.Error("Quiz not started for user", zap.String("username", username))
		return 0, errors.New("quiz not started")
	}
	s.Logger.Info("Final score retrieved", zap.String("username", username), zap.Int("score", score))
	return score, nil
}

var _ IQuizService = &QuizService{}
