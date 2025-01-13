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
	"strconv"

	_ "github.com/Dzsodie/quiz_app/docs" // Import generated Swagger docs
	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
)

type Question struct {
	Question string   `json:"question"`
	Options  []string `json:"options"`
	Answer   int      `json:"answer"`
}

var questions []Question

func readCSV(filename string) ([]Question, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var questions []Question
	for i, record := range records {
		if i == 0 {
			// Skip the header row
			continue
		}
		if len(record) < 5 {
			return nil, fmt.Errorf("invalid record: %v", record)
		}
		answer, err := strconv.Atoi(record[4])
		if err != nil {
			return nil, fmt.Errorf("invalid answer format in record: %v", record)
		}
		questions = append(questions, Question{
			Question: record[0],
			Options:  record[1:4],
			Answer:   answer,
		})
	}
	return questions, nil
}

// @Summary Get all questions
// @Description Get the list of all questions
// @Tags questions
// @Accept  json
// @Produce  json
// @Success 200 {array} Question
// @Router /questions [get]
func getQuestions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(questions)
}

// @title Questions API
// @version 1.0
// @description This is a sample server for managing questions.
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

	// API routes
	r.HandleFunc("/questions", getQuestions).Methods("GET")

	// Swagger route
	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	fmt.Println("Server is running on port 8080...")
	if err := http.ListenAndServe(":8080", r); err != nil {
		fmt.Println("Error starting server:", err)
	}
}
