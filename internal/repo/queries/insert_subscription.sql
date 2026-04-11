INSERT INTO
  subscriptions (
    id,
    email,
    repo,
    confirmed,
    confirm_token,
    unsubscribe_token,
    last_seen_tag
  )
VALUES
  ($1, $2, $3, $4, $5, $6, $7);
