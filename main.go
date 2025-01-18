package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/Dzsodie/quiz_app/cmd"
	"github.com/Dzsodie/quiz_app/config"
	"github.com/Dzsodie/quiz_app/internal/database"
	"github.com/Dzsodie/quiz_app/internal/handlers"
	"github.com/Dzsodie/quiz_app/internal/health"
	"github.com/Dzsodie/quiz_app/internal/middleware"
	"github.com/Dzsodie/quiz_app/internal/services"
	"github.com/Dzsodie/quiz_app/internal/utils"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"go.uber.org/zap"

	_ "github.com/Dzsodie/quiz_app/docs"
	httpSwagger "github.com/swaggo/http-swagger"
)

type QuizApp struct {
	DB *database.MemoryDB
}

// Run starts the QuizApp
func (app *QuizApp) Run() {
	cliMode := flag.Bool("cli", false, "Run the application in CLI mode")
	flag.Parse()

	cfg := config.LoadConfig()

	logger, err := utils.InitializeLogger(cfg.Environment, cfg.LogFilePath)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer func() {
		if err := logger.Sync(); err != nil {
			log.Printf("Failed to sync logger: %v", err)
		}
	}()
	sugar := logger.Sugar()

	if *cliMode {
		sugar.Info("Starting Quiz in CLI mode...")
		setupRESTAPIServer(cfg, sugar, true)
		handlers.StartQuizCLI(cfg.APIBaseURL)
		return
	}

	sugar.Infof("Application started in %s mode", cfg.Environment)

	// Start REST API server
	setupRESTAPIServer(cfg, sugar, false)
	cmd.Execute()
}

func newQuizApp(db *database.MemoryDB) *QuizApp {
	return &QuizApp{DB: db}
}

func main() {
	// Initialize the in-memory database
	memoryDB := database.NewMemoryDB()
	app := newQuizApp(memoryDB)
	app.Run()
}

// setupRESTAPIServer configures and starts the REST API server
func setupRESTAPIServer(cfg config.Config, sugar *zap.SugaredLogger, suppressLogs bool) {
	// Services and handlers setup
	quizService := &services.QuizService{}
	authService := &services.AuthService{}
	statsService := &services.StatsService{}

	sugar.Info("Loading questions from CSV...")
	questions, err := utils.ReadCSV(cfg.QuestionsFilePath)
	if err != nil {
		sugar.Fatalf("Failed to load questions: %v", err)
	}
	sugar.Infof("Successfully loaded %d questions", len(questions))
	quizService.LoadQuestions(questions)

	quizHandler := handlers.NewQuizHandler(quizService)
	authHandler := handlers.NewAuthHandler(authService)
	statsHandler := handlers.NewStatsHandler(statsService)

	r := mux.NewRouter()
	handlers.SessionStore = sessions.NewCookieStore([]byte(cfg.SessionSecret))
	middleware.SetSessionStore(handlers.SessionStore)

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

	// Health check
	inMemoryDB := make(map[string]string)
	healthChecker := health.NewHealthCheck(handlers.SessionStore, inMemoryDB)
	r.HandleFunc("/health", healthChecker.HealthCheckHandler).Methods("GET")

	// Start server in a separate goroutine
	go func() {
		if suppressLogs {
			log.SetOutput(os.Stdout)
		}
		sugar.Infof("Server is running on port %s...", cfg.ServerPort)
		if err := http.ListenAndServe(cfg.ServerPort, r); err != nil {
			sugar.Fatalf("Server failed: %v", err)
		}
	}()
}
