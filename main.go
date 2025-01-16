package main

import (
	"log"
	"net/http"
	"os"

	"github.com/Dzsodie/quiz_app/internal/handlers"
	"github.com/Dzsodie/quiz_app/internal/health"
	"github.com/Dzsodie/quiz_app/internal/middleware"
	"github.com/Dzsodie/quiz_app/internal/services"
	"github.com/Dzsodie/quiz_app/internal/utils"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	_ "github.com/Dzsodie/quiz_app/docs"
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title Quiz App API
// @version 1.0
// @description This is a quiz app API.
// @termsOfService http://swagger.io/terms/

// @contact.name Zsuzsa Makara
// @contact.url https://dzsodie.github.io/
// @contact.email dzsodie@gmail.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /
func main() {
	env := os.Getenv("ENV")
	if env == "" {
		env = "development"
	}
	logFilePath := os.Getenv("LOG_FILE_PATH")
	if logFilePath == "" {
		logFilePath = "logs/app.log"
	}

	logger, err := utils.InitializeLogger(env, logFilePath)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()
	sugar := logger.Sugar()

	sugar.Infof("Application started in %s mode", env)

	utils.SetLogger(logger)

	quizService := &services.QuizService{Logger: logger}
	authService := &services.AuthService{Logger: logger}
	statsService := &services.StatsService{Logger: logger}

	sugar.Info("Loading questions from CSV...")
	questions, err := utils.ReadCSV("questions.csv")
	if err != nil {
		sugar.Fatalf("Failed to load questions: %v", err)
	}
	sugar.Infof("Successfully loaded %d questions", len(questions))

	quizService.LoadQuestions(questions)

	quizHandler := handlers.NewQuizHandler(quizService, sugar)
	authHandler := handlers.NewAuthHandler(authService, sugar)
	statsHandler := handlers.NewStatsHandler(statsService, logger)

	r := mux.NewRouter()
	handlers.SessionStore = sessions.NewCookieStore([]byte("quiz-secret"))
	middleware.SetSessionStore(handlers.SessionStore)
	middleware.SetLogger(logger)

	r.HandleFunc("/register", authHandler.RegisterUser).Methods("POST")
	r.HandleFunc("/login", authHandler.LoginUser).Methods("POST")
	r.HandleFunc("/questions", quizHandler.GetQuestions).Methods("GET")

	api := r.PathPrefix("/quiz").Subrouter()
	api.Use(middleware.AuthMiddleware)
	api.HandleFunc("/start", quizHandler.StartQuiz).Methods("POST")
	api.HandleFunc("/next", quizHandler.NextQuestion).Methods("GET")
	api.HandleFunc("/submit", quizHandler.SubmitAnswer).Methods("POST")
	api.HandleFunc("/results", quizHandler.GetResults).Methods("GET")
	api.HandleFunc("/stats", statsHandler.GetStats).Methods("GET")

	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	inMemoryDB := make(map[string]string)
	healthChecker := health.NewHealthCheck(sugar, handlers.SessionStore, inMemoryDB)
	r.HandleFunc("/health", healthChecker.HealthCheckHandler).Methods("GET")

	sugar.Info("Server is running on port 8080...")
	sugar.Fatal(http.ListenAndServe(":8080", r))
}
