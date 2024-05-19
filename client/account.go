package client

import (
	"fmt"
	"time"
)

type Account struct {
	Id             string    `json:"id"`
	Username       string    `json:"username"`
	Acct           string    `json:"acct"`
	DisplayName    string    `json:"display_name"`
	Locked         bool      `json:"locked"`
	Bot            bool      `json:"bot"`
	Discoverable   bool      `json:"discoverable"`
	Group          bool      `json:"group"`
	CreatedAt      time.Time `json:"created_at"`
	Note           string    `json:"note"`
	URL            string    `json:"url"`
	URI            string    `json:"uri"`
	Avatar         string    `json:"avatar"`
	AvatarStatic   string    `json:"avatar_static"`
	Header         string    `json:"header"`
	HeaderStatic   string    `json:"header_static"`
	FollowersCount int       `json:"followers_count"`
	FollowingCount int       `json:"following_count"`
	StatusesCount  int       `json:"statuses_count"`
	LastStatusAt   string    `json:"last_status_at"`
}

func FetchAccount(baseURL string, handle string) (Account, error) {
	var account Account
	lookupUrl := fmt.Sprintf(
		"%s/api/v1/accounts/lookup?acct=%s",
		baseURL,
		handle,
	)

	headers := make(map[string]string)
	err := Fetch(lookupUrl, &account, headers)

	if err != nil {
		return account, err
	}

	return account, nil
}
