package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/bilmak/github-release-notifier/internal/email"
	"github.com/bilmak/github-release-notifier/internal/handler"
	"github.com/bilmak/github-release-notifier/internal/repo"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {

	pool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	store := repo.NewStorage(pool)
	mail := email.NewSender(
		os.Getenv("SMTP_FROM"),
		os.Getenv("SMTP_PASSWORD"),
		os.Getenv("SMTP_HOST"),
		os.Getenv("SMTP_PORT"),
	)

	mux := http.NewServeMux()
	gh := repo.NewClient(os.Getenv("GITHUB_TOKEN"))
	hand := handler.New(gh, store, mail)

	mux.HandleFunc("POST /api/subscribe", hand.Subscribe)
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("GET /api/confirm", hand.Confirm)
	mux.HandleFunc("GET /api/unsubscribe", hand.Unsubscribe)
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
