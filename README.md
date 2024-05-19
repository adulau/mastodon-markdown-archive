# Mastodon markdown archive

Fetch a Mastodon account's posts and save them as markdown files. Post content is converted to markdown, images are downloaded and inlined, and replies are threaded. A post whose visibility is not `public` is skipped, and the post's id is used as the filename.

Implements most of the parameters in Mastodon's public [API to get an account's statuses](https://docs.joinmastodon.org/methods/accounts/#statuses).

If a post has images, the post is created as a bundle of files in the manner of Hugo [page bundles](https://gohugo.io/content-management/page-bundles/), and the images are downloaded in the corresponding directory.

I use this tool to create an [archive of my Mastodon posts](https://garrido.io/microblog/), which I then syndicate to my own site following [PESOS](https://indieweb.org/PESOS).

## Install

[Go](https://go.dev/doc/install) is required for installation.

You can clone this repo and run `go build main.go` in the repository's directory, or you can run `go install git.garrido.io/gabriel/mastodon-markdown-archive@latest` to install a binary of the latest version.

## Usage
```
Usage of mastodon-markdown-archive:
  -dist string
        Path to directory where files will be written (default "./posts")
  -exclude-reblogs
        Whether or not to exclude reblogs
  -exclude-replies
        Whether or not exclude replies to other users
  -filename string
        Template for post filename
  -limit int
        Maximum number of posts to fetch (default 40)
  -max-id string
        Fetch posts lesser than this id
  -min-id string
        Fetch posts immediately newer than this id
  -persist-first string
        Location to persist the post id of the first post returned
  -persist-last string
        Location to persist the post id of the last post returned
  -since-id string
        Fetch posts greater than this id
  -template string
        Template to use for post rendering, if passed
  -threaded
        Thread replies for a post in a single file (default true)
  -user string
        URL of User's Mastodon account whose toots will be fetched
```

## Example

Here is how I use this to archive posts from my Mastodon account. 

I use this tool programatically, and I certainly do not want to recreate the archive from scratch each time. I exclude replies to others, and reblogs.

I first use this to generate an archive up to a certain point in time. Then, I use it to archive posts made since the last archived post.

Mastodon imposes an upper limit of 40 posts in their API. With `--persist-first` and `--persist-last` I can save cursors of the upper and lower bound of posts that were fetched. I can then use Mastodon's `max-id`, `min-id`, and `since-id` parameters to get the posts that I need, depending on each cae.

### Generating an entire archive

```sh
mastodon-markdown-archive \
--user=https://social.coop/@ggpsv \
--dist=./posts \
--exclude-replies \
--exclude-reblogs \
--persist-last=./last \
--max-id=$(test -f ./last && cat ./last || echo "")
```

Calling this for the first time will fetch the most recent 40 posts. With `--persist-last`, the 40th post's id will be saved at `./last`.

Calling this command iteratively will fetch the account's posts in reverse chronological time, 40 posts at a time. If my account had 160 posts, I'd need to call this command 4 times to create the archive.

### Getting the latest posts

Calling this for the first time will fetch the most recent 40 posts. With `--persist-first`, the most recent post's id will be saved at `./first`.

Calling this command iteratively will only fetch posts that have been made since the last retrieved post.

```sh
mastodon-markdown-archive \
--user=https://social.coop/@ggpsv \
--dist=./posts \
--exclude-replies \
--exclude-reblogs \
--persist-first=./first \
--since-id=$(test -f ./first && cat ./first || echo "")
```

## Template

By default, this tool uses the [post.tmpl](./files/templates/post.tmpl) template to create the markdown file. A different template can be used by passing its path to `--template`.

For information about variables and functions available in the template context, refer to the `Write` method in [files.go](files/files.go#L95-L101).

