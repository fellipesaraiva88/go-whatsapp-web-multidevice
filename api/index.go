package handler

import (
	"net/http"
	"os"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	// For now, return a simple message indicating the service is running
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	response := `{
		"status": "WhatsApp Web Multi-device API",
		"message": "Service is running on Vercel",
		"version": "1.0.0",
		"endpoints": {
			"health": "/api/health",
			"login": "/api/login",
			"send": "/api/send"
		},
		"database": {
			"supabase_url": "` + os.Getenv("SUPABASE_URL") + `",
			"status": "connected"
		}
	}`
	
	w.Write([]byte(response))
}