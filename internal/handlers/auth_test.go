package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Dzsodie/quiz_app/internal/models"
	"github.com/gorilla/sessions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) GetUserID(username string) (string, error) {
	args := m.Called(username)
	return args.String(0), args.Error(1)
}

func (m *MockAuthService) RegisterUser(username, password string) error {
	args := m.Called(username, password)
	return args.Error(0)
}

func (m *MockAuthService) AuthenticateUser(username, password string) error {
	args := m.Called(username, password)
	return args.Error(0)
}

func TestRegisterUserHandler(t *testing.T) {
	mockService := new(MockAuthService)
	authHandler := NewAuthHandler(mockService)

	tests := []struct {
		name            string
		input           models.User
		mockReturnErr   error
		mockUserID      string
		mockUserIDError error
		expectedStatus  int
		expectedBody    string
	}{
		{
			name:            "Valid user registration",
			input:           models.User{Username: "testuser", Password: "password"},
			mockReturnErr:   nil,
			mockUserID:      "12345",
			mockUserIDError: nil,
			expectedStatus:  http.StatusCreated,
			expectedBody:    `{"userID":"12345","message":"User registered successfully"}`,
		},
		{
			name:           "Invalid input - empty body",
			input:          models.User{},
			mockReturnErr:  nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"message":"Invalid input"}`,
		},
		{
			name:           "User already exists",
			input:          models.User{Username: "testuser", Password: "password"},
			mockReturnErr:  errors.New("user already exists"),
			expectedStatus: http.StatusConflict,
			expectedBody:   `{"message":"user already exists"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up mocks
			if tt.input.Username != "" || tt.input.Password != "" {
				mockService.On("RegisterUser", tt.input.Username, tt.input.Password).Return(tt.mockReturnErr).Once()
			}
			if tt.mockReturnErr == nil && tt.mockUserID != "" {
				mockService.On("GetUserID", tt.input.Username).Return(tt.mockUserID, tt.mockUserIDError).Once()
			}

			body, _ := json.Marshal(tt.input)
			req, _ := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
			rr := httptest.NewRecorder()

			// Call the handler
			authHandler.RegisterUser(rr, req)

			// Validate response
			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.JSONEq(t, tt.expectedBody, rr.Body.String())
			mockService.AssertExpectations(t)
		})
	}
}

func TestLoginUserHandler(t *testing.T) {
	mockService := new(MockAuthService)
	authHandler := NewAuthHandler(mockService)

	SessionStore = sessions.NewCookieStore([]byte("test-secret"))

	tests := []struct {
		name           string
		input          models.User
		mockReturnErr  error
		expectedStatus int
		expectedBody   map[string]interface{}
		validateToken  bool // Whether to validate the presence of session_token
	}{
		{
			name:           "Valid login",
			input:          models.User{Username: "testuser", Password: "password"},
			mockReturnErr:  nil,
			expectedStatus: http.StatusOK,
			expectedBody:   map[string]interface{}{"message": "Login successful"},
			validateToken:  true,
		},
		{
			name:           "Invalid input - empty body",
			input:          models.User{},
			mockReturnErr:  nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   map[string]interface{}{"message": "Invalid input"},
			validateToken:  false,
		},
		{
			name:           "Invalid credentials",
			input:          models.User{Username: "testuser", Password: "wrongpassword"},
			mockReturnErr:  errors.New("invalid username or password"),
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   map[string]interface{}{"message": "invalid username or password"},
			validateToken:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up mock behavior for AuthenticateUser
			if tt.input.Username != "" || tt.input.Password != "" {
				mockService.On("AuthenticateUser", tt.input.Username, tt.input.Password).Return(tt.mockReturnErr).Once()
			}

			// Prepare the HTTP request
			body, _ := json.Marshal(tt.input)
			req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
			rr := httptest.NewRecorder()

			// Call the handler
			authHandler.LoginUser(rr, req)

			// Validate the HTTP status code
			assert.Equal(t, tt.expectedStatus, rr.Code)

			// Parse and validate the JSON response
			var responseBody map[string]interface{}
			err := json.Unmarshal(rr.Body.Bytes(), &responseBody)
			assert.NoError(t, err)

			// Check for expected message
			assert.Equal(t, tt.expectedBody["message"], responseBody["message"])

			// Validate session_token if required
			if tt.validateToken {
				assert.NotEmpty(t, responseBody["session_token"], "Expected session_token to be present")
			}

			mockService.AssertExpectations(t)
		})
	}
}
