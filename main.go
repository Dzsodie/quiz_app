/*
Copyright © 2025 Zsuzsa Makara <dzsodie@gmail.com>
*/
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Dzsodie/quiz_app/internal/handlers"
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
	quizService := &services.QuizService{}
	// Load questions from the CSV file into the service
	questions, err := utils.ReadCSV("questions.csv")
	if err != nil {
		log.Fatalf("Error reading CSV: %v", err)
	}
	quizService.LoadQuestions(questions)
	quizHandler := handlers.NewQuizHandler(quizService)

	authService := &services.AuthService{}
	authHandler := handlers.NewAuthHandler(authService)

	statsService := &services.StatsService{}
	statsHandler := handlers.NewStatsHandler(statsService)

	// Initialize router
	r := mux.NewRouter()

	// Configure session store for middleware and handlers
	handlers.SessionStore = sessions.NewCookieStore([]byte("quiz-secret"))
	middleware.SetSessionStore(handlers.SessionStore)

	// Public routes
	r.HandleFunc("/register", authHandler.RegisterUser).Methods("POST")
	r.HandleFunc("/login", authHandler.LoginUser).Methods("POST")
	r.HandleFunc("/questions", quizHandler.GetQuestions).Methods("GET")

	// Protected routes
	api := r.PathPrefix("/quiz").Subrouter()
	api.Use(middleware.AuthMiddleware)
	api.HandleFunc("/start", quizHandler.StartQuiz).Methods("POST")
	api.HandleFunc("/next", quizHandler.NextQuestion).Methods("GET")
	api.HandleFunc("/submit", quizHandler.SubmitAnswer).Methods("POST")
	api.HandleFunc("/results", quizHandler.GetResults).Methods("GET")
	api.HandleFunc("/stats", statsHandler.GetStats).Methods("GET")

	// Swagger routes
	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	// Start server
	fmt.Println("Server is running on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", r))
}
