/*
Copyright © 2025 Zsuzsa Makara <dzsodie@gmail.com>
*/
package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strconv"
	"sync"
	"time"

	_ "github.com/Dzsodie/quiz_app/docs"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	httpSwagger "github.com/swaggo/http-swagger"
)

type Question struct {
	Question string   `json:"question"`
	Options  []string `json:"options"`
	Answer   int      `json:"answer"`
}

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AnswerPayload struct {
	QuestionIndex int `json:"question_index"`
	Answer        int `json:"answer"`
}

var (
	questions    []Question
	users        = make(map[string]User)                          // Registered users
	sessionStore = sessions.NewCookieStore([]byte("quiz-secret")) // Session store
	userScores   = make(map[string]int)                           // User scores
	quizTimers   = make(map[string]*time.Timer)                   // Timers for each quiz
	userProgress = make(map[string]int)                           // Tracks progress per user
	mu           sync.Mutex                                       // Mutex for concurrent safety
)

func readCSV(filename string) ([]Question, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV file: %w", err)
	}

	var questions []Question
	for i, record := range records {
		if i == 0 {
			// Skip header
			continue
		}
		if len(record) < 5 {
			return nil, fmt.Errorf("invalid record: %v", record)
		}
		answer, err := strconv.Atoi(record[4])
		if err != nil {
			return nil, fmt.Errorf("invalid answer format in record %v: %w", record, err)
		}
		questions = append(questions, Question{
			Question: record[0],
			Options:  record[1:4],
			Answer:   answer,
		})
	}
	return questions, nil
}

// @Summary Register a new user
// @Description Register a user with a username and password
// @Tags User
// @Accept json
// @Produce json
// @Param user body User true "User details"
// @Success 201 {object} map[string]string "message"
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 409 {object} map[string]string "User already exists"
// @Router /register [post]
func registerUser(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	if _, exists := users[user.Username]; exists {
		http.Error(w, "User already exists", http.StatusConflict)
		return
	}

	users[user.Username] = user
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "User registered successfully"})
}

// @Summary Login a user
// @Description Login with a username and password
// @Tags User
// @Accept json
// @Produce json
// @Param user body User true "User details"
// @Success 200 {object} map[string]string "message"
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 401 {object} map[string]string "Invalid credentials"
// @Router /login [post]
func loginUser(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	storedUser, exists := users[user.Username]
	if !exists || storedUser.Password != user.Password {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	session, _ := sessionStore.Get(r, "quiz-session")
	session.Values["username"] = user.Username
	session.Save(r, w)

	json.NewEncoder(w).Encode(map[string]string{"message": "Login successful"})
}

// Middleware to validate session
func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _ := sessionStore.Get(r, "quiz-session")
		username, ok := session.Values["username"].(string)
		if !ok || username == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// @Summary Get all questions
// @Description Retrieve the list of questions
// @Tags Quiz
// @Produce json
// @Success 200 {array} Question
// @Router /questions [get]
func getQuestions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(questions)
}

// @Summary Start the quiz
// @Description Initialize a new quiz session
// @Tags Quiz
// @Success 200 {object} map[string]string "message"
// @Router /quiz/start [post]
func startQuiz(w http.ResponseWriter, r *http.Request) {
	session, _ := sessionStore.Get(r, "quiz-session")
	username, _ := session.Values["username"].(string)

	mu.Lock()
	userScores[username] = 0
	userProgress[username] = 0

	if quizTimers[username] != nil {
		quizTimers[username].Stop()
	}

	quizTimers[username] = time.AfterFunc(10*time.Minute, func() {
		mu.Lock()
		delete(userProgress, username)
		mu.Unlock()
	})
	mu.Unlock()

	json.NewEncoder(w).Encode(map[string]string{"message": "Quiz started", "next_endpoint": "/quiz/next"})
}

func nextQuestionHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := sessionStore.Get(r, "quiz-session")
	username, _ := session.Values["username"].(string)

	mu.Lock()
	progress := userProgress[username]
	if progress >= len(questions) {
		mu.Unlock()
		json.NewEncoder(w).Encode(map[string]string{"message": "Quiz complete", "results_endpoint": "/quiz/results"})
		return
	}
	question := questions[progress]
	userProgress[username]++
	mu.Unlock()

	json.NewEncoder(w).Encode(question)
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := sessionStore.Get(r, "quiz-session")
	username, _ := session.Values["username"].(string)

	mu.Lock()
	userScore := userScores[username]
	allScores := make([]int, 0, len(userScores))
	for _, score := range userScores {
		allScores = append(allScores, score)
	}
	mu.Unlock()

	sort.Ints(allScores)
	betterScores := 0
	for _, score := range allScores {
		if userScore > score {
			betterScores++
		}
	}

	totalUsers := len(allScores)
	percentage := (float64(betterScores) / float64(totalUsers)) * 100
	json.NewEncoder(w).Encode(map[string]string{
		"message": fmt.Sprintf("Your score is %d and that is %.2f%% better than other users' scores.", userScore, percentage),
	})
}

// Update the Swagger comment
// @Summary Submit an answer
// @Description Submit an answer to a question
// @Tags Quiz
// @Accept json
// @Produce json
// @Param payload body AnswerPayload true "Answer payload"
// @Success 200 {object} map[string]string "message"
// @Failure 400 {object} map[string]string "Invalid input"
// @Router /quiz/submit [post]
func submitAnswer(w http.ResponseWriter, r *http.Request) {
	var payload AnswerPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	session, _ := sessionStore.Get(r, "quiz-session")
	username, _ := session.Values["username"].(string)

	if payload.QuestionIndex < 0 || payload.QuestionIndex >= len(questions) {
		http.Error(w, "Invalid question index", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	correctAnswer := questions[payload.QuestionIndex].Answer
	if payload.Answer == correctAnswer {
		userScores[username]++
		json.NewEncoder(w).Encode(map[string]string{"message": "Correct answer"})
	} else {
		json.NewEncoder(w).Encode(map[string]string{"message": "Wrong answer"})
	}
}

// @Summary Get quiz results
// @Description Retrieve the results of the quiz
// @Tags Quiz
// @Produce json
// @Success 200 {object} map[string]int "score"
// @Router /quiz/results [get]
func getResults(w http.ResponseWriter, r *http.Request) {
	session, _ := sessionStore.Get(r, "quiz-session")
	username, _ := session.Values["username"].(string)

	mu.Lock()
	score := userScores[username]
	mu.Unlock()

	json.NewEncoder(w).Encode(map[string]int{"score": score})
}

// @title Quiz App API
// @version 1.0
// @description This is a quiz app API.
// @termsOfService http://swagger.io/terms/

// @contact.name Zsuzsa Makara
// @contact.url http://www.zsuzsa-makara.com
// @contact.email dzsodie@gmail.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /
func main() {
	var err error
	questions, err = readCSV("questions.csv")
	if err != nil {
		fmt.Println("Error reading CSV:", err)
		os.Exit(1)
	}

	r := mux.NewRouter()

	// Public routes
	r.HandleFunc("/register", registerUser).Methods("POST")
	r.HandleFunc("/login", loginUser).Methods("POST")
	r.HandleFunc("/questions", getQuestions).Methods("GET")

	// Protected routes
	api := r.PathPrefix("/quiz").Subrouter()
	api.Use(authMiddleware)
	api.HandleFunc("/start", startQuiz).Methods("POST")
	api.HandleFunc("/next", nextQuestionHandler).Methods("GET")
	api.HandleFunc("/submit", submitAnswer).Methods("POST")
	api.HandleFunc("/results", getResults).Methods("GET")
	api.HandleFunc("/stats", statsHandler).Methods("GET")

	// Swagger routes
	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	fmt.Println("Server is running on port 8080...")
	http.ListenAndServe(":8080", r)
}
