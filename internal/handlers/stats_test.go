package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Dzsodie/quiz_app/internal/models"
	"github.com/Dzsodie/quiz_app/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockStatsService struct {
	mock.Mock
}

// Updated to match the StatsService.GetStats method signature
func (m *MockStatsService) GetStats(username string) ([]models.User, string, error) {
	args := m.Called(username)
	users, _ := args.Get(0).([]models.User) // Convert the first return value to []models.User
	return users, args.String(1), args.Error(2)
}

func TestStatsHandlerGetStats(t *testing.T) {
	mockService := new(MockStatsService)
	statsHandler := NewStatsHandler(mockService)

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
			mockService.On("GetStats", tt.username).Return(nil, tt.mockReturnMsg, tt.mockReturnErr).Once()

			req, err := http.NewRequest(http.MethodGet, "/quiz/stats", nil)
			assert.NoError(t, err)

			rr := httptest.NewRecorder()
			session, _ := SessionStore.Get(req, "quiz-session")
			session.Values["username"] = tt.username
			err = session.Save(req, rr)
			assert.NoError(t, err, "Failed to save session")

			statsHandler.GetStats(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.JSONEq(t, tt.expectedBody, rr.Body.String())
			mockService.AssertExpectations(t)
		})
	}
}
