UPDATE subscriptions
SET
  confirmed = true
WHERE
  confirm_token = $1
  AND confirmed = false;
