DELETE FROM subscriptions
WHERE
  unsubscribe_token = $1;
