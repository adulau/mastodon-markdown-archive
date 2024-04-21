package main

import (
	"flag"
	"fmt"
	"git.hq.ggpsv.com/gabriel/mastodon-pesos/client"
	"git.hq.ggpsv.com/gabriel/mastodon-pesos/files"
	"log"
	"os"
)

func main() {
	dist := flag.String("dist", "", "Path to directory where files will be written")
	user := flag.String("user", "", "URL of User's Mastodon account whose toots will be fetched")
	excludeReplies := flag.Bool("exclude-replies", false, "Whether or not exclude replies to other users")
	excludeReblogs := flag.Bool("exclude-reblogs", false, "Whether or not to exclude reblogs")
	limit := flag.Int("limit", 40, "Maximum number of posts to fetch")
	sinceId := flag.String("since-id", "", "Fetch only posts made since passed post id")
	persist := flag.Bool("persist", false, "Persist most recent post id to /tmp/mastodon-pesos-fid")

	flag.Parse()

	c, err := client.New(*user)

	if err != nil {
		log.Panicln(fmt.Errorf("error instantiating client: %w", err))
	}

	posts, err := c.GetPosts(client.PostsFilter{
		ExcludeReplies: *excludeReplies,
		ExcludeReblogs: *excludeReblogs,
		Limit:          *limit,
		SinceId:        *sinceId,
	})

	if err != nil {
		log.Panicln(err)
	}

	fileWriter, err := files.New(*dist)

	if err != nil {
		log.Panicln(err)
	}

	log.Println(fmt.Sprintf("Fetched %d posts", len(posts)))

	for _, post := range posts {
		if err := fileWriter.Write(post); err != nil {
			log.Panicln("error writing post to file: %w", err)
			break
		}
	}

	if *persist && len(posts) > 0 {
		lastPost := posts[0]

		fid := []byte(lastPost.Id)
		os.WriteFile("/tmp/mastodon-pesos-fid", fid, 0644)
	}
}
