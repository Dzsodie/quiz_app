package database

import (
	"errors"
	"sync"
)

type MemoryDB struct {
	questions map[string]Question
	users     map[string]User
	mu        sync.RWMutex
}

func NewMemoryDB() *MemoryDB {
	return &MemoryDB{questions: make(map[string]Question), users: make(map[string]User)}
}

func (db *MemoryDB) AddUser(user User) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, exists := db.users[user.Username]; exists {
		return errors.New("user already exists")
	}
	db.users[user.Username] = user
	return nil
}

func (db *MemoryDB) GetUser(username string) (User, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	user, exists := db.users[username]
	if !exists {
		return User{}, errors.New("user not found")
	}
	return user, nil
}

func (db *MemoryDB) GetAllUsers() []User {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var users []User
	for _, user := range db.users {
		users = append(users, user)
	}
	return users
}

func (db *MemoryDB) GetQuestion(id string) (Question, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	question, exists := db.questions[id]
	if !exists {
		return Question{}, errors.New("question not found")
	}
	return question, nil
}

func (db *MemoryDB) ListQuestions() ([]Question, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var questions []Question
	for _, q := range db.questions {
		questions = append(questions, q)
	}
	return questions, nil
}

// Clear clears all data from the in-memory database
func (db *MemoryDB) Clear() {
	db.mu.Lock()
	defer db.mu.Unlock()

	for k := range db.questions {
		delete(db.questions, k)
	}
	for k := range db.users {
		delete(db.users, k)
	}
}
