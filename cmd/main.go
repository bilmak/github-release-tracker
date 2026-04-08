package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/bilmak/github-release-notifier/internal/handler"
)

func main() {
	mux := http.NewServeMux()
	hand := handler.New()
	mux.HandleFunc("POST /api/subscribe", hand.Subscribe)
	mux.HandleFunc("/health", healthHandler)
	log.Println("Server running on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{
		"status": "ok",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
