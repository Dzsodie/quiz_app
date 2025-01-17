package services

import (
	"errors"
	"sync"
	"time"

	"github.com/Dzsodie/quiz_app/internal/models"
	"github.com/Dzsodie/quiz_app/internal/utils"
	"go.uber.org/zap"
)

type QuizService struct{}

var (
	questions    []models.Question
	userScores   = make(map[string]int)
	userProgress = make(map[string]int)
	quizTimers   = make(map[string]*time.Timer)
	quizMu       sync.Mutex
)

func (s *QuizService) GetQuestions() ([]models.Question, error) {
	logger := utils.GetLogger().Sugar()
	quizMu.Lock()
	defer quizMu.Unlock()

	if questions == nil {
		logger.Warn("Attempted to get questions but none are available")
		return nil, errors.New("no questions available")
	}
	logger.Info("Questions retrieved successfully", zap.Int("count", len(questions)))
	return questions, nil
}

func (s *QuizService) LoadQuestions(qs []models.Question) {
	logger := utils.GetLogger().Sugar()
	logger.Info("Loading questions into QuizService", zap.Int("question_count", len(qs)))

	quizMu.Lock()
	defer quizMu.Unlock()

	questions = qs
	logger.Info("Questions loaded successfully", zap.Int("count", len(qs)))
}

func (s *QuizService) StartQuiz(username string) error {
	logger := utils.GetLogger().Sugar()
	quizMu.Lock()
	defer quizMu.Unlock()

	userScores[username] = 0
	userProgress[username] = 0

	if timer, exists := quizTimers[username]; exists {
		timer.Stop()
		logger.Warn("Existing quiz session timer stopped", zap.String("username", username))
	}

	quizTimers[username] = time.AfterFunc(10*time.Minute, func() {
		quizMu.Lock()
		delete(userProgress, username)
		delete(userScores, username)
		delete(quizTimers, username)
		quizMu.Unlock()
		logger.Info("Quiz session expired", zap.String("username", username))
	})
	logger.Info("Quiz session started", zap.String("username", username))
	return nil
}

func (s *QuizService) GetNextQuestion(username string) (*models.Question, error) {
	logger := utils.GetLogger().Sugar()
	quizMu.Lock()
	defer quizMu.Unlock()

	progress, exists := userProgress[username]
	if !exists {
		logger.Error("Quiz not started for user", zap.String("username", username))
		return nil, errors.New("quiz not started")
	}

	if progress >= len(questions) {
		logger.Warn("No more questions available for user", zap.String("username", username))
		return nil, errors.New("quiz complete")
	}

	question := questions[progress]
	logger.Info("Next question retrieved", zap.String("username", username), zap.Int("progress", userProgress[username]))
	userProgress[username]++
	return &question, nil
}

func (s *QuizService) SubmitAnswer(username string, questionIndex, answer int) (bool, error) {
	logger := utils.GetLogger().Sugar()
	quizMu.Lock()
	defer quizMu.Unlock()

	if questionIndex < 0 || questionIndex >= len(questions) {
		logger.Error("Invalid question index", zap.Int("questionIndex", questionIndex))
		return false, errors.New("question index is out of range")
	}

	correctAnswer := questions[questionIndex].Answer
	if answer == correctAnswer {
		userScores[username]++
		logger.Info("Correct answer submitted", zap.String("username", username), zap.Int("score", userScores[username]))
		return true, nil
	} else {
		logger.Info("Incorrect answer submitted", zap.String("username", username), zap.Int("score", userScores[username]))
		return false, nil
	}
}

func (s *QuizService) GetResults(username string) (int, error) {
	logger := utils.GetLogger().Sugar()
	quizMu.Lock()
	defer quizMu.Unlock()

	score, exists := userScores[username]
	if !exists {
		logger.Error("Quiz not started for user", zap.String("username", username))
		return 0, errors.New("quiz not started")
	}
	logger.Info("Final score retrieved", zap.String("username", username), zap.Int("score", score))
	return score, nil
}

var _ IQuizService = &QuizService{}
