package checker

import (
	"context"
	"testing"
	"time"

	"github.com/bilmak/github-release-notifier/internal/domain"
)

type fakeGitHub struct {
	tag string
	err error
}

func (f *fakeGitHub) LatestRelease(ctx context.Context, repo string) (string, error) {
	return f.tag, f.err
}

type fakeStorage struct {
	repos      []string
	subs       []domain.Subscription
	updatedIDs []string
}

func (f *fakeStorage) GetTrackedRepos(ctx context.Context) ([]string, error) {
	return f.repos, nil
}

func (f *fakeStorage) GetConfirmed(ctx context.Context, repo string) ([]domain.Subscription, error) {
	return f.subs, nil
}

func (f *fakeStorage) UpdateLastSeenTag(ctx context.Context, id, tag string) error {
	f.updatedIDs = append(f.updatedIDs, id)
	return nil
}

type fakeEmail struct {
	sentTo []string
}

func (f *fakeEmail) SendNotification(to, repo, tag, unsubToken string) error {
	f.sentTo = append(f.sentTo, to)
	return nil
}

func TestCheck_NewRelease_SendsEmail(t *testing.T) {
	gh := &fakeGitHub{tag: "v2.0.0"}
	st := &fakeStorage{
		repos: []string{"owner/repo"},
		subs: []domain.Subscription{
			{ID: "1", Email: "user@test.com", LastSeenTag: "v1.0.0", UnsubscribeToken: "tok1"},
		},
	}
	em := &fakeEmail{}
	ch := New(gh, st, em, 1*time.Hour)

	ch.check(context.Background())

	if len(em.sentTo) != 1 {
		t.Fatalf("want 1 email sent, got %d", len(em.sentTo))
	}
	if em.sentTo[0] != "user@test.com" {
		t.Fatalf("want email to user@test.com, got %s", em.sentTo[0])
	}
}

func TestCheck_NewRelease_UpdatesTag(t *testing.T) {
	gh := &fakeGitHub{tag: "v2.0.0"}
	st := &fakeStorage{
		repos: []string{"owner/repo"},
		subs: []domain.Subscription{
			{ID: "1", Email: "user@test.com", LastSeenTag: "v1.0.0", UnsubscribeToken: "tok1"},
		},
	}
	em := &fakeEmail{}
	ch := New(gh, st, em, 1*time.Hour)

	ch.check(context.Background())

	if len(st.updatedIDs) != 1 {
		t.Fatalf("want 1 tag updated, got %d", len(st.updatedIDs))
	}
	if st.updatedIDs[0] != "1" {
		t.Fatalf("want updated ID 1, got %s", st.updatedIDs[0])
	}
}

func TestCheck_SameTag_NoEmail(t *testing.T) {
	gh := &fakeGitHub{tag: "v1.0.0"}
	st := &fakeStorage{
		repos: []string{"owner/repo"},
		subs: []domain.Subscription{
			{ID: "1", Email: "user@test.com", LastSeenTag: "v1.0.0", UnsubscribeToken: "tok1"},
		},
	}
	em := &fakeEmail{}
	ch := New(gh, st, em, 1*time.Hour)

	ch.check(context.Background())

	if len(em.sentTo) != 0 {
		t.Fatalf("want 0 emails, got %d", len(em.sentTo))
	}
}

func TestCheck_NoRepos_NoEmail(t *testing.T) {
	gh := &fakeGitHub{tag: "v1.0.0"}
	st := &fakeStorage{repos: []string{}}
	em := &fakeEmail{}
	ch := New(gh, st, em, 1*time.Hour)

	ch.check(context.Background())

	if len(em.sentTo) != 0 {
		t.Fatalf("want 0 emails, got %d", len(em.sentTo))
	}
}

func TestCheck_NoRelease_NoEmail(t *testing.T) {
	gh := &fakeGitHub{tag: ""}
	st := &fakeStorage{
		repos: []string{"owner/repo"},
		subs: []domain.Subscription{
			{ID: "1", Email: "user@test.com", LastSeenTag: "", UnsubscribeToken: "tok1"},
		},
	}
	em := &fakeEmail{}
	ch := New(gh, st, em, 1*time.Hour)

	ch.check(context.Background())

	if len(em.sentTo) != 0 {
		t.Fatalf("want 0 emails, got %d", len(em.sentTo))
	}
}
