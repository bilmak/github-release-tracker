package handler

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"

	"github.com/bilmak/github-release-notifier/internal/domain"
)

type Handler struct{}

func New() *Handler {
	return &Handler{}
}

func (h Handler) Subscribe(w http.ResponseWriter, r *http.Request) {
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
