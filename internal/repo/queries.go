package repo

import _ "embed"

//go:embed queries/insert_subscription.sql
var InsertSubscriptionSQL string

//go:embed queries/confirm_subscription.sql
var ConfirmSubscriptionSQL string

//go:embed queries/unsubscribe.sql
var UnsubscribeSQL string

//go:embed queries/get_confirmed.sql
var GetConfirmedSQL string

//go:embed queries/update_last_seen_tag.sql
var UpdateLastSeenTagSQL string

//go:embed queries/get_tracked_repos.sql
var GetTrackedReposSQL string

//go:embed queries/get_subscription_by_email.sql
var GetSubscriptionsByEmailSQL string

//go:embed migrations/create_subscriptions.sql
var MigrationSQL string
