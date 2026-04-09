package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strings"

	"github.com/bilmak/github-release-notifier/internal/domain"
	"github.com/bilmak/github-release-notifier/internal/repo"
)

type GitHubInter interface {
	RepoExists(ctx context.Context, repo string) (bool, error)
}
type Handler struct {
	github GitHubInter
}

func New(gh GitHubInter) *Handler {
	return &Handler{github: gh}
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
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Subscription successful. Confirmation email sent.",
	})
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func isValid(repo string) bool {
	parts := strings.Split(repo, "/")
	return len(parts) == 2 && parts[0] != "" && parts[1] != ""
}
