package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/bilmak/github-release-notifier/internal/checker"
	"github.com/bilmak/github-release-notifier/internal/email"
	"github.com/bilmak/github-release-notifier/internal/handler"
	"github.com/bilmak/github-release-notifier/internal/repo"
	"github.com/bilmak/github-release-notifier/pkg/github"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	pool, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	store := repo.NewStorage(pool)
	if err := store.Migrate(ctx); err != nil {
		log.Fatal("migration failed: ", err)
	}
	mail := email.NewSender(
		os.Getenv("SMTP_FROM"),
		os.Getenv("SMTP_PASSWORD"),
		os.Getenv("SMTP_HOST"),
		os.Getenv("SMTP_PORT"),
		os.Getenv("BASE_URL"),
	)

	mux := http.NewServeMux()
	gh := github.NewClient(os.Getenv("GITHUB_TOKEN"))
	hand := handler.New(gh, store, mail)

	mux.HandleFunc("POST /api/subscribe", hand.Subscribe)
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("GET /api/confirm/{token}", hand.Confirm)
	mux.HandleFunc("GET /api/unsubscribe/{token}", hand.Unsubscribe)
	mux.HandleFunc("GET /api/subscriptions", hand.GetSubscriptions)

	interval := 2 * time.Hour
	if v := os.Getenv("CHECK_INTERVAL_MIN"); v != "" {
		if mins, err := strconv.Atoi(v); err == nil {
			interval = time.Duration(mins) * time.Minute
		}
	}
	ch := checker.New(gh, store, mail, interval)
	go ch.Run(ctx)

	srv := &http.Server{Addr: ":8080", Handler: mux}

	go func() {
		<-ctx.Done()
		log.Println("Shutting down server...")
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("HTTP server shutdown error: %v", err)
		}
	}()

	log.Println("Server running on :8080")
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatal(err)
	}
	log.Println("Server stopped gracefully")
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{
		"status": "ok",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
