package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Dzsodie/quiz_app/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

type MockStatsService struct {
	mock.Mock
}

func (m *MockStatsService) GetStats(username string) (string, error) {
	args := m.Called(username)
	return args.String(0), args.Error(1)
}

func TestStatsHandler_GetStats(t *testing.T) {
	mockService := new(MockStatsService)
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	statsHandler := NewStatsHandler(mockService, logger)

	SessionStore = createTestSessionStore()

	tests := []struct {
		name           string
		username       string
		mockReturnMsg  string
		mockReturnErr  error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Valid stats retrieval",
			username:       "testuser",
			mockReturnMsg:  "Your score is 80 and that is 90% better than other users' scores.",
			mockReturnErr:  nil,
			expectedStatus: http.StatusOK,
			expectedBody:   `{"message":"Your score is 80 and that is 90% better than other users' scores."}`,
		},
		{
			name:           "No stats available for user",
			username:       "unknownuser",
			mockReturnMsg:  "",
			mockReturnErr:  services.ErrNoStatsForUser,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"message":"No stats available for user"}`,
		},
		{
			name:           "Internal server error",
			username:       "erroruser",
			mockReturnMsg:  "",
			mockReturnErr:  errors.New("Internal server error"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"message":"Internal server error"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService.On("GetStats", tt.username).Return(tt.mockReturnMsg, tt.mockReturnErr).Once()

			req, err := http.NewRequest(http.MethodGet, "/quiz/stats", nil)
			assert.NoError(t, err)

			rr := httptest.NewRecorder()

			session, _ := SessionStore.Get(req, "quiz-session")
			session.Values["username"] = tt.username
			session.Save(req, rr)

			statsHandler.GetStats(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.JSONEq(t, tt.expectedBody, rr.Body.String())
			mockService.AssertExpectations(t)
		})
	}
}
