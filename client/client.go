package client

import (
	"fmt"
	"net/url"
	"strings"
)

type Client struct {
	handle  string
	baseURL string
	filters PostsFilter
	account Account
	// Map of Post.InReplyToId:PostId. Tracks the replies on a 1:1 basis.
	replies map[string]string
	// List of Post.Id. Tracks posts whose parent is not within the bounds of
	// the returned posts.
	orphans []string
	// Map of Post.Id:*Post.
	postIdMap map[string]*Post
	// List of Post.Id. Tracks the posts which will be written as individual files.
	output []string
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

	postIdMap := make(map[string]*Post)
	var orphans []string
	var output []string

	for i := range posts {
		post := posts[i]
		postIdMap[post.Id] = &post
		if !threaded {
			output = append(output, post.Id)
		}
	}

	client = Client{
		baseURL:   baseURL,
		handle:    handle,
		filters:   filters,
		account:   account,
		postIdMap: postIdMap,
		replies:   make(map[string]string),
		orphans:   orphans,
		output:    output,
	}

	if threaded {
		client.threadReplies(posts)

		if len(client.orphans) > 0 {
			for _, postId := range client.orphans {
				statusContext, err := FetchStatusContext(baseURL, postId)

				if err != nil {
					return client, err
				}

				top := statusContext.Ancestors[0]

				for i := range statusContext.Ancestors[1:] {
					post := statusContext.Ancestors[i+1]
					client.postIdMap[post.Id] = &post
					top.descendants = append(top.descendants, &post)
				}

				top.descendants = append(top.descendants, client.postIdMap[postId])

				for i := range statusContext.Descendants {
					post := statusContext.Descendants[i]
					if post.Account.Id != client.account.Id {
						continue
					}

					client.postIdMap[post.Id] = &post
					top.descendants = append(top.descendants, &post)
				}

				client.postIdMap[top.Id] = &top
				client.output = append(client.output, top.Id)
			}
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
		p = append(p, c.postIdMap[i])
	}

	return p
}

func (c *Client) flushReplies(post *Post, descendants *[]*Post) {
	if pid, ok := c.replies[post.Id]; ok {
		reply := c.postIdMap[pid]
		*descendants = append(*descendants, reply)
		c.flushReplies(reply, descendants)
	}
}

func (c *Client) threadReplies(posts []Post) {
	for i := range posts {
		post := &posts[i]
		if post.InReplyToId == "" {
			c.flushReplies(post, &post.descendants)
			c.output = append(c.output, post.Id)
			continue
		}

		if _, ok := c.postIdMap[post.InReplyToId]; ok {
			c.replies[post.InReplyToId] = post.Id
		} else {
			c.orphans = append(c.orphans, post.Id)
		}
	}
}
