package repo

import (
	"context"
	"errors"

	"github.com/bilmak/github-release-notifier/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	db *pgxpool.Pool
}

func NewStorage(db *pgxpool.Pool) *Storage {
	return &Storage{db: db}
}

var ErrNotFound = errors.New("Subscription not found")

func (s *Storage) Migrate(ctx context.Context) error {
	_, err := s.db.Exec(ctx, MigrationSQL)
	return err
}

func (s *Storage) SaveSubscription(ctx context.Context, subs domain.Subscription) error {
	_, err := s.db.Exec(ctx, InsertSubscriptionSQL,
		subs.ID, subs.Email, subs.Repo, subs.Confirmed, subs.ConfirmToken, subs.UnsubscribeToken, subs.LastSeenTag,
	)
	return err
}

func (s *Storage) ConfirmSubscription(ctx context.Context, token string) error {
	result, err := s.db.Exec(ctx, ConfirmSubscriptionSQL, token)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Storage) Unsubscribe(ctx context.Context, token string) error {
	res, err := s.db.Exec(ctx, UnsubscribeSQL, token)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Storage) GetConfirmed(ctx context.Context, repository string) ([]domain.Subscription, error) {
	rows, err := s.db.Query(ctx, GetConfirmedSQL, repository)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []domain.Subscription
	for rows.Next() {
		var sub domain.Subscription
		if err := rows.Scan(&sub.ID, &sub.Email, &sub.Repo, &sub.Confirmed, &sub.ConfirmToken, &sub.UnsubscribeToken, &sub.LastSeenTag); err != nil {
			return nil, err
		}
		subs = append(subs, sub)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return subs, nil
}

func (s *Storage) UpdateLastSeenTag(ctx context.Context, id, tag string) error {
	_, err := s.db.Exec(ctx, UpdateLastSeenTagSQL, tag, id)
	return err
}

func (s *Storage) GetTrackedRepos(ctx context.Context) ([]string, error) {
	rows, err := s.db.Query(ctx, GetTrackedReposSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var repos []string
	for rows.Next() {
		var r string
		if err := rows.Scan(&r); err != nil {
			return nil, err
		}
		repos = append(repos, r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return repos, nil
}

func (s *Storage) GetSubscriptionsByEmail(ctx context.Context, email string) ([]domain.Subscription, error) {
	rows, err := s.db.Query(ctx, GetSubscriptionsByEmailSQL, email)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []domain.Subscription
	for rows.Next() {
		var sub domain.Subscription
		if err := rows.Scan(&sub.ID, &sub.Email, &sub.Repo, &sub.Confirmed, &sub.ConfirmToken, &sub.UnsubscribeToken, &sub.LastSeenTag); err != nil {
			return nil, err
		}
		subs = append(subs, sub)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return subs, nil
}
