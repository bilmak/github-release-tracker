package repo

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"
)

var ErrRateLimit = errors.New("github rate limit exceeded")

type Client struct {
	httpClient *http.Client
	token      string
}

func NewClient(token string) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		token:      token,
	}
}

func (c *Client) RepoExists(ctx context.Context, repo string) (bool, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s", repo)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusOK:
		return true, nil
	case http.StatusNotFound:
		return false, nil
	case http.StatusTooManyRequests, http.StatusForbidden:
		return false, ErrRateLimit
	default:
		return false, fmt.Errorf("github API returned status %d", resp.StatusCode)
	}
}
