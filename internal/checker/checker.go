package checker

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/bilmak/github-release-notifier/internal/domain"
	"github.com/bilmak/github-release-notifier/pkg/github"
)

type GitHub interface {
	LatestRelease(ctx context.Context, repo string) (string, error)
}

type Storage interface {
	GetTrackedRepos(ctx context.Context) ([]string, error)
	GetConfirmed(ctx context.Context, repository string) ([]domain.Subscription, error)
	UpdateLastSeenTag(ctx context.Context, id, tag string) error
}

type Email interface {
	SendNotification(to, repo, tag, unsubToken string) error
}

type Checker struct {
	github   GitHub
	storage  Storage
	email    Email
	interval time.Duration
}

func New(gh GitHub, st Storage, em Email, in time.Duration) *Checker {
	return &Checker{github: gh, storage: st, email: em, interval: in}
}

// worker
func (c *Checker) Run(ctx context.Context) {
	c.check(ctx)

	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.check(ctx)
		}
	}
}

func (c *Checker) check(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, c.interval/2)
	defer cancel()

	repos, err := c.storage.GetTrackedRepos(ctx)
	if err != nil {
		log.Printf("checker: failed to get repos: %v", err)
		return
	}
	for _, repoName := range repos {
		tag, err := c.github.LatestRelease(ctx, repoName)
		if err != nil {
			log.Printf("checker: failed to get release for %s: %v", repoName, err)
			if errors.Is(err, github.ErrRateLimit) {
				log.Printf("checker: rate limited, stopping this cycle")
				return
			}
			continue
		}
		if tag == "" {
			continue
		}
		subscriptions, err := c.storage.GetConfirmed(ctx, repoName)
		if err != nil {
			log.Printf("checker: failed to get subscription for %s: %v", repoName, err)
			continue
		}
		for _, sub := range subscriptions {
			if sub.LastSeenTag == tag {
				continue
			}
			if err := c.email.SendNotification(sub.Email, repoName, tag, sub.UnsubscribeToken); err != nil {
				log.Printf("checker: failed to send email to %s: %v", sub.Email, err)
				continue
			}
			if err := c.storage.UpdateLastSeenTag(ctx, sub.ID, tag); err != nil {
				log.Printf("checker: failed to update tag for %s: %v", sub.ID, err)
			}
		}
	}

}
