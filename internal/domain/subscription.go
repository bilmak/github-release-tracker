package domain

type Subscription struct {
	ID               string
	Email            string
	Repo             string
	Confirmed        bool
	ConfirmToken     string
	UnsubscribeToken string
	LastSeenTag      string
}

type SubscribeRequest struct {
	Email string `json:"email"`
	Repo  string `json:"repo"`
}
