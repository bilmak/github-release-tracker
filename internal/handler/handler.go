package handler

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/bilmak/github-release-notifier/internal/domain"
	"github.com/bilmak/github-release-notifier/internal/repo"
)

type GitHubInter interface {
	RepoExists(ctx context.Context, repo string) (bool, error)
}
type SubscriptionStorage interface {
	SaveSubscription(ctx context.Context, sub domain.Subscription) error
	ConfirmSubscription(ctx context.Context, token string) error
	Unsubscribe(ctx context.Context, token string) error
}
type EmailInter interface {
	SendConfirmation(to, token string) error
}
type Handler struct {
	github  GitHubInter
	storage SubscriptionStorage
	email   EmailInter
}

func New(gh GitHubInter, st SubscriptionStorage, em EmailInter) *Handler {
	return &Handler{github: gh, storage: st, email: em}
}

func (h *Handler) Subscribe(w http.ResponseWriter, r *http.Request) {
	var req domain.SubscribeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}
	if !emailRegex.MatchString(req.Email) {
		http.Error(w, `{"error":"invalid email format"}`,
			http.StatusBadRequest)
		return
	}
	if !isValid(req.Repo) {
		http.Error(w, `{"error":"invalid repo format"}`, http.StatusBadRequest)
		return
	}
	exist, err := h.github.RepoExists(r.Context(), req.Repo)
	if err != nil {
		if errors.Is(err, repo.ErrRateLimit) {
			http.Error(w, `{"error":"GitHub API rate limit exceeded"}`,
				http.StatusTooManyRequests)
			return
		}
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}
	if !exist {
		http.Error(w, `{"error":"repository not found"}`, http.StatusNotFound)
		return
	}
	sub := domain.Subscription{
		ID:               generateID(),
		Email:            req.Email,
		Repo:             req.Repo,
		ConfirmToken:     generateToken(),
		UnsubscribeToken: generateToken(),
	}

	if err := h.storage.SaveSubscription(r.Context(), sub); err != nil {
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}

	if err := h.email.SendConfirmation(sub.Email, sub.ConfirmToken); err != nil {
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Subscription successful. Confirmation email sent.",
	})
}

func (h *Handler) Confirm(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, `{"error":"missing token"}`, http.StatusBadRequest)
		return
	}
	if err := h.storage.ConfirmSubscription(r.Context(), token); err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			http.Error(w, `{"error":"invalid or expired token"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Subscription confirmed",
	})
}

func (h *Handler) Unsubscribe(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, `{"error":"missing token"}`, http.StatusBadRequest)
		return
	}

	if err := h.storage.Unsubscribe(r.Context(), token); err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			http.Error(w, `{"error":"subscription not found"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Successfully unsubscribed.",
	})
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func isValid(repo string) bool {
	parts := strings.Split(repo, "/")
	return len(parts) == 2 && parts[0] != "" && parts[1] != ""
}

func generateToken() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

func generateID() string {
	return generateToken()
}
