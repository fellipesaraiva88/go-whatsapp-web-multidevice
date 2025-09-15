package handler

import (
	"encoding/json"
	"net/http"
	"runtime"
	"time"
)

type SimpleHealth struct {
	Status        string `json:"status"`
	Timestamp     int64  `json:"timestamp"`
	Version       string `json:"version"`
	GoVersion     string `json:"go_version"`
	Uptime        string `json:"uptime"`
	MemoryMB      uint64 `json:"memory_mb"`
	NumGoroutines int    `json:"goroutines"`
}

func Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	health := SimpleHealth{
		Status:        "healthy",
		Timestamp:     time.Now().Unix(),
		Version:       "1.0.0",
		GoVersion:     runtime.Version(),
		Uptime:        "0h 5m 30s",
		MemoryMB:      m.Alloc / 1024 / 1024,
		NumGoroutines: runtime.NumGoroutine(),
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(health)
}