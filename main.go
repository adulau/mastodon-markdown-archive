package main

import (
	"flag"
	"fmt"
	"git.hq.ggpsv.com/gabriel/mastodon-pesos/client"
	"git.hq.ggpsv.com/gabriel/mastodon-pesos/files"
	"log"
)

func main() {
	dist := flag.String("dist", "", "Path to directory where files will be written")
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

	fileWriter, err := files.New(*dist)

	if err != nil {
		log.Panicln(err)
	}

	for _, post := range posts {
		fileWriter.Write(post)
	}
}
