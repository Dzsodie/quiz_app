package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"

	"go.uber.org/zap"

	"github.com/Dzsodie/quiz_app/internal/database"
	"github.com/Dzsodie/quiz_app/internal/utils"
)

type Question struct {
	QuestionID int      `json:"question_id"`
	Question   string   `json:"question"`
	Options    []string `json:"options"`
}

type StartQuizCLIService struct {
	ApiBaseURL string
	HttpClient *http.Client
	DB         *database.MemoryDB
}

func NewStartQuizCLIService(apiBaseURL string, client *http.Client, db *database.MemoryDB) *StartQuizCLIService {
	if client.Jar == nil {
		cookieJar, _ := cookiejar.New(nil)
		client.Jar = cookieJar
	}
	return &StartQuizCLIService{ApiBaseURL: apiBaseURL, HttpClient: &http.Client{}, DB: db}
}

func (s *StartQuizCLIService) RegisterUser(username, password string) (string, error) {
	payload := map[string]interface{}{
		"username": username,
		"password": password,
		"progress": []int{},
		"score":    0,
	}
	body, _ := json.Marshal(payload)

	resp, err := s.HttpClient.Post(s.ApiBaseURL+"/register", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("registration failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusOK {
		var responsePayload map[string]string
		if err := json.NewDecoder(resp.Body).Decode(&responsePayload); err != nil {
			return "", fmt.Errorf("failed to parse response: %v", err)
		}

		userID, ok := responsePayload["userID"]
		if !ok {
			return "", fmt.Errorf("response does not contain userID")
		}

		return userID, nil
	}

	bodyBytes, _ := io.ReadAll(resp.Body)
	return "", fmt.Errorf("registration failed: %s", string(bodyBytes))
}

func (s *StartQuizCLIService) LoginUser(username, password string) (string, error) {
	logger := utils.GetLogger().Sugar()

	payload := map[string]string{"username": username, "password": password}
	body, _ := json.Marshal(payload)

	resp, err := s.HttpClient.Post(s.ApiBaseURL+"/login", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("login failed: %v", err)
	}
	defer resp.Body.Close()

	// Check the response status
	if resp.StatusCode == http.StatusOK {
		var responsePayload map[string]string
		if err := json.NewDecoder(resp.Body).Decode(&responsePayload); err != nil {
			return "", fmt.Errorf("failed to parse response: %v", err)
		}

		sessionToken, ok := responsePayload["session_token"]
		if !ok {
			return "", fmt.Errorf("response does not contain session_token")
		}
		logger.Debug("Session token_login_s", "session_token: ", sessionToken)
		return sessionToken, nil
	}

	// Handle unexpected response status
	bodyBytes, _ := io.ReadAll(resp.Body)
	return "", fmt.Errorf("login failed: %s", string(bodyBytes))
}

func (s *StartQuizCLIService) StartQuiz(sessionToken string) error {
	logger := utils.GetLogger().Sugar()

	// Ensure HttpClient uses a CookieJar
	if s.HttpClient.Jar == nil {
		cookieJar, _ := cookiejar.New(nil)
		s.HttpClient.Jar = cookieJar
	}

	// Explicitly set the cookie in the jar
	cookie := &http.Cookie{
		Name:  "quiz-session",
		Value: sessionToken,
		Path:  "/",
	}
	u, _ := url.Parse(s.ApiBaseURL)
	s.HttpClient.Jar.SetCookies(u, []*http.Cookie{cookie})

	// Debug cookies before making the request
	cookies := s.HttpClient.Jar.Cookies(u)
	logger.Debug("Client cookies before request", zap.Any("cookies", cookies))

	req, _ := http.NewRequest(http.MethodPost, s.ApiBaseURL+"/quiz/start", nil)
	resp, err := s.HttpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to start quiz: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to start quiz: %s", string(body))
	}

	return nil
}

func (s *StartQuizCLIService) GetNextQuestion(sessionToken string) (*Question, bool, error) {
	req, _ := http.NewRequest(http.MethodGet, s.ApiBaseURL+"/quiz/next", nil)
	//req.AddCookie(&http.Cookie{Name: "quiz-session", Value: sessionToken})
	resp, err := s.HttpClient.Do(req)
	if err != nil {
		return nil, false, fmt.Errorf("failed to fetch next question")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var question Question
		body, _ := io.ReadAll(resp.Body)
		_ = json.Unmarshal(body, &question)
		return &question, false, nil
	} else if resp.StatusCode == http.StatusNoContent || resp.StatusCode == http.StatusGone {
		return nil, true, nil
	}

	return nil, false, fmt.Errorf("unexpected response")
}

func (s *StartQuizCLIService) SubmitAnswer(sessionToken string, questionID, answer int) (string, error) {
	payload := map[string]int{"QuestionIndex": questionID, "Answer": answer}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPost, s.ApiBaseURL+"/quiz/submit", bytes.NewBuffer(body))
	//req.AddCookie(&http.Cookie{Name: "quiz-session", Value: sessionToken})
	resp, err := s.HttpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to submit answer")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var result map[string]string
		body, _ := io.ReadAll(resp.Body)
		_ = json.Unmarshal(body, &result)
		return result["message"], nil
	}

	return "", fmt.Errorf("unexpected response")
}

func (s *StartQuizCLIService) FetchResults(sessionToken string) (int, error) {
	req, _ := http.NewRequest(http.MethodGet, s.ApiBaseURL+"/quiz/results", nil)
	//req.AddCookie(&http.Cookie{Name: "quiz-session", Value: sessionToken})
	resp, err := s.HttpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch results")
	}
	defer resp.Body.Close()

	var score struct {
		Score int `json:"score"`
	}
	body, _ := io.ReadAll(resp.Body)
	_ = json.Unmarshal(body, &score)
	return score.Score, nil
}

func (s *StartQuizCLIService) FetchStats(sessionToken string) (map[string]string, error) {
	req, _ := http.NewRequest(http.MethodGet, s.ApiBaseURL+"/quiz/stats", nil)
	//req.AddCookie(&http.Cookie{Name: "quiz-session", Value: sessionToken})
	resp, err := s.HttpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch stats")
	}
	defer resp.Body.Close()

	var stats map[string]string
	body, _ := io.ReadAll(resp.Body)
	_ = json.Unmarshal(body, &stats)
	return stats, nil
}
