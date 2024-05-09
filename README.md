# Mastodon markdown archive

Fetch a Mastodon account's posts and save them as markdown files. Post content is converted to markdown, images are downloaded and inlined, and replies are threaded. Implements most of the parameters in Mastodon's public [API to get an account's statuses](https://docs.joinmastodon.org/methods/accounts/#statuses).

For the time being this formats the files in accordance to [Hugo's](https://gohugo.io) front-matter. 

If a post has images, the post is created as a Hugo [page bundle](https://gohugo.io/content-management/page-bundles/) and images are downloaded in the corresponding post directory.

I use this tool to create an [archive of my Mastodon posts](https://garrido.io/microblog/), which I then syndicate to my own site following [PESOS](https://indieweb.org/PESOS).

## Flags
```
Usage of ./mastodon-pesos:
  -dist string
        Path to directory where files will be written (default "./posts")
  -exclude-reblogs
        Whether or not to exclude reblogs
  -exclude-replies
        Whether or not exclude replies to other users
  -limit int
        Maximum number of posts to fetch (default 40)
  -max-id string
        Fetch posts lesser than this id
  -min-id string
        Fetch posts immediately newer than this id
  -persist
        Persist most recent post id to /tmp/mastodon-pesos-fid
  -since-id string
        Fetch posts greater than this id
  -user string
        URL of User's Mastodon account whose toots will be fetched
```

## Example

Here is how I use this to fetch the 15 most recent posts in my Mastodon account. It excludes replies to others, and reblogs. 

Lastly, I use `--persist` to save the most recent id to a file and use `--since-id` so that subsequent runs fetch posts only after the most recently fetched post.

```sh
./mastodon-pesos \
--user https://social.coop/@ggpsv \
--dist ./posts \
--exclude-replies \
--exclude-reblogs \
--limit=15 \
--persist \
--since-id=$(test -f /tmp/mastodon-pesos-fid && cat /tmp/mastodon-pesos-fid || echo "")
```
