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
// @contact.url http://www.zsuzsa-makara.com
// @contact.email dzsodie@gmail.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /
func main() {
	// Load questions from the CSV file into the service
	questions, err := utils.ReadCSV("questions.csv")
	if err != nil {
		log.Fatalf("Error reading CSV: %v", err)
	}
	services.LoadQuestions(questions)

	// Initialize router
	r := mux.NewRouter()

	// Configure session store for middleware and handlers
	handlers.SessionStore = sessions.NewCookieStore([]byte("quiz-secret"))
	middleware.SetSessionStore(handlers.SessionStore)

	// Public routes
	r.HandleFunc("/register", handlers.RegisterUser).Methods("POST")
	r.HandleFunc("/login", handlers.LoginUser).Methods("POST")
	r.HandleFunc("/questions", handlers.GetQuestions).Methods("GET")

	// Protected routes
	api := r.PathPrefix("/quiz").Subrouter()
	api.Use(middleware.AuthMiddleware)
	api.HandleFunc("/start", handlers.StartQuiz).Methods("POST")
	api.HandleFunc("/next", handlers.NextQuestion).Methods("GET")
	api.HandleFunc("/submit", handlers.SubmitAnswer).Methods("POST")
	api.HandleFunc("/results", handlers.GetResults).Methods("GET")
	api.HandleFunc("/stats", handlers.GetStats).Methods("GET")

	// Swagger routes
	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	// Start server
	fmt.Println("Server is running on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", r))
}
