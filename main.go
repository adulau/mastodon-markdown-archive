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
	user := flag.String("user", "", "URL of Mastodon account whose toots will be fetched")
	excludeReplies := flag.Bool("exclude-replies", false, "Mastodon API parameter: Filter out statuses in reply to a different account")
	excludeReblogs := flag.Bool("exclude-reblogs", false, "Mastodon API parameter: Filter out boosts from the response")
	limit := flag.Int("limit", 40, "Mastodon API parameter: Maximum number of results to return. Defaults to 20 statuses. Max 40 statuses")
	onlyMedia := flag.Bool("only-media", false, "Mastodon API parameter: Filter out status without attachments")
	pinned := flag.Bool("pinned", false, "Mastodon API parameter: Filter for pinned statuses only")
	sinceId := flag.String("since-id", "", "Mastodon API parameter: All results returned will be greater than this ID. In effect, sets a lower bound on results.")
	maxId := flag.String("max-id", "", "Mastodon API parameter: All results returned will be lesser than this ID. In effect, sets an upper bound on results.")
	minId := flag.String("min-id", "", "Mastodon API parameter: Returns results immediately newer than this ID. In effect, sets a cursor at this ID and paginates forward.")
	tagged := flag.String("tagged", "", "Mastodon API parameter: Filter for statuses using a specific hashtag")
	persistFirst := flag.String("persist-first", "", "Location to persist the post id of the first post returned")
	persistLast := flag.String("persist-last", "", "Location to persist the post id of the last post returned")
	templateFile := flag.String("template", "", "Template to use for post rendering, if passed")
	threaded := flag.Bool("threaded", false, "Thread replies for a post in a single file")
	filenameTemplate := flag.String("filename", "", "Template for post filename")
	porcelain := flag.Bool("porcelain", false, "Prints the amount of fetched posts to stdout in a parsable manner")
	downloadMedia := flag.String("download-media", "", "Path where post attachments will be downloaded. Omit to skip downloading attachments.")
	visibility := flag.String("visibility", "", "Filter out posts whose visibility does not match the passed visibility value")

	flag.Parse()

	c, err := client.New(*user, client.PostsFilter{
		ExcludeReplies: *excludeReplies,
		ExcludeReblogs: *excludeReblogs,
		Limit:          *limit,
		SinceId:        *sinceId,
		MaxId:          *maxId,
		MinId:          *minId,
		OnlyMedia:      *onlyMedia,
		Pinned:         *pinned,
		Tagged:         *tagged,
	}, client.ClientOptions{
		Threaded:   *threaded,
		Visibility: *visibility,
	})

	if err != nil {
		log.Panicln(err)
	}

	fileWriter, err := files.New(*dist, *templateFile, *filenameTemplate, *downloadMedia)
	posts := c.Posts()
	postsCount := len(posts)

	if *porcelain {
		fmt.Println(postsCount)
	} else {
		log.Println(fmt.Sprintf("Fetched %d posts", postsCount))
	}

	if postsCount == 0 {
		return
	}

	for _, post := range posts {
		if err := fileWriter.Write(post); err != nil {
			log.Panicln("error writing post to file: %w", err)
		}
	}

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
