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
	excludeReplies := flag.Bool("exclude-replies", false, "Exclude replies to other users")
	excludeReblogs := flag.Bool("exclude-reblogs", false, "Exclude reblogs")
	limit := flag.Int("limit", 40, "Maximum number of posts to fetch")
	sinceId := flag.String("since-id", "", "Fetch posts greater than this id")
	maxId := flag.String("max-id", "", "Fetch posts older than this id")
	minId := flag.String("min-id", "", "Fetch posts newer than this id")
	persistFirst := flag.String("persist-first", "", "Location to persist the post id of the first post returned")
	persistLast := flag.String("persist-last", "", "Location to persist the post id of the last post returned")
	templateFile := flag.String("template", "", "Template to use for post rendering, if passed")
	threaded := flag.Bool("threaded", false, "Thread replies for a post in a single file")
	filenameTemplate := flag.String("filename", "", "Template for post filename")
	porcelain := flag.Bool("porcelain", false, "Prints the amount of fetched posts to stdout in a parsable manner")
	downloadMedia := flag.String("download-media", "", "Download media in a post. Omit or pass an empty string to not download media. Pass 'bundle' to download the media inline in a single directory with its original post. Pass a path to a directory to download all media there.")
	visibility := flag.String("visibility", "", "Filter out posts whose visibility does not match the passed visibility value")

	flag.Parse()

	c, err := client.New(*user, client.PostsFilter{
		ExcludeReplies: *excludeReplies,
		ExcludeReblogs: *excludeReblogs,
		Limit:          *limit,
		SinceId:        *sinceId,
		MaxId:          *maxId,
		MinId:          *minId,
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
