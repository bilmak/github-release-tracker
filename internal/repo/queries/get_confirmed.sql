SELECT
  id,
  email,
  repo,
  confirmed,
  confirm_token,
  unsubscribe_token,
  last_seen_tag
FROM
  subscriptions
WHERE
  repo = $1
  AND confirmed = true;
