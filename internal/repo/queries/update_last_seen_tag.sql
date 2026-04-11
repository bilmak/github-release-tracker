UPDATE subscriptions
SET
  last_seen_tag = $1
WHERE
  id = $2;
