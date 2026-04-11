package handler

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bilmak/github-release-notifier/internal/domain"
	"github.com/bilmak/github-release-notifier/internal/repo"
	"github.com/bilmak/github-release-notifier/pkg/github"
)

type fakeGitHub struct {
	exists bool
	err    error
}

func (f *fakeGitHub) RepoExists(ctx context.Context, r string) (bool, error) {
	return f.exists, f.err
}

type fakeStorage struct {
	saveErr    error
	confirmErr error
	unsubErr   error
	subs       []domain.Subscription
}

func (f *fakeStorage) SaveSubscription(ctx context.Context, sub domain.Subscription) error {
	return f.saveErr
}

func (f *fakeStorage) ConfirmSubscription(ctx context.Context, token string) error {
	return f.confirmErr
}

func (f *fakeStorage) Unsubscribe(ctx context.Context, token string) error {
	return f.unsubErr
}

func (f *fakeStorage) GetSubscriptionsByEmail(ctx context.Context, email string) ([]domain.Subscription, error) {
	return f.subs, nil
}

type fakeEmail struct {
	err error
}

func (f *fakeEmail) SendConfirmation(to, token string) error {
	return f.err
}

func TestSubscribe_OK(t *testing.T) {
	gh := &fakeGitHub{exists: true}
	st := &fakeStorage{}
	em := &fakeEmail{}
	h := New(gh, st, em)

	body := `{"email":"user@example.com","repo":"owner/repo"}`
	r := httptest.NewRequest(http.MethodPost, "/api/subscribe", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.Subscribe(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("want status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestSubscribe_BadJSON(t *testing.T) {
	gh := &fakeGitHub{exists: true}
	st := &fakeStorage{}
	em := &fakeEmail{}
	h := New(gh, st, em)

	r := httptest.NewRequest(http.MethodPost, "/api/subscribe", bytes.NewBufferString("not json"))
	w := httptest.NewRecorder()

	h.Subscribe(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("want status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestSubscribe_BadEmail(t *testing.T) {
	gh := &fakeGitHub{exists: true}
	st := &fakeStorage{}
	em := &fakeEmail{}
	h := New(gh, st, em)

	body := `{"email":"not-an-email","repo":"owner/repo"}`
	r := httptest.NewRequest(http.MethodPost, "/api/subscribe", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.Subscribe(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("want status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestSubscribe_BadRepo(t *testing.T) {
	gh := &fakeGitHub{exists: true}
	st := &fakeStorage{}
	em := &fakeEmail{}
	h := New(gh, st, em)

	body := `{"email":"user@example.com","repo":"justrepo"}`
	r := httptest.NewRequest(http.MethodPost, "/api/subscribe", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.Subscribe(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("want status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestSubscribe_RepoNotFound(t *testing.T) {
	gh := &fakeGitHub{exists: false}
	st := &fakeStorage{}
	em := &fakeEmail{}
	h := New(gh, st, em)

	body := `{"email":"user@example.com","repo":"owner/repo"}`
	r := httptest.NewRequest(http.MethodPost, "/api/subscribe", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.Subscribe(w, r)

	if w.Code != http.StatusNotFound {
		t.Fatalf("want status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestSubscribe_RateLimit(t *testing.T) {
	gh := &fakeGitHub{err: github.ErrRateLimit}
	st := &fakeStorage{}
	em := &fakeEmail{}
	h := New(gh, st, em)

	body := `{"email":"user@example.com","repo":"owner/repo"}`
	r := httptest.NewRequest(http.MethodPost, "/api/subscribe", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.Subscribe(w, r)

	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("want status %d, got %d", http.StatusTooManyRequests, w.Code)
	}
}

func TestConfirm_OK(t *testing.T) {
	gh := &fakeGitHub{}
	st := &fakeStorage{}
	em := &fakeEmail{}
	h := New(gh, st, em)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/confirm/{token}", h.Confirm)

	r := httptest.NewRequest(http.MethodGet, "/api/confirm/abc123", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("want status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestConfirm_BadToken(t *testing.T) {
	gh := &fakeGitHub{}
	st := &fakeStorage{confirmErr: repo.ErrNotFound}
	em := &fakeEmail{}
	h := New(gh, st, em)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/confirm/{token}", h.Confirm)

	r := httptest.NewRequest(http.MethodGet, "/api/confirm/wrong", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)

	if w.Code != http.StatusNotFound {
		t.Fatalf("want status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestUnsubscribe_OK(t *testing.T) {
	gh := &fakeGitHub{}
	st := &fakeStorage{}
	em := &fakeEmail{}
	h := New(gh, st, em)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/unsubscribe/{token}", h.Unsubscribe)

	r := httptest.NewRequest(http.MethodGet, "/api/unsubscribe/abc123", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("want status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestUnsubscribe_NotFound(t *testing.T) {
	gh := &fakeGitHub{}
	st := &fakeStorage{unsubErr: repo.ErrNotFound}
	em := &fakeEmail{}
	h := New(gh, st, em)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/unsubscribe/{token}", h.Unsubscribe)

	r := httptest.NewRequest(http.MethodGet, "/api/unsubscribe/wrong", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)

	if w.Code != http.StatusNotFound {
		t.Fatalf("want status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestGetSubscriptions_OK(t *testing.T) {
	gh := &fakeGitHub{}
	st := &fakeStorage{
		subs: []domain.Subscription{
			{Email: "user@example.com", Repo: "owner/repo", Confirmed: true},
		},
	}
	em := &fakeEmail{}
	h := New(gh, st, em)

	r := httptest.NewRequest(http.MethodGet, "/api/subscriptions?email=user@example.com", nil)
	w := httptest.NewRecorder()

	h.GetSubscriptions(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("want status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestGetSubscriptions_BadEmail(t *testing.T) {
	gh := &fakeGitHub{}
	st := &fakeStorage{}
	em := &fakeEmail{}
	h := New(gh, st, em)

	r := httptest.NewRequest(http.MethodGet, "/api/subscriptions?email=not-valid", nil)
	w := httptest.NewRecorder()

	h.GetSubscriptions(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("want status %d, got %d", http.StatusBadRequest, w.Code)
	}
}
