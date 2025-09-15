package handler

import (
	"encoding/json"
	"net/http"
	"time"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	response := map[string]interface{}{
		"status": "ok",
		"message": "WhatsApp API is running",
		"timestamp": time.Now().Unix(),
		"endpoints": []string{
			"/api/health",
			"/api/status", 
			"/api/login",
			"/api/auth/login",
		},
	}
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}