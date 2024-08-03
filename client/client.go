package client

import (
	"fmt"
	"net/url"
	"strings"
)

type ClientOptions struct {
	Visibility string
	Threaded   bool
}

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
	output  []string
	options ClientOptions
}

func New(userURL string, filters PostsFilter, opts ClientOptions) (Client, error) {
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
	replies := make(map[string]string)
	var orphans []string
	var output []string

	for i := range posts {
		post := posts[i]
		postIdMap[post.Id] = &post
		if !opts.Threaded && !post.ShouldSkip(opts.Visibility) {
			output = append(output, post.Id)
		}
	}

	client = Client{
		baseURL:   baseURL,
		handle:    handle,
		filters:   filters,
		account:   account,
		postIdMap: postIdMap,
		replies:   replies,
		orphans:   orphans,
		output:    output,
		options:   opts,
	}

	if opts.Threaded {
		for _, post := range posts {
			client.threadPost(post.Id)
		}

		if len(client.orphans) > 0 {
			if err := client.buildOrphans(); err != nil {
				return client, nil
			}
		}
	}

	return client, nil
}

func (c Client) Account() Account {
	return c.account
}

func (c Client) Posts() []*Post {
	var posts []*Post

	for _, postId := range c.output {
		posts = append(posts, c.postIdMap[postId])
	}

	return posts
}

func (c *Client) buildOrphans() error {
	for _, postId := range c.orphans {
		statusContext, err := FetchStatusContext(c.baseURL, postId)

		if err != nil {
			return err
		}

		var top Post;

		// When building a thread from the status context endpoint, 
		// start from the greatest ancestor and add the other ancestors
		// below it as descendants.
		// Otherwise, use the orphan as the start. 
		if len(statusContext.Ancestors) > 0 {
			top = statusContext.Ancestors[0]

			for i := range statusContext.Ancestors[1:] {
				post := statusContext.Ancestors[i+1]
				if post.Account.Id != c.account.Id {
					continue
				}

				c.postIdMap[post.Id] = &post
				top.descendants = append(top.descendants, &post)
			}

			top.descendants = append(top.descendants, c.postIdMap[postId])
		} else {
			top = *c.postIdMap[postId]
		}

		for i := range statusContext.Descendants {
			post := statusContext.Descendants[i]
			if post.Account.Id != c.account.Id {
				continue
			}

			c.postIdMap[post.Id] = &post
			top.descendants = append(top.descendants, &post)
		}

		c.postIdMap[top.Id] = &top
		c.output = append(c.output, top.Id)
	}

	return nil
}

func (c *Client) flushReplies(post *Post, descendants *[]*Post) {
	if pid, ok := c.replies[post.Id]; ok {
		reply := c.postIdMap[pid]
		*descendants = append(*descendants, reply)
		c.flushReplies(reply, descendants)
	}
}

func (c *Client) threadPost(postId string) {
	post := c.postIdMap[postId]

	if post.InReplyToId == "" && !post.ShouldSkip(c.options.Visibility) {
		c.flushReplies(post, &post.descendants)
		c.output = append(c.output, post.Id)
		return
	}

	if _, ok := c.postIdMap[post.InReplyToId]; ok {
		c.replies[post.InReplyToId] = post.Id
	} else {
		c.orphans = append(c.orphans, post.Id)
	}
}
