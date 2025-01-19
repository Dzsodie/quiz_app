package handlers

import (
	"encoding/json"
	"net/http"

	"errors"

	"github.com/Dzsodie/quiz_app/internal/models"
	"github.com/Dzsodie/quiz_app/internal/services"
	"github.com/Dzsodie/quiz_app/internal/utils"
	"go.uber.org/zap"
)

type QuizHandler struct {
	QuizService services.IQuizService
}

func NewQuizHandler(quizService services.IQuizService) *QuizHandler {
	return &QuizHandler{QuizService: quizService}
}

// GetQuestions retrieves all available quiz questions
// @Summary Get all quiz questions
// @Description Fetches all quiz questions available in the system
// @Tags Quiz
// @Produce json
// @Success 200 {array} models.Question "List of questions"
// @Failure 500 {string} string "Internal server error"
// @Router /questions [get]
func (h *QuizHandler) GetQuestions(w http.ResponseWriter, r *http.Request) {
	logger := utils.GetLogger().Sugar()
	w.Header().Set("Content-Type", "application/json")
	allQuestions, err := h.QuizService.GetQuestions()
	if err != nil {
		logger.Error("Failed to retrieve questions", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	logger.Info("Questions retrieved successfully")
	if err := json.NewEncoder(w).Encode(allQuestions); err != nil {
		logger.Warn("Failed to encode questions response", zap.Error(err))
	}
}

// StartQuiz starts a new quiz session for the user
// @Summary Start a quiz
// @Description Initiates a quiz session for the logged-in user
// @Tags Quiz
// @Produce json
// @Success 200 {object} map[string]string "Quiz started and next endpoint"
// @Failure 401 {string} string "Invalid session"
// @Failure 500 {string} string "Internal server error"
// @Router /quiz/start [post]
func (h *QuizHandler) StartQuiz(w http.ResponseWriter, r *http.Request) {
	logger := utils.GetLogger().Sugar()
	session, err := utils.SessionStore.Get(r, "quiz-session")
	if err != nil {
		logger.Warn("Failed to retrieve session", zap.Error(err))
		http.Error(w, "Invalid session", http.StatusUnauthorized)
		return
	}

	username, ok := session.Values["username"].(string)
	sessionToken, tokenOk := session.Values["session_token"].(string)
	if !ok || !tokenOk || username == "" || sessionToken == "" {
		logger.Warn("Invalid session: missing username or token")
		http.Error(w, "Invalid session", http.StatusUnauthorized)
		return
	}

	storedUsername, exists := utils.SessionDB[sessionToken]
	if !exists || storedUsername != username {
		logger.Warn("Session token not found or mismatched",
			zap.String("session_token", sessionToken),
			zap.String("stored_username", storedUsername))
		http.Error(w, "Invalid session", http.StatusUnauthorized)
		return
	}

	if err := h.QuizService.StartQuiz(username); err != nil {
		logger.Error("Failed to start quiz", zap.String("username", username), zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	logger.Info("Quiz started successfully", zap.String("username", username))
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status":        "quiz started",
		"next_endpoint": "/quiz/next",
	})
}

// NextQuestion retrieves the next question for the user
// @Summary Get the next quiz question
// @Description Provides the next question for the ongoing quiz session
// @Tags Quiz
// @Produce json
// @Success 200 {object} models.Question "Next question"
// @Failure 410 {object} map[string]string "Quiz complete"
// @Failure 500 {string} string "Internal server error"
// @Router /quiz/next [get]
func (h *QuizHandler) NextQuestion(w http.ResponseWriter, r *http.Request) {
	logger := utils.GetLogger().Sugar()
	session, _ := utils.SessionStore.Get(r, "quiz-session")
	username, _ := session.Values["username"].(string)

	question, err := h.QuizService.GetNextQuestion(username)
	if err != nil {
		if err.Error() == "quiz complete" {
			logger.Info("Quiz complete", zap.String("username", username))
			w.WriteHeader(http.StatusGone)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"status":           "quiz complete",
				"results_endpoint": "/quiz/results",
			})
			return
		}
		logger.Error("Failed to retrieve next question", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	logger.Info("Next question retrieved successfully", zap.String("username", username))
	if err := json.NewEncoder(w).Encode(question); err != nil {
		logger.Warn("Failed to encode question response", zap.Error(err))
	}
}

// SubmitAnswer submits an answer for the current question
// @Summary Submit an answer
// @Description Validates and submits the user's answer to the current question
// @Tags Quiz
// @Accept json
// @Produce json
// @Param payload body models.AnswerPayload true "Answer payload"
// @Success 200 {object} map[string]string "Answer feedback"
// @Failure 400 {string} string "Invalid input"
// @Failure 500 {string} string "Internal server error"
// @Router /quiz/answer [post]
func (h *QuizHandler) SubmitAnswer(w http.ResponseWriter, r *http.Request) {
	logger := utils.GetLogger().Sugar()
	var payload models.AnswerPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		logger.Warn("Invalid input for answer submission", zap.Error(err))
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	session, _ := utils.SessionStore.Get(r, "quiz-session")
	username, _ := session.Values["username"].(string)

	allQuestions, err := h.QuizService.GetQuestions()
	if err != nil {
		logger.Error("Failed to retrieve questions during answer submission", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if err := utils.ValidateAnswerPayload(payload.QuestionIndex, payload.Answer, allQuestions); err != nil {
		logger.Warn("Validation failed for answer submission", zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	correct, err := h.QuizService.SubmitAnswer(username, payload.QuestionIndex, payload.Answer)
	if err != nil {
		logger.Error("Failed to submit answer", zap.String("username", username), zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	message := "Wrong answer"
	if correct {
		message = "Correct answer"
		logger.Info("Correct answer submitted", zap.String("username", username), zap.Int("questionIndex", payload.QuestionIndex))
	} else {
		logger.Info("Wrong answer submitted", zap.String("username", username), zap.Int("questionIndex", payload.QuestionIndex))
	}

	if err := json.NewEncoder(w).Encode(map[string]string{"message": message}); err != nil {
		logger.Warn("Failed to encode answer response", zap.Error(err))
	}
}

// GetResults retrieves the quiz results for the user
// @Summary Get quiz results
// @Description Fetches the quiz results for the logged-in user
// @Tags Quiz
// @Produce json
// @Success 200 {object} map[string]int "Quiz score"
// @Failure 500 {string} string "Internal server error"
// @Router /quiz/results [get]
func (h *QuizHandler) GetResults(w http.ResponseWriter, r *http.Request) {
	logger := utils.GetLogger().Sugar()
	session, _ := utils.SessionStore.Get(r, "quiz-session")
	username, _ := session.Values["username"].(string)

	score, err := h.QuizService.GetResults(username)
	if err != nil {
		logger.Error("Failed to retrieve quiz results", zap.String("username", username), zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	logger.Info("Quiz results retrieved successfully", zap.String("username", username), zap.Int("score", score))
	if err := json.NewEncoder(w).Encode(map[string]int{"score": score}); err != nil {
		logger.Warn("Failed to encode results response", zap.Error(err))
	}
}

// GetStats retrieves the statistics for the user
// @Summary Get user stats
// @Description Provides the statistics of the user, including their quiz performance
// @Tags Quiz
// @Produce json
// @Success 200 {object} map[string]interface{} "User statistics"
// @Failure 401 {string} string "Invalid session"
// @Failure 404 {string} string "No stats available"
// @Failure 500 {string} string "Internal server error"
// @Router /quiz/stats [get]
func (h *QuizHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	logger := utils.GetLogger().Sugar()
	session, err := utils.SessionStore.Get(r, "quiz-session")
	if err != nil {
		logger.Error("Failed to retrieve session", zap.Error(err))
		http.Error(w, `{"message":"Invalid session"}`, http.StatusUnauthorized)
		return
	}

	username, ok := session.Values["username"].(string)
	if !ok || username == "" {
		logger.Warn("No username found in session")
		http.Error(w, `{"message":"Invalid session"}`, http.StatusUnauthorized)
		return
	}

	logger.Info("Processing stats request", zap.String("username", username))

	users, statsMessage, err := h.QuizService.GetStats(username)
	if err != nil {
		if errors.Is(err, services.ErrNoStatsForUser) {
			logger.Warn("No stats available for user", zap.String("username", username))
			http.Error(w, `{"message":"No stats available"}`, http.StatusNotFound)
			return
		}
		logger.Error("Failed to retrieve statistics", zap.String("username", username), zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"users":   users,
		"message": statsMessage,
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Warn("Failed to encode statistics response", zap.Error(err))
	}
}
