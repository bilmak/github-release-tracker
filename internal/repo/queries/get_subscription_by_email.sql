SELECT id, email, repo, confirmed, confirm_token, unsubscribe_token, last_seen_tag
FROM subscriptions 
WHERE email = $1;
