package health

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/gorilla/sessions"
	"go.uber.org/zap"
)

// HealthStatus holds the status of application components.
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

// NewHealthCheck initializes the HealthCheck struct.
func NewHealthCheck(logger *zap.SugaredLogger, sessionStore *sessions.CookieStore, inMemoryDB map[string]string) *HealthCheck {
	return &HealthCheck{
		Logger:       logger,
		SessionStore: sessionStore,
		Mutex:        &sync.Mutex{},
		InMemoryDB:   inMemoryDB,
	}
}

// HealthCheckHandler handles the health check endpoint.
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
	// Simulate in-memory DB check
	if _, exists := hc.InMemoryDB["key"]; exists {
		return "OK"
	}
	return "Not Initialized"
}

func (hc *HealthCheck) checkSessionStore() string {
	// Simulate session store check
	if hc.SessionStore != nil {
		return "OK"
	}
	return "Not Configured"
}

func (hc *HealthCheck) checkMutexState() string {
	// Simulate mutex check
	if hc.Mutex.TryLock() {
		defer hc.Mutex.Unlock()
		return "Unlocked"
	}
	return "Locked"
}
