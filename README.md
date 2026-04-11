# GitHub Release Notifier

Go service that allows to subscribe to email notifications about new releases of GitHub repositories.

## What does it do

User subscribes with their email and a GitHub repo (like `golang/go`). The service checks if the repo actually exists on GitHub, saves the subscription, and sends a confirmation email. After confirming, the user starts getting emails whenever there's a new release.

There's a background worker that checks GitHub for new releases every 2 hours. If it finds a new tag, it sends an email to everyone who's subscribed to that repo.

## Project structure

- `cmd/main.go` - starts the server and connects everything together
- `internal/handler/` - HTTP handlers for all endpoints
- `internal/repo/client.go` - talks to GitHub API
- `internal/repo/postgres.go` - database queries
- `internal/checker/` - background job that checks for new releases
- `internal/email/` - sends emails
- `internal/domain/` - data structures
- `migrations/` - SQL for creating the database table

## API

- `POST /api/subscribe` - subscribe to a repo (send email + repo in JSON body)
- `GET /api/confirm/{token}` - confirm your subscription (link comes in email)
- `GET /api/unsubscribe/{token}` - unsubscribe (link is in every notification email)
- `GET /api/subscriptions?email=...` - see all your active subscriptions
- `GET /health` - health check

Full spec is in `swagger.yaml`.

## How the subscription flow works

1. User sends POST with email and repo
2. We validate the email and check that repo format is `owner/repo`
3. We call GitHub API to make sure the repo exists (returns 404 if not)
4. Save to database, generate confirmation token
5. Send confirmation email with a link
6. User clicks the link -> subscription is confirmed
7. Now they'll get emails about new releases

## How the release checker works

It runs in a goroutine with a ticker (every 2 hours by default)

1. Get all repos that have at least one confirmed subscriber
2. For each repo, ask GitHub API for the latest release tag
3. Compare with `last_seen_tag` in the database
4. If it's different - send email to all subscribers and update the tag
5. If GitHub returns rate limit error (429) - stop and wait for the next cycle

## How to run

You need Docker installed.

```bash
docker compose up --build
```

This will start the app and PostgreSQL. The database migration runs automatically on startup.

## Environment variables

Create a `.env` file in the project root:

```
DATABASE_URL=postgres://postgres:postgres@db:5432/notifier?sslmode=disable
GITHUB_TOKEN=your_github_token
SMTP_FROM=your@email.com
SMTP_PASSWORD=your_password
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
BASE_URL=http://localhost:8080
CHECK_INTERVAL_MIN=120
```

`GITHUB_TOKEN` is optional but without it GitHub limits you to 60 requests per hour. With a token it's 5000.

## Tests

```bash
go test ./...
```

There are unit tests for the handler (subscribe, confirm, unsubscribe, get subscriptions) and for the checker (new release, same tag, no repos, etc). All business logic is tested through interfaces so no real database or GitHub API is needed.

## Built with

- Go, net/http for the server
- PostgreSQL + pgx for the database
- net/smtp for sending emails
- Docker for running everything
