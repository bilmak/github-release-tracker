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

func (s *Storage) SaveSubscription(ctx context.Context, subs domain.Subscription) error {
	_, err := s.db.Exec(ctx, `INSERT INTO subscriptions (id, email, repo, confirmed, confirm_token, unsubscribe_token, last_seen_tag)                         
                 VALUES ($1, $2, $3, $4, $5, $6, $7)`, subs.ID, subs.Email, subs.Repo, subs.Confirmed, subs.ConfirmToken, subs.UnsubscribeToken, subs.LastSeenTag,
	)
	return err
}

func (s *Storage) ConfirmSubscription(ctx context.Context, token string) error {
	result, err := s.db.Exec(ctx, `UPDATE subscriptions SET confirmed = true WHERE confirm_token = $1 AND confirmed = false`, token)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Storage) Unsubscribe(ctx context.Context, token string) error {
	res, err := s.db.Exec(ctx, `DELETE FROM subscriptions WHERE unsubscribe_token = $1`, token)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
