package health

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/Dzsodie/quiz_app/internal/utils"

	"github.com/gorilla/sessions"
)

type HealthStatus struct {
	InMemoryDB string `json:"in_memory_db"`
	Sessions   string `json:"sessions"`
	Mutex      string `json:"mutex"`
}

type HealthCheck struct {
	SessionStore *sessions.CookieStore
	Mutex        *sync.Mutex
	InMemoryDB   map[string]string
}

func NewHealthCheck(sessionStore *sessions.CookieStore, inMemoryDB map[string]string) *HealthCheck {
	return &HealthCheck{
		SessionStore: sessionStore,
		Mutex:        &sync.Mutex{},
		InMemoryDB:   inMemoryDB,
	}
}

func (hc *HealthCheck) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	logger := utils.GetLogger().Sugar()

	status := HealthStatus{
		InMemoryDB: hc.checkInMemoryDB(),
		Sessions:   hc.checkSessionStore(),
		Mutex:      hc.checkMutexState(),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(status); err != nil {
		logger.Warnw("Failed to encode health status: %v", err)
	}

	logger.Infow("Health check executed", "status", status)
}

func (hc *HealthCheck) checkInMemoryDB() string {
	if _, exists := hc.InMemoryDB["key"]; exists {
		return "OK"
	}
	return "Not Initialized"
}

func (hc *HealthCheck) checkSessionStore() string {
	if hc.SessionStore != nil {
		return "OK"
	}
	return "Not Configured"
}

func (hc *HealthCheck) checkMutexState() string {
	if hc.Mutex.TryLock() {
		defer hc.Mutex.Unlock()
		return "Unlocked"
	}
	return "Locked"
}
