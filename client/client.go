<<<<<<< Updated upstream
package client

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	handle  string
	baseURL string
}

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
	LastStatusAt   string `json:"last_status_at"`
}

type MediaAttachment struct {
	Type        string `json:"type"`
	URL         string `json:"url"`
	Description string `json:"description"`
	Id          string `json:"id"`
	Path        string
}

type Application struct {
	Name    string `json:"name"`
	Website string `json:"website"`
}

type Tag struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type Post struct {
	CreatedAt          time.Time         `json:"created_at"`
	Id                 string            `json:"id"`
	Visibility         string            `json:"visibility"`
	InReplyToId        string            `json:"in_reply_to_id"`
	InReplyToAccountId string            `json:"in_reply_to_account_id"`
	Sensitive          bool              `json:"sensitive"`
	SpoilerText        string            `json:"spoiler_text"`
	Language           string            `json:"language"`
	URI                string            `json:"uri"`
	URL                string            `json:"url"`
	Application        Application       `json:"application"`
	Content            string            `json:"content"`
	MediaAttachments   []MediaAttachment `json:"media_attachments"`
	RepliesCount       int               `json:"replies_count"`
	ReblogsCount       int               `json:"reblogs_count"`
	FavoritesCount     int               `json:"favourites_count"`
	Pinned             bool              `json:"pinned"`
	Tags               []Tag             `json:"tags"`
	Favourited         bool              `json:"favourited"`
	Reblogged          bool              `json:"reblogged"`
	Muted              bool              `json:"muted"`
	Bookmarked         bool              `json:"bookmarked"`
	Account            Account           `json:"account"`
}

type PostsFilter struct {
	ExcludeReplies bool
	ExcludeReblogs bool
	Limit          int
	SinceId        string
	MinId          string
	MaxId          string
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

func (c Client) Posts(filter PostsFilter) ([]Post, error) {
	var posts []Post
	account, err := c.getAccount()

	if err != nil {
		return posts, err
	}

	queryValues := url.Values{}

	if filter.ExcludeReplies {
		queryValues.Add("exclude_replies", strconv.Itoa(1))
	}

	if filter.ExcludeReblogs {
		queryValues.Add("exclude_reblogs", strconv.Itoa(1))
	}

	if filter.SinceId != "" {
		queryValues.Add("since_id", filter.SinceId)
	}

	if filter.MaxId != "" {
		queryValues.Add("max_id", filter.MaxId)
	}

	if filter.MinId != "" {
		queryValues.Add("min_id", filter.MinId)
	}

	queryValues.Add("limit", strconv.Itoa(filter.Limit))

	query := fmt.Sprintf("?%s", queryValues.Encode())

	postsUrl := fmt.Sprintf(
		"%s/api/v1/accounts/%s/statuses/%s",
		c.baseURL,
		account.Id,
		query,
	)

	log.Println(fmt.Sprintf("Fetching posts from %s", postsUrl))

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

func TagsForPost(post Post, descendants []Post) []Tag {
	var tags []Tag

	for _, tag := range post.Tags {
		tags = append(tags, tag)
	}

	for _, descendant := range descendants {
		for _, tag := range descendant.Tags {
			tags = append(tags, tag)
		}
	}

	return tags
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

func ShouldSkipPost(post Post) bool {
	return post.Visibility != "public"
}
=======
package client

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	handle    string
	baseURL   string
	filters   PostsFilter
	account   Account
	posts     []Post
	replies   map[string]Post
	orphans   []string
	postIdMap map[string]Post
}

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

type MediaAttachment struct {
	Type        string `json:"type"`
	URL         string `json:"url"`
	Description string `json:"description"`
	Id          string `json:"id"`
	Path        string
}

type Application struct {
	Name    string `json:"name"`
	Website string `json:"website"`
}

type Tag struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type Post struct {
	CreatedAt          time.Time         `json:"created_at"`
	Id                 string            `json:"id"`
	Visibility         string            `json:"visibility"`
	InReplyToId        string            `json:"in_reply_to_id"`
	InReplyToAccountId string            `json:"in_reply_to_account_id"`
	Sensitive          bool              `json:"sensitive"`
	SpoilerText        string            `json:"spoiler_text"`
	Language           string            `json:"language"`
	URI                string            `json:"uri"`
	URL                string            `json:"url"`
	Application        Application       `json:"application"`
	Content            string            `json:"content"`
	MediaAttachments   []MediaAttachment `json:"media_attachments"`
	RepliesCount       int               `json:"replies_count"`
	ReblogsCount       int               `json:"reblogs_count"`
	FavoritesCount     int               `json:"favourites_count"`
	Pinned             bool              `json:"pinned"`
	Tags               []Tag             `json:"tags"`
	Favourited         bool              `json:"favourited"`
	Reblogged          bool              `json:"reblogged"`
	Muted              bool              `json:"muted"`
	Bookmarked         bool              `json:"bookmarked"`
	Account            Account           `json:"account"`
	descendants        []Post
}

type PostsFilter struct {
	ExcludeReplies bool
	ExcludeReblogs bool
	Limit          int
	SinceId        string
	MinId          string
	MaxId          string
}

func New(userURL string, filters PostsFilter, threaded bool) (Client, error) {
	var client Client
	parsedURL, err := url.Parse(userURL)

	if err != nil {
		return client, fmt.Errorf("error parsing user url: %w", err)
	}

	baseURL := fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)
	acc := strings.TrimPrefix(parsedURL.Path, "/")
	handle := strings.TrimPrefix(acc, "@")

	account, err := getAccount(baseURL, handle)

	if err != nil {
		return client, err
	}

	posts, err := getPosts(baseURL, account.Id, filters)

	if err != nil {
		return client, err
	}

	var orphans []string
	client = Client{
		baseURL: baseURL,
		handle:  handle,
		filters: filters,
		account: account,
		posts:   posts,
		postIdMap: make(map[string]Post),
		replies: make(map[string]Post),
		orphans: orphans,
	}

	client.populateIdMap()

	if threaded {
		client.generateReplies()
	}

	for _, orphan := range client.orphans {
		log.Println(fmt.Sprintf("Orphan: %s", orphan))
	}

	return client, nil
}

func (c Client) Account() Account {
	return c.account
}

func (c Client) Posts() []Post {
	return c.posts
}

func (p Post) ShouldSkip() bool {
	return p.Visibility != "public"
}

func (p Post) Descendants() []Post {
	return p.descendants
}

func (c Client) populateIdMap() {
	for _, post := range c.posts {
		c.postIdMap[post.Id] = post
	}
}

func (c Client) flushReplies(post Post, descendants *[]Post) {
	if reply, ok := c.replies[post.Id]; ok {
		*descendants = append(*descendants, reply)
		c.flushReplies(reply, descendants)
	}
}

func (c Client) generateReplies() {
	for _, post := range c.posts {
		if post.InReplyToId == "" {
			c.flushReplies(post, &post.descendants)
			continue
		}

		if _, ok := c.postIdMap[post.Id]; ok {
			log.Println(fmt.Sprintf("Adding %s to replies of %s", post.Id, post.InReplyToId))
			c.replies[post.InReplyToId] = post
		} else {
			log.Println("Found orphan")
			c.orphans = append(c.orphans, post.Id)
		}
	}
}

func getPosts(baseURL string, accountId string, filters PostsFilter) ([]Post, error) {
	var posts []Post

	queryValues := url.Values{}

	if filters.ExcludeReplies {
		queryValues.Add("exclude_replies", strconv.Itoa(1))
	}

	if filters.ExcludeReblogs {
		queryValues.Add("exclude_reblogs", strconv.Itoa(1))
	}

	if filters.SinceId != "" {
		queryValues.Add("since_id", filters.SinceId)
	}

	if filters.MaxId != "" {
		queryValues.Add("max_id", filters.MaxId)
	}

	if filters.MinId != "" {
		queryValues.Add("min_id", filters.MinId)
	}

	queryValues.Add("limit", strconv.Itoa(filters.Limit))

	query := fmt.Sprintf("?%s", queryValues.Encode())

	postsUrl := fmt.Sprintf(
		"%s/api/v1/accounts/%s/statuses/%s",
		baseURL,
		accountId,
		query,
	)

	log.Println(fmt.Sprintf("Fetching posts from %s", postsUrl))

	if err := get(postsUrl, &posts); err != nil {
		return posts, err
	}

	return posts, nil
}

func getAccount(baseURL string, handle string) (Account, error) {
	var account Account
	lookupUrl := fmt.Sprintf(
		"%s/api/v1/accounts/lookup?acct=%s",
		baseURL,
		handle,
	)

	err := get(lookupUrl, &account)

	if err != nil {
		return account, err
	}

	return account, nil
}

func (p Post) AllTags() []Tag {
	var tags []Tag

	for _, tag := range p.Tags {
		tags = append(tags, tag)
	}

	for _, descendant := range p.descendants {
		for _, tag := range descendant.Tags {
			tags = append(tags, tag)
		}
	}

	return tags
}

func (p Post) AllMedia() []MediaAttachment {
	var media []MediaAttachment

	for _, item := range p.MediaAttachments {
		media = append(media, item)
	}

	for _, descendant := range p.descendants {
		for _, item := range descendant.MediaAttachments {
			media = append(media, item)
		}
	}

	return media
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
>>>>>>> Stashed changes
