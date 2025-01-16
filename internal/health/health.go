package health

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/gorilla/sessions"
	"go.uber.org/zap"
)

type HealthStatus struct {
	InMemoryDB string `json:"in_memory_db"`
	Sessions   string `json:"sessions"`
	Mutex      string `json:"mutex"`
}

type HealthCheck struct {
	Logger       *zap.SugaredLogger
	SessionStore *sessions.CookieStore
	Mutex        *sync.Mutex
	InMemoryDB   map[string]string
}

func NewHealthCheck(logger *zap.SugaredLogger, sessionStore *sessions.CookieStore, inMemoryDB map[string]string) *HealthCheck {
	return &HealthCheck{
		Logger:       logger,
		SessionStore: sessionStore,
		Mutex:        &sync.Mutex{},
		InMemoryDB:   inMemoryDB,
	}
}

func (hc *HealthCheck) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	status := HealthStatus{
		InMemoryDB: hc.checkInMemoryDB(),
		Sessions:   hc.checkSessionStore(),
		Mutex:      hc.checkMutexState(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)

	hc.Logger.Infow("Health check executed", "status", status)
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
