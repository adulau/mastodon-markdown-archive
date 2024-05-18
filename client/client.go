package client

import (
	"fmt"
	"net/url"
	"strings"
)

type Client struct {
	handle    string
	baseURL   string
	filters   PostsFilter
	account   Account
	posts     []Post
	replies   map[string]string
	orphans   []string
	postIdMap map[string]int
	output    []int
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

	account, err := FetchAccount(baseURL, handle)

	if err != nil {
		return client, err
	}

	posts, err := FetchPosts(baseURL, account.Id, filters)

	if err != nil {
		return client, err
	}

	var orphans []string
	client = Client{
		baseURL:   baseURL,
		handle:    handle,
		filters:   filters,
		account:   account,
		posts:     posts,
		postIdMap: make(map[string]int),
		replies:   make(map[string]string),
		orphans:   orphans,
	}

	client.populateIdMap()

	if threaded {
		client.threadReplies()

		if len(client.orphans) > 0 {
			for _, pid := range client.orphans {
				statusContext, err := FetchStatusContext(baseURL, pid)

				if err != nil {
					return client, err
				}

				top := statusContext.Ancestors[0]

				for _, post := range statusContext.Ancestors[1:] {
					client.posts = append(client.posts, post)
					top.descendants = append(top.descendants, &client.posts[len(client.posts)-1])
				}

				top.descendants = append(top.descendants, &client.posts[client.postIdMap[pid]])

				for _, post := range statusContext.Descendants {
					if post.Account.Id != client.account.Id {
						continue
					}

					client.posts = append(client.posts, post)
					top.descendants = append(top.descendants, &client.posts[len(client.posts)-1])
				}

				client.posts = append(client.posts, top)
				client.output = append(client.output, len(client.posts)-1)
			}
		}
	} else {
		for i := range client.posts {
			client.output = append(client.output, i)
		}
	}

	return client, nil
}

func (c Client) Account() Account {
	return c.account
}

func (c Client) Posts() []*Post {
	var p []*Post

	for _, i := range c.output {
		p = append(p, &c.posts[i])
	}

	return p
}

func (c *Client) populateIdMap() {
	for i, post := range c.posts {
		c.postIdMap[post.Id] = i
	}
}

func (c *Client) flushReplies(post *Post, descendants *[]*Post) {
	if pid, ok := c.replies[post.Id]; ok {
		reply := c.posts[c.postIdMap[pid]]
		*descendants = append(*descendants, &reply)
		c.flushReplies(&reply, descendants)
	}
}

func (c *Client) threadReplies() {
	for i := range c.posts {
		post := &c.posts[i]
		if post.InReplyToId == "" {
			c.flushReplies(post, &post.descendants)
			c.output = append(c.output, i)
			continue
		}

		if _, ok := c.postIdMap[post.InReplyToId]; ok {
			c.replies[post.InReplyToId] = post.Id
		} else {
			c.orphans = append(c.orphans, post.Id)
		}
	}
}
