package main

import (
	"flag"
	"log"
	"net/http"
	"os"

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
		go setupRESTAPIServer(cfg, sugar, true, app.DB)
		time.Sleep(2 * time.Second)
		handler := handlers.NewStartQuizCliHandler(
			services.NewStartQuizCLIService(cfg.APIBaseURL, &http.Client{}, app.DB),
		)
		handler.StartQuizCLI(cfg.APIBaseURL, app.DB)
		return
	}

	sugar.Infof("Application started in %s mode", cfg.Environment)

	setupRESTAPIServer(cfg, sugar, false, app.DB)
	cmd.Execute()
}

func newQuizApp(db *database.MemoryDB) *QuizApp {
	return &QuizApp{DB: db}
}

func main() {
	utils.InitializeSessionStore()

	memoryDB := database.NewMemoryDB()
	memoryDB.AddUser(database.User{
		UserID:   uuid.New().String(),
		Username: "testuser1",
		Password: "password-1",
		Progress: []int{},
		Score:    10,
	})

	memoryDB.AddUser(database.User{
		UserID:   uuid.New().String(),
		Username: "testuser2",
		Password: "password@25",
		Progress: []int{},
		Score:    20,
	})

	memoryDB.AddUser(database.User{
		UserID:   uuid.New().String(),
		Username: "testuser3",
		Password: "password!34",
		Progress: []int{},
		Score:    15,
	})

	app := newQuizApp(memoryDB)

	app.Run()
}

func setupRESTAPIServer(cfg config.Config, sugar *zap.SugaredLogger, suppressLogs bool, db *database.MemoryDB) {

	quizService := &services.QuizService{DB: db}
	authService := &services.AuthService{DB: db}

	sugar.Info("Loading questions from CSV...")
	questions, err := utils.ReadCSV(cfg.QuestionsFilePath)
	if err != nil {
		sugar.Fatalf("Failed to load questions: %v", err)
	}
	sugar.Infof("Successfully loaded %d questions", len(questions))
	quizService.LoadQuestions(questions)

	quizHandler := handlers.NewQuizHandler(quizService)
	authHandler := handlers.NewAuthHandler(authService)

	r := mux.NewRouter()
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
