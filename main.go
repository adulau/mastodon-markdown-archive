package main

import (
	"flag"
	"fmt"
	"git.hq.ggpsv.com/gabriel/mastodon-pesos/client"
	"log"
)

func main() {
	user := flag.String("user", "", "URL of User's Mastodon account whose toots will be fetched")

	flag.Parse()

	client, err := client.New(*user)

	if err != nil {
		log.Panicln(fmt.Errorf("error instantiating client: %w", err))
	}

	posts, err := client.GetPosts("?exclude_replies=1&exclude_reblogs=1&limit=10")

	if err != nil {
		log.Panicln(err)
	}

	for _, post := range posts {
		log.Println(post.Id)
	}
}
