/*
Copyright © 2025 Zsuzsa Makara <dzsodie@gmail.com>
*/
package main

import (
	"encoding/csv"
	"fmt"
	"os"
)

type Question struct {
	Question string
	Options  []string
	Answer   int
}

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
	for _, record := range records {
		answer := int(record[4][0] - '0')
		questions = append(questions, Question{
			Question: record[0],
			Options:  record[1:4],
			Answer:   answer,
		})
	}
	return questions, nil
}

func main() {
	questions, err := readCSV("questions.csv")
	if err != nil {
		fmt.Println("Error reading CSV:", err)
		return
	}

	fmt.Println("Questions loaded:", questions)
}
