# Mastodon PESOS

Fetch a Mastodon account's posts and save them as Markdown files. Posts are transformed to Markdown, images are inlined, and replies are threaded.

For the time being this formats the files with [Hugo](https://gohugo.io) front-matter.

I use this small tool to create an archive of my Mastodon posts, which I then [syndicate to my own site](https://indieweb.org/PESOS).

## Flags
```
Usage of ./mastodon-pesos:
  -dist string
        Path to directory where files will be written
  -exclude-reblogs
        Whether or not to exclude reblogs
  -exclude-replies
        Whether or not exclude replies to other users
  -limit int
        Maximum number of posts to fetch (default 40)
  -persist
        Persist most recent post id to /tmp/mastodon-pesos-fid
  -since-id string
        Fetch only posts made since passed post id
  -user string
        URL of User's Mastodon account whose toots will be fetched
```

## Usage

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
