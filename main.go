package main

import (
	"flag"
	"log"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/Dzsodie/quiz_app/cmd"
	"github.com/Dzsodie/quiz_app/config"
	"github.com/Dzsodie/quiz_app/internal/database"
	"github.com/Dzsodie/quiz_app/internal/handlers"
	"github.com/Dzsodie/quiz_app/internal/health"
	"github.com/Dzsodie/quiz_app/internal/middleware"
	"github.com/Dzsodie/quiz_app/internal/services"
	"github.com/Dzsodie/quiz_app/internal/utils"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"go.uber.org/zap"

	_ "github.com/Dzsodie/quiz_app/docs"
	httpSwagger "github.com/swaggo/http-swagger"
)

type QuizApp struct {
	DB *database.MemoryDB
}

func (app *QuizApp) Run() {
	cfg := config.LoadConfig()

	cliMode := flag.Bool("cli", false, "Run the application in CLI mode")
	flag.Parse()

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
		go setupRESTAPIServer(cfg, sugar, app.DB)
		time.Sleep(2 * time.Second)

		req, err := http.NewRequest("POST", "/quiz/start", nil)
		if err != nil {
			sugar.Fatalf("Failed to create request: %v", err)
		}
		rr := httptest.NewRecorder()
		quizHandler := handlers.NewQuizHandler(&services.QuizService{DB: app.DB})
		quizHandler.StartQuiz(rr, req)
		return
	}

	sugar.Infof("Application started in %s mode", cfg.Environment)

	setupRESTAPIServer(cfg, sugar, app.DB)
	cmd.Execute()
}

func newQuizApp(db *database.MemoryDB) *QuizApp {
	return &QuizApp{DB: db}
}

func main() {
	cfg := config.LoadConfig()
	utils.InitializeSessionStore(cfg)

	memoryDB := database.NewMemoryDB()
	initializeMockUsers(memoryDB)

	app := newQuizApp(memoryDB)
	app.Run()
}

func initializeMockUsers(memoryDB *database.MemoryDB) {
	mockUsers := []database.User{
		{
			UserID:     uuid.New().String(),
			Username:   "testuser1",
			Password:   "password-1",
			Progress:   []int{},
			Score:      10,
			QuizTaken:  1,
			Percentage: 50,
		},
		{
			UserID:     uuid.New().String(),
			Username:   "testuser2",
			Password:   "password@25",
			Progress:   []int{},
			Score:      20,
			QuizTaken:  2,
			Percentage: 50,
		},
		{
			UserID:     uuid.New().String(),
			Username:   "testuser3",
			Password:   "password!34",
			Progress:   []int{},
			Score:      15,
			QuizTaken:  2,
			Percentage: 0,
		},
	}

	for _, user := range mockUsers {
		memoryDB.AddUser(user)
	}
}

func setupRESTAPIServer(cfg config.Config, sugar *zap.SugaredLogger, db *database.MemoryDB) {
	quizService := &services.QuizService{DB: db}
	authService := &services.AuthService{DB: db}

	sugar.Info("Loading questions from CSV...")
	questions, err := utils.ReadCSV(cfg.QuestionsFilePath)
	if err != nil {
		sugar.Fatalf("Failed to load questions: %v", err)
	}
	sugar.Infof("Successfully loaded %d questions", len(questions))
	quizService.LoadQuestions(questions)

	r := setupRoutes(quizService, authService)

	sugar.Infof("Server is running on port %s...", cfg.ServerPort)
	if err := http.ListenAndServe(cfg.ServerPort, r); err != nil {
		sugar.Fatalf("Server failed: %v", err)
	}
}

func setupRoutes(quizService *services.QuizService, authService *services.AuthService) *mux.Router {
	r := mux.NewRouter()

	quizHandler := handlers.NewQuizHandler(quizService)
	authHandler := handlers.NewAuthHandler(authService)

	r.HandleFunc("/register", authHandler.RegisterUser).Methods("POST")
	r.HandleFunc("/login", authHandler.LoginUser).Methods("POST")
	r.HandleFunc("/questions", quizHandler.GetQuestions).Methods("GET")

	api := r.PathPrefix("/quiz").Subrouter()
	api.Use(middleware.AuthMiddleware)
	api.HandleFunc("/start", quizHandler.StartQuiz).Methods("POST")
	api.HandleFunc("/next", quizHandler.NextQuestion).Methods("GET")
	api.HandleFunc("/submit", quizHandler.SubmitAnswer).Methods("POST")
	api.HandleFunc("/results", quizHandler.GetResults).Methods("GET")
	api.HandleFunc("/stats", quizHandler.GetStats).Methods("GET")

	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	inMemoryDB := make(map[string]string)
	healthChecker := health.NewHealthCheck(inMemoryDB)
	r.HandleFunc("/health", healthChecker.HealthCheckHandler).Methods("GET")

	return r
}
