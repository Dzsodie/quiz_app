package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/sessions"
	"github.com/stretchr/testify/assert"
)

// Helper function to create a test session store
func createTestSessionStore() *sessions.CookieStore {
	store := sessions.NewCookieStore([]byte("test-secret"))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   3600,
		HttpOnly: true,
	}
	return store
}

func TestAuthMiddleware(t *testing.T) {
	SetSessionStore(createTestSessionStore())

	tests := []struct {
		name           string
		setupSession   func(req *http.Request, rr *httptest.ResponseRecorder)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Valid session with username",
			setupSession: func(req *http.Request, rr *httptest.ResponseRecorder) {
				session, _ := sessionStore.Get(req, "quiz-session")
				session.Values["username"] = "testuser"
				session.Save(req, rr)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "OK",
		},
		{
			name: "Session missing username",
			setupSession: func(req *http.Request, rr *httptest.ResponseRecorder) {
				session, _ := sessionStore.Get(req, "quiz-session")
				session.Save(req, rr)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Unauthorized\n",
		},
		{
			name: "No session available",
			setupSession: func(req *http.Request, rr *httptest.ResponseRecorder) {
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Unauthorized\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("OK"))
			})

			handler := AuthMiddleware(dummyHandler)

			req, err := http.NewRequest(http.MethodGet, "/protected", nil)
			assert.NoError(t, err)

			rr := httptest.NewRecorder()

			tt.setupSession(req, rr)

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Equal(t, tt.expectedBody, rr.Body.String())
		})
	}
}
