package handlers

import (
	"bytes"

	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Dzsodie/quiz_app/internal/models"
	"github.com/gorilla/sessions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockQuizService is a mock implementation of the IQuizService interface.
type MockQuizService struct {
	mock.Mock
}

func (m *MockQuizService) GetQuestions() []models.Question {
	args := m.Called()
	return args.Get(0).([]models.Question)
}

func (m *MockQuizService) LoadQuestions(qs []models.Question) {
	m.Called(qs)
}

func (m *MockQuizService) StartQuiz(username string) {
	m.Called(username)
}

func (m *MockQuizService) GetNextQuestion(username string) (*models.Question, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Question), args.Error(1)
}

func (m *MockQuizService) SubmitAnswer(username string, questionIndex, answer int) (bool, error) {
	args := m.Called(username, questionIndex, answer)
	return args.Bool(0), args.Error(1)
}

func (m *MockQuizService) GetResults(username string) (int, error) {
	args := m.Called(username)
	return args.Int(0), args.Error(1)
}

func TestQuizHandler_GetQuestions(t *testing.T) {
	mockService := new(MockQuizService)
	quizHandler := NewQuizHandler(mockService)

	questions := []models.Question{
		{Question: "What is 2+2?", Options: []string{"3", "4", "5"}, Answer: 1},
	}

	mockService.On("GetQuestions").Return(questions)

	req, err := http.NewRequest(http.MethodGet, "/questions", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	quizHandler.GetQuestions(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.JSONEq(t, `[{"Question":"What is 2+2?","Options":["3","4","5"],"Answer":1}]`, rr.Body.String())
	mockService.AssertExpectations(t)
}

func TestQuizHandler_StartQuiz(t *testing.T) {
	mockService := new(MockQuizService)
	quizHandler := NewQuizHandler(mockService)

	SessionStore = createTestSessionStore() // Helper to create a session store
	req, err := http.NewRequest(http.MethodPost, "/quiz/start", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	session, _ := SessionStore.Get(req, "quiz-session")
	session.Values["username"] = "testuser"
	session.Save(req, rr)

	mockService.On("StartQuiz", "testuser").Return()

	quizHandler.StartQuiz(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.JSONEq(t, `{"message":"Quiz started","next_endpoint":"/quiz/next"}`, rr.Body.String())
	mockService.AssertExpectations(t)
}

func TestQuizHandler_NextQuestion(t *testing.T) {
	mockService := new(MockQuizService)
	quizHandler := NewQuizHandler(mockService)

	SessionStore = createTestSessionStore()
	req, err := http.NewRequest(http.MethodGet, "/quiz/next", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	session, _ := SessionStore.Get(req, "quiz-session")
	session.Values["username"] = "testuser"
	session.Save(req, rr)

	question := &models.Question{Question: "What is 2+2?", Options: []string{"3", "4", "5"}, Answer: 1}
	mockService.On("GetNextQuestion", "testuser").Return(question, nil)

	quizHandler.NextQuestion(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.JSONEq(t, `{"Question":"What is 2+2?","Options":["3","4","5"],"Answer":1}`, rr.Body.String())
	mockService.AssertExpectations(t)
}

func TestQuizHandler_SubmitAnswer(t *testing.T) {
	mockService := new(MockQuizService)
	quizHandler := NewQuizHandler(mockService)

	SessionStore = createTestSessionStore()
	reqBody := `{"QuestionIndex":1,"Answer":2}`
	req, err := http.NewRequest(http.MethodPost, "/quiz/submit", bytes.NewBufferString(reqBody))
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	session, _ := SessionStore.Get(req, "quiz-session")
	session.Values["username"] = "testuser"
	session.Save(req, rr)

	mockService.On("SubmitAnswer", "testuser", 1, 2).Return(true, nil)

	quizHandler.SubmitAnswer(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.JSONEq(t, `{"message":"Correct answer"}`, rr.Body.String())
	mockService.AssertExpectations(t)
}

func TestQuizHandler_GetResults(t *testing.T) {
	mockService := new(MockQuizService)
	quizHandler := NewQuizHandler(mockService)

	SessionStore = createTestSessionStore()
	req, err := http.NewRequest(http.MethodGet, "/quiz/results", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	session, _ := SessionStore.Get(req, "quiz-session")
	session.Values["username"] = "testuser"
	session.Save(req, rr)

	mockService.On("GetResults", "testuser").Return(10, nil)

	quizHandler.GetResults(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.JSONEq(t, `{"score":10}`, rr.Body.String())
	mockService.AssertExpectations(t)
}

// Helper function to create a test session store
func createTestSessionStore() *sessions.CookieStore {
	store := sessions.NewCookieStore([]byte("test-secret"))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   3600, // 1 hour
		HttpOnly: true,
	}
	return store
}
