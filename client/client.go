package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	handle  string
	baseURL string
}

type Account struct {
	Id string `json:"id"`
}

type MediaAttachment struct {
	Type        string `json:"type"`
	Url         string `json:"url"`
	Description string `json:"description"`
}

type Post struct {
	CreatedAt        time.Time         `json:"created_at"`
	Id               string            `json:"id"`
	Visibility       string            `json:"visibility"`
	InReplyToId      string            `json:"in_reply_to_id"`
	URI              string            `json:"uri"`
	Content          string            `json:"content"`
	MediaAttachments []MediaAttachment `json:"media_attachments"`
}

func New(userURL string) (Client, error) {
	var client Client
	parsedURL, err := url.Parse(userURL)

	if err != nil {
		return client, fmt.Errorf("error parsing user url: %w", err)
	}

	baseURL := fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)
	acc := strings.TrimPrefix(parsedURL.Path, "/")
	handle := strings.TrimPrefix(acc, "@")

	return Client{
		baseURL: baseURL,
		handle:  handle,
	}, nil
}

func (c Client) GetPosts(params string) ([]Post, error) {
	var posts []Post
	account, err := c.getAccount()

	if err != nil {
		return posts, err
	}

	postsUrl := fmt.Sprintf(
		"%s/api/v1/accounts/%s/statuses/%s",
		c.baseURL,
		account.Id,
		params,
	)

	if err := get(postsUrl, &posts); err != nil {
		return posts, err
	}

	return posts, nil
}

func (c Client) getAccount() (Account, error) {
	var account Account
	lookupUrl := fmt.Sprintf(
		"%s/api/v1/accounts/lookup?acct=%s",
		c.baseURL,
		c.handle,
	)

	err := get(lookupUrl, &account)

	if err != nil {
		return account, err
	}

	return account, nil
}

func get(requestUrl string, variable interface{}) error {
	res, err := http.Get(requestUrl)

	if err != nil {
		return err
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)

	if err := json.Unmarshal(body, variable); err != nil {
		return err
	}

	return nil
}
