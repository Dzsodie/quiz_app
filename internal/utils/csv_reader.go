package utils

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"

	"github.com/Dzsodie/quiz_app/internal/models"
)

func ReadCSV(filename string) ([]models.Question, error) {
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

	if len(records) == 0 {
		return nil, fmt.Errorf("CSV file is empty")
	}

	var questions []models.Question
	for i, record := range records {
		if i == 0 { // Skip header
			continue
		}
		if len(record) < 5 {
			return nil, fmt.Errorf("invalid record: %v", record)
		}
		answer, err := strconv.Atoi(record[4])
		if err != nil {
			return nil, fmt.Errorf("invalid answer format in record %v: %w", record, err)
		}
		questions = append(questions, models.Question{
			Question: record[0],
			Options:  record[1:4],
			Answer:   answer,
		})
	}
	return questions, nil
}
