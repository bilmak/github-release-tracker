package handler

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/mail"
	"strings"

	"github.com/bilmak/github-release-notifier/internal/domain"
	"github.com/bilmak/github-release-notifier/internal/repo"
	"github.com/bilmak/github-release-notifier/pkg/github"
)

type GitHubInter interface {
	RepoExists(ctx context.Context, repo string) (bool, error)
}
type SubscriptionStorage interface {
	SaveSubscription(ctx context.Context, sub domain.Subscription) error
	ConfirmSubscription(ctx context.Context, token string) error
	Unsubscribe(ctx context.Context, token string) error
	GetSubscriptionsByEmail(ctx context.Context, email string) ([]domain.Subscription, error)
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
	r.Body = http.MaxBytesReader(w, r.Body, 1048576) // 1MB limit
	var req domain.SubscribeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}
	if !isEmailValid(req.Email) {
		http.Error(w, `{"error":"invalid email format"}`,
			http.StatusBadRequest)
		return
	}
	if !isRepoValid(req.Repo) {
		http.Error(w, `{"error":"invalid repo format"}`, http.StatusBadRequest)
		return
	}
	exist, err := h.github.RepoExists(r.Context(), req.Repo)
	if err != nil {
		if errors.Is(err, github.ErrRateLimit) {
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
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "UNIQUE constraint") {
			http.Error(w, `{"error":"already subscribed to this repository"}`, http.StatusConflict)
			return
		}
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
	token := r.PathValue("token")
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
	token := r.PathValue("token")
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

type subscriptionResponse struct {
	Email       string `json:"email"`
	Repo        string `json:"repo"`
	Confirmed   bool   `json:"confirmed"`
	LastSeenTag string `json:"last_seen_tag"`
}

func (h *Handler) GetSubscriptions(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	if !isEmailValid(email) {
		http.Error(w, `{"error":"invalid email"}`, http.StatusBadRequest)
		return
	}
	subs, err := h.storage.GetSubscriptionsByEmail(r.Context(), email)
	if err != nil {
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}
	result := make([]subscriptionResponse, len(subs))
	for i, sub := range subs {
		result[i] = subscriptionResponse{
			Email:       sub.Email,
			Repo:        sub.Repo,
			Confirmed:   sub.Confirmed,
			LastSeenTag: sub.LastSeenTag,
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func isRepoValid(repo string) bool {
	parts := strings.Split(repo, "/")
	return len(parts) == 2 && parts[0] != "" && parts[1] != ""
}

func isEmailValid(address string) bool {
	_, err := mail.ParseAddress(address)
	if err != nil {
		return false
	}
	return true

}

func generateToken() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

func generateID() string {
	return generateToken()
}
