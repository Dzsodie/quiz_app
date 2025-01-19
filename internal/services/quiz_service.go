package services

import (
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/Dzsodie/quiz_app/internal/database"
	"github.com/Dzsodie/quiz_app/internal/models"
	"github.com/Dzsodie/quiz_app/internal/utils"
	"go.uber.org/zap"
)

type QuizService struct {
	DB *database.MemoryDB
}

func NewQuizService(db *database.MemoryDB) *QuizService {
	return &QuizService{DB: db}
}

var (
	questions    []models.Question
	userScores   = make(map[string]int)
	userProgress = make(map[string]int)
	quizTimers   = make(map[string]*time.Timer)
	quizMu       sync.Mutex
)
var ErrNoStatsForUser = errors.New("no stats available for user")

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
		defer quizMu.Unlock()
		delete(userProgress, username)
		delete(userScores, username)
		delete(quizTimers, username)
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

	if answer < 0 || answer > 4 {
		logger.Warn("Invalid answer provided", zap.Int("answer", answer))
		return false, errors.New("answer must be 1, 2, or 3")
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

func (s *QuizService) GetStats(username string) ([]models.User, string, error) {
	logger := utils.GetLogger().Sugar()
	logger.Info("Fetching stats for user", zap.String("username", username))

	user, err := s.DB.GetUser(username)
	if err != nil {
		if errors.Is(err, database.ErrUserNotFound) {
			logger.Warn("User not found in database", zap.String("username", username))
			return nil, "", ErrNoStatsForUser
		}
		logger.Error("Database error while fetching user stats", zap.String("username", username), zap.Error(err))
		return nil, "", fmt.Errorf("error fetching user stats: %w", err)
	}

	allUsers := s.DB.GetAllUsers()
	if len(allUsers) == 0 {
		logger.Warn("No users found in database")
		return nil, "", ErrNoStatsForUser
	}

	allScores := make([]int, len(allUsers))
	for i, u := range allUsers {
		allScores[i] = u.Score
	}
	sort.Ints(allScores)

	betterScores := 0
	for _, score := range allScores {
		if user.Score > score {
			betterScores++
		}
	}

	percentage := (float64(betterScores) / float64(len(allScores))) * 100
	message := fmt.Sprintf(
		"Your score is %d and that is %.2f%% better than other users' scores.",
		user.Score, percentage,
	)

	logger.Info("Stats calculated successfully",
		zap.String("username", username),
		zap.Int("score", user.Score),
		zap.Float64("better_than_percentage", percentage),
	)

	modelUsers := make([]models.User, len(allUsers))
	for i, u := range allUsers {
		modelUsers[i] = models.User{
			Username: u.Username,
			Score:    u.Score,
		}
	}

	return modelUsers, message, nil
}

var _ IQuizService = &QuizService{}
