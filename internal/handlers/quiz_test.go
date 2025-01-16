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

type MockQuizService struct {
	mock.Mock
}

func (m *MockQuizService) GetQuestions() ([]models.Question, error) {
	args := m.Called()
	return args.Get(0).([]models.Question), args.Error(1)
}

func (m *MockQuizService) LoadQuestions(qs []models.Question) {
	m.Called(qs)
}

func (m *MockQuizService) StartQuiz(username string) error {
	args := m.Called(username)
	return args.Error(0)
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

func TestQuizHandlerGetQuestions(t *testing.T) {
	mockService := new(MockQuizService)
	quizHandler := NewQuizHandler(mockService)

	questions := []models.Question{
		{Question: "What is 2+2?", Options: []string{"3", "4", "5"}, Answer: 1},
	}

	mockService.On("GetQuestions").Return(questions, nil)

	req, err := http.NewRequest(http.MethodGet, "/questions", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	quizHandler.GetQuestions(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.JSONEq(t, `[{"question":"What is 2+2?","options":["3","4","5"],"answer":1}]`, rr.Body.String())
	mockService.AssertExpectations(t)
}

func TestQuizHandlerStartQuiz(t *testing.T) {
	mockService := new(MockQuizService)
	quizHandler := NewQuizHandler(mockService)

	SessionStore = createTestSessionStore()

	req, err := http.NewRequest(http.MethodPost, "/quiz/start", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()

	session, _ := SessionStore.Get(req, "quiz-session")
	session.Values["username"] = "testuser"
	err = session.Save(req, rr)
	assert.NoError(t, err, "Failed to save session")

	mockService.On("StartQuiz", "testuser").Return(nil).Once()

	quizHandler.StartQuiz(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.JSONEq(t, `{"status":"quiz started","next_endpoint":"/quiz/next"}`, rr.Body.String())
	mockService.AssertExpectations(t)
}

func TestQuizHandlerNextQuestion(t *testing.T) {
	mockService := new(MockQuizService)
	quizHandler := NewQuizHandler(mockService)

	SessionStore = createTestSessionStore()
	req, err := http.NewRequest(http.MethodGet, "/quiz/next", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	session, _ := SessionStore.Get(req, "quiz-session")
	session.Values["username"] = "testuser"
	if err := session.Save(req, rr); err != nil {
		t.Errorf("Failed to save session: %v", err)
	}

	question := &models.Question{Question: "What is 2+2?", Options: []string{"3", "4", "5"}, Answer: 1}
	mockService.On("GetNextQuestion", "testuser").Return(question, nil)

	quizHandler.NextQuestion(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.JSONEq(t, `{"question":"What is 2+2?","options":["3","4","5"],"answer":1}`, rr.Body.String())
	mockService.AssertExpectations(t)
}

func TestQuizHandlerSubmitAnswer(t *testing.T) {
	mockService := new(MockQuizService)
	quizHandler := NewQuizHandler(mockService)

	SessionStore = createTestSessionStore()

	reqBody := `{"QuestionIndex":0,"Answer":1}`
	req, err := http.NewRequest(http.MethodPost, "/quiz/submit", bytes.NewBufferString(reqBody))
	assert.NoError(t, err)

	rr := httptest.NewRecorder()

	session, _ := SessionStore.Get(req, "quiz-session")
	session.Values["username"] = "testuser"
	err = session.Save(req, rr)
	assert.NoError(t, err, "Failed to save session")

	questions := []models.Question{
		{Question: "What is 2+2?", Options: []string{"3", "4", "5"}, Answer: 1},
	}
	mockService.On("GetQuestions").Return(questions, nil).Once()
	mockService.On("SubmitAnswer", "testuser", 0, 1).Return(true, nil).Once()

	quizHandler.SubmitAnswer(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.JSONEq(t, `{"message":"Correct answer"}`, rr.Body.String())
	mockService.AssertExpectations(t)
}

func TestQuizHandlerGetResults(t *testing.T) {
	mockService := new(MockQuizService)
	quizHandler := NewQuizHandler(mockService)

	SessionStore = createTestSessionStore()
	req, err := http.NewRequest(http.MethodGet, "/quiz/results", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	session, _ := SessionStore.Get(req, "quiz-session")
	session.Values["username"] = "testuser"
	if err := session.Save(req, rr); err != nil {
		t.Errorf("Failed to save session: %v", err)
	}

	mockService.On("GetResults", "testuser").Return(10, nil)

	quizHandler.GetResults(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.JSONEq(t, `{"score":10}`, rr.Body.String())
	mockService.AssertExpectations(t)
}

func createTestSessionStore() *sessions.CookieStore {
	store := sessions.NewCookieStore([]byte("test-secret"))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   3600, // 1 hour
		HttpOnly: true,
	}
	return store
}
