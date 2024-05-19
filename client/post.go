package client

import (
	"fmt"
	"net/url"
	"strconv"
	"time"
)

type StatusContext struct {
	Ancestors   []Post `json:"ancestors"`
	Descendants []Post `json:"descendants"`
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
	descendants        []*Post
}

type PostsFilter struct {
	ExcludeReplies bool
	ExcludeReblogs bool
	Limit          int
	SinceId        string
	MinId          string
	MaxId          string
	OnlyMedia      bool
	Pinned         bool
	Tagged         string
}

func (p Post) ShouldSkip(visibility string) bool {
	if visibility == "" {
		return false
	}

	return p.Visibility != visibility
}

func (p Post) Descendants() []*Post {
	return p.descendants
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

func FetchPosts(baseURL string, accountId string, filters PostsFilter) ([]Post, error) {
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

	if filters.Tagged != "" {
		queryValues.Add("tagged", filters.Tagged)
	}

	if filters.OnlyMedia {
		queryValues.Add("only_media", strconv.Itoa(1))
	}

	if filters.Pinned {
		queryValues.Add("pinned", strconv.Itoa(1))
	}

	queryValues.Add("limit", strconv.Itoa(filters.Limit))

	query := fmt.Sprintf("?%s", queryValues.Encode())

	postsUrl := fmt.Sprintf(
		"%s/api/v1/accounts/%s/statuses/%s",
		baseURL,
		accountId,
		query,
	)

	if err := Fetch(postsUrl, &posts); err != nil {
		return posts, err
	}

	return posts, nil
}

func FetchStatusContext(baseURL, postId string) (StatusContext, error) {
	var status StatusContext
	statusUrl := fmt.Sprintf(
		"%s/api/v1/statuses/%s/context",
		baseURL,
		postId,
	)

	if err := Fetch(statusUrl, &status); err != nil {
		return status, err
	}

	return status, nil
}
