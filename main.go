package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"git.garrido.io/gabriel/mastodon-markdown-archive/client"
	"git.garrido.io/gabriel/mastodon-markdown-archive/files"
)

func main() {
	dist := flag.String("dist", "./posts", "Path to directory where files will be written")
	user := flag.String("user", "", "URL of User's Mastodon account whose toots will be fetched")
	excludeReplies := flag.Bool("exclude-replies", false, "Whether or not exclude replies to other users")
	excludeReblogs := flag.Bool("exclude-reblogs", false, "Whether or not to exclude reblogs")
	limit := flag.Int("limit", 40, "Maximum number of posts to fetch")
	sinceId := flag.String("since-id", "", "Fetch posts greater than this id")
	maxId := flag.String("max-id", "", "Fetch posts lesser than this id")
	minId := flag.String("min-id", "", "Fetch posts immediately newer than this id")
	persistFirst := flag.String("persist-first", "", "Location to persist the post id of the first post returned")
	persistLast := flag.String("persist-last", "", "Location to persist the post id of the last post returned")
	templateFile := flag.String("template", "", "Template to use for post rendering, if passed")
	threaded := flag.Bool("threaded", true, "Thread replies for a post in a single file")

	flag.Parse()

	c, err := client.New(*user, client.PostsFilter{
		ExcludeReplies: *excludeReplies,
		ExcludeReblogs: *excludeReblogs,
		Limit:          *limit,
		SinceId:        *sinceId,
		MaxId:          *maxId,
		MinId:          *minId,
	}, *threaded)

	if err != nil {
		log.Panicln(err)
	}

	fileWriter, err := files.New(*dist)
	posts := c.Posts()
	postsCount := len(posts)

	log.Println(fmt.Sprintf("Fetched %d posts", postsCount))

	for _, post := range posts {
		if post.ShouldSkip() {
			continue
		}

		if err := fileWriter.Write(post, *templateFile); err != nil {
			log.Panicln("error writing post to file: %w", err)
			break
		}
	}

	if postsCount > 0 {
		if *persistFirst != "" {
			firstPost := posts[0]
			err := persistId(firstPost.Id, *persistFirst)

			if err != nil {
				log.Panicln(err)
			}
		}

		if *persistLast != "" {
			lastPost := posts[postsCount-1]
			err := persistId(lastPost.Id, *persistLast)

			if err != nil {
				log.Panicln(err)
			}
		}
	}
}

func persistId(postId string, path string) error {
	persistPath, err := filepath.Abs(path)

	if err != nil {
		return err
	}

	if err := os.WriteFile(persistPath, []byte(postId), 0644); err != nil {
		return err
	}

	return nil
}
