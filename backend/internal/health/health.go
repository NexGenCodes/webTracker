package health

import (
	"encoding/json"
	"net/http"
	"sync/atomic"
	"time"

	"webtracker-bot/internal/logger"
)

var (
	startTime    = time.Now()
	requestCount int64
	errorCount   int64
	isHealthy    int32 = 1 // 1 = healthy, 0 = unhealthy
)

type HealthStatus struct {
	Status          string `json:"status"`
	Uptime          string `json:"uptime"`
	RequestsHandled int64  `json:"requests_handled"`
	Errors          int64  `json:"errors"`
	DatabasePingMs  int64  `json:"database_ping_ms"`
	Timestamp       string `json:"timestamp"`
}

// MarkUnhealthy marks the service as unhealthy (called on critical failures)
func MarkUnhealthy() {
	atomic.StoreInt32(&isHealthy, 0)
}

// MarkHealthy marks the service as healthy
func MarkHealthy() {
	atomic.StoreInt32(&isHealthy, 1)
}

// IncRequest increments the request counter
func IncRequest() {
	atomic.AddInt64(&requestCount, 1)
}

// IncError increments the error counter
func IncError() {
	atomic.AddInt64(&errorCount, 1)
}

// StartHealthServer starts the health check HTTP server
func StartHealthServer(port string, dbPingFn func() error) {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		handleHealth(w, r, dbPingFn)
	})

	http.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		handleReadiness(w, r, dbPingFn)
	})

	logger.Info().Str("port", port).Msg("Health check server starting")
	go func() {
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			logger.Error().Err(err).Msg("Health server failed")
		}
	}()
}

func handleHealth(w http.ResponseWriter, r *http.Request, dbPingFn func() error) {
	// Basic liveness check
	if atomic.LoadInt32(&isHealthy) == 0 {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("UNHEALTHY"))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func handleReadiness(w http.ResponseWriter, r *http.Request, dbPingFn func() error) {
	status := HealthStatus{
		Status:          "healthy",
		Uptime:          time.Since(startTime).Round(time.Second).String(),
		RequestsHandled: atomic.LoadInt64(&requestCount),
		Errors:          atomic.LoadInt64(&errorCount),
		Timestamp:       time.Now().UTC().Format(time.RFC3339),
	}

	// Check database
	start := time.Now()
	if err := dbPingFn(); err != nil {
		status.Status = "degraded"
		status.DatabasePingMs = -1
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		status.DatabasePingMs = time.Since(start).Milliseconds()
		w.WriteHeader(http.StatusOK)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}
