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
	questions  []models.Question
	quizTimers = make(map[string]*time.Timer)
	quizMu     sync.Mutex
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

	// Retrieve the user from the in-memory database
	user, err := s.DB.GetUser(username)
	if err != nil {
		logger.Error("User not found in database", zap.String("username", username), zap.Error(err))
		return fmt.Errorf("user not found: %w", err)
	}

	// Reset user progress and score for the new quiz
	user.Progress = []int{}
	user.Score = 0
	user.QuizTaken++

	// Save the updated user data back to the database
	if err := s.DB.UpdateUser(user); err != nil {
		logger.Error("Failed to update user in database", zap.String("username", username), zap.Error(err))
		return fmt.Errorf("failed to update user: %w", err)
	}

	// Set or reset the quiz session timer for the user
	if timer, exists := quizTimers[username]; exists {
		timer.Stop()
		logger.Warn("Existing quiz session timer stopped", zap.String("username", username))
	}
	quizTimers[username] = time.AfterFunc(10*time.Minute, func() {
		quizMu.Lock()
		defer quizMu.Unlock()

		// Cleanup expired quiz session
		logger.Info("Quiz session expired", zap.String("username", username))
		user.Progress = []int{}
		user.Score = 0
		if err := s.DB.UpdateUser(user); err != nil {
			logger.Error("Failed to reset user progress on quiz expiry", zap.String("username", username), zap.Error(err))
		}
		delete(quizTimers, username)
	})

	logger.Info("Quiz session started successfully", zap.String("username", username))
	return nil
}

func (s *QuizService) GetNextQuestion(username string) (*models.Question, error) {
	logger := utils.GetLogger().Sugar()
	quizMu.Lock()
	defer quizMu.Unlock()

	// Retrieve the user from the in-memory database
	user, err := s.DB.GetUser(username)
	if err != nil {
		logger.Error("User not found in database", zap.String("username", username), zap.Error(err))
		return nil, fmt.Errorf("quiz not started: %w", err)
	}

	// Get user's current progress
	progress := len(user.Progress)

	// Check if there are remaining questions
	if progress >= len(questions) {
		logger.Warn("No more questions available for user", zap.String("username", username))
		return nil, errors.New("quiz complete")
	}

	// Retrieve the next question
	question := questions[progress]
	logger.Info("Next question retrieved", zap.String("username", username), zap.Int("progress", progress))

	// Update user's progress
	user.Progress = append(user.Progress, question.QuestionID)
	if err := s.DB.UpdateUser(user); err != nil {
		logger.Error("Failed to update user progress in database", zap.String("username", username), zap.Error(err))
		return nil, fmt.Errorf("failed to update user progress: %w", err)
	}

	return &question, nil
}

func (s *QuizService) SubmitAnswer(username string, questionIndex, answer int) (bool, error) {
	logger := utils.GetLogger().Sugar()
	quizMu.Lock()
	defer quizMu.Unlock()

	// Validate question index
	if questionIndex < 0 || questionIndex >= len(questions) {
		logger.Error("Invalid question index", zap.Int("questionIndex", questionIndex))
		return false, errors.New("question index is out of range")
	}

	// Retrieve the user from the in-memory database
	user, err := s.DB.GetUser(username)
	if err != nil {
		logger.Error("User not found in database", zap.String("username", username), zap.Error(err))
		return false, fmt.Errorf("user not found: %w", err)
	}

	// Validate the answer
	correctAnswer := questions[questionIndex].Answer
	if answer == correctAnswer {
		user.Score++
		logger.Info("Correct answer submitted", zap.String("username", username), zap.Int("score", user.Score))
	} else {
		logger.Info("Incorrect answer submitted", zap.String("username", username), zap.Int("score", user.Score))
	}

	// Save the updated user data back to the database
	if err := s.DB.UpdateUser(user); err != nil {
		logger.Error("Failed to update user score in database", zap.String("username", username), zap.Error(err))
		return false, fmt.Errorf("failed to update user score: %w", err)
	}

	return answer == correctAnswer, nil
}

func (s *QuizService) GetResults(username string) (int, error) {
	logger := utils.GetLogger().Sugar()
	quizMu.Lock()
	defer quizMu.Unlock()

	// Retrieve the user from the in-memory database
	user, err := s.DB.GetUser(username)
	if err != nil {
		logger.Error("User not found in database", zap.String("username", username), zap.Error(err))
		return 0, fmt.Errorf("user not found: %w", err)
	}

	// Return the user's score
	logger.Info("Final score retrieved", zap.String("username", username), zap.Int("score", user.Score))
	return user.Score, nil
}

func (s *QuizService) GetStats(username string) ([]models.User, string, error) {
	logger := utils.GetLogger().Sugar()
	logger.Info("Fetching stats for user", zap.String("username", username))

	// Retrieve the user from the database
	user, err := s.DB.GetUser(username)
	if err != nil {
		if errors.Is(err, database.ErrUserNotFound) {
			logger.Warn("User not found in database", zap.String("username", username))
			return nil, "", ErrNoStatsForUser
		}
		logger.Error("Database error while fetching user stats", zap.String("username", username), zap.Error(err))
		return nil, "", fmt.Errorf("error fetching user stats: %w", err)
	}

	// Retrieve all users from the database
	allUsers := s.DB.GetAllUsers()
	if len(allUsers) == 0 {
		logger.Warn("No users found in database")
		return nil, "", ErrNoStatsForUser
	}

	// Collect all scores
	allScores := make([]int, len(allUsers))
	for i, u := range allUsers {
		allScores[i] = u.Score
	}

	// Sort scores for ranking
	sort.Ints(allScores)

	// Calculate the percentage of users with lower scores
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

	// Map all users for response
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
