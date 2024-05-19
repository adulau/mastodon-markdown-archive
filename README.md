# Mastodon markdown archive

Fetch a Mastodon account's posts and save them as markdown files. Post content is converted to markdown, images are downloaded and inlined, and replies are threaded. A post whose visibility is not `public` is skipped, and the post's id is used as the filename.

Implements most of the parameters in Mastodon's public [API to get an account's statuses](https://docs.joinmastodon.org/methods/accounts/#statuses).

I use this tool to create an [archive of my Mastodon posts](https://garrido.io/microblog/), which I then syndicate to my own site following [PESOS](https://indieweb.org/PESOS).

## Install

[Go](https://go.dev/doc/install) is required for installation.

You can clone this repo and run `go build main.go` in the repository's directory, or you can run `go install git.garrido.io/gabriel/mastodon-markdown-archive@latest` to install a binary of the latest version.

## Usage
```
Usage of mastodon-markdown-archive:
  -dist string
        Path to directory where files will be written (default "./posts")
  -download-media string
        Download media in a post. Omit or pass an empty string to not download media. Pass 'bundle' to download the media inline in a single directory with its original post. Pass a path to a directory to download all media there.
  -exclude-reblogs
        Exclude reblogs
  -exclude-replies
        Exclude replies to other users
  -filename string
        Template for post filename
  -limit int
        Maximum number of posts to fetch (default 40)
  -max-id string
        Fetch posts older than this id
  -min-id string
        Fetch posts newer than this id
  -persist-first string
        Location to persist the post id of the first post returned
  -persist-last string
        Location to persist the post id of the last post returned
  -porcelain
        Prints the amount of fetched posts to stdout in a parsable manner
  -since-id string
        Fetch posts greater than this id
  -template string
        Template to use for post rendering, if passed
  -threaded
        Thread replies for a post in a single file
  -user string
        URL of Mastodon account whose toots will be fetched
  -visibility string
        Filter out posts whose visibility does not match the passed visibility value
```

## Example

Here is how I use this to archive posts from my Mastodon account. 

I use this tool programatically, and I do not want to recreate the archive from scratch each time. I exclude replies to others, and reblogs.

I first used this to generate an archive of all the posts that I had published to date. Then, I run it programatically to archive any new posts made.

Mastodon imposes a maximum limit of 40 posts in this API. With `--persist-first` and `--persist-last` I can save cursors of the upper and lower bound of posts that were fetched. I then use the API's `max-id`, `min-id`, and `since-id` parameters to get the posts that I need, depending on each case.

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

Calling this for the first time will fetch the most recent 40 posts. With `--persist-last./last`, the oldest fetched post id will be saved at `./last`.

Calling this command iteratively will fetch the account's posts in reverse chronological order, 40 posts at a time. 

You can use simple bash script to automate this process. Adding the `--porcelain` flag prints the amount of fetched posts to stdout, which can then be used continue or stop fetching posts:

```bash
#!/bin/bash

while true; do
  command="mastodon-markdown-archive --dist=./example \
    --exclude-replies=true \
    --exclude-reblogs=true \
    --user=https://social.coop/@ggpsv \
    --porcelain=true \
    --threaded=true \
    --persist-last=./last \
    --max-id=$(test -f ./last && cat ./last || echo '')"
  output=$($command)

  if [[ "$output" -eq 0 ]]; then
    echo "No posts returned. Exiting"
    break
  fi
  echo "Fetched $output posts. Continuing."
  sleep 1
done
```

### Getting the latest posts

With `--persist-first=./first`, the most recent post id will be saved at `./first`.

Calling this command iteratively will only fetch posts that have been made since the last retrieved post:

```sh
mastodon-markdown-archive \
--user=https://social.coop/@ggpsv \
--dist=./posts \
--exclude-replies \
--exclude-reblogs \
--persist-first=./first \
--since-id=$(test -f ./first && cat ./first || echo "")
```

## Threading 

By default, posts by the author in reply to another post by the author will be written out as separate files.

However, posts can be threaded together using the `--threaded=true` flag. With threading, the descendants of a post will not be written out as a separate files. Instead, only the top post will be written out. 

The program will aggregate the post's descendants in reverse chronological order and make them available in the template via the [Descendants](https://pkg.go.dev/git.garrido.io/gabriel/mastodon-markdown-archive/client#Post.Descendants) method. This can be used in [templates](#templating) to render threaded posts as a single post, which the default template does.

When threading, the `AllMedia` and `AllTags` methods will yield the aggregated [MediaAttachment](https://pkg.go.dev/git.garrido.io/gabriel/mastodon-markdown-archive/client#MediaAttachment) and [Tag](https://pkg.go.dev/git.garrido.io/gabriel/mastodon-markdown-archive/client#Tag), respectively.

### Orphaned posts

Mastodon limits their statuses API to a maximum 40 posts at a time, and the `--limit` flag can be used to limit this further.

Because of this limit, it is possible that posts in a thread end up split across different responses. Or, a user may maintain a long-lived thread of posts that gets updated sporadically. This results in an orphaned post, which is a post whose parent is not within the same batch of posts returned by a single API call.

In either case, the program will fallback to using the [status context](https://docs.joinmastodon.org/methods/statuses/#context) endpoint to rebuild the corresponding thread from the top.

## Templating

The contents of the file and the filename for each post can be customized using templates. This provides enough flexibility to use this tool for various purposes. The templates are evaluated as Go [text templates](https://pkg.go.dev/text/template), so it should be possible to do anything that's supported in a Go template.

For example, if you're using this to syndicate posts to a site built using a static site generator, you can customize the output so that it adheres to specific requirements around front-matter structure or filename formats.

### Post
Out of the box, this tool uses the [post.tmpl](./files/templates/post.tmpl) template to create the post file. It converts the post content to markdown,  threads replies, and defines some attributes in the front-matter using YAML. 

For example, for this [post](https://social.coop/@ggpsv/112326240503555949):

```md
---
date: 2024-04-24 12:40:10.029 +0000 UTC
post_uri: https://social.coop/users/ggpsv/statuses/112326240503555949
post_id: 112326240503555949
tags:
- FrameworkLaptop
- fedora
---
Back at dual-booting on the [#FrameworkLaptop](https://social.coop/tags/FrameworkLaptop). Last time it was Ubuntu, but now I have gone with [#Fedora](https://social.coop/tags/Fedora) 40 KDE.

I'm impressed with how things just work with this laptop. Major props to the [@frameworkcomputer](https://fosstodon.org/@frameworkcomputer) team for supporting these distros out of the box.

I simply decrypted my drive, shrunk it, created a partition, booted off a USB key, installed Fedora, encrypted both partitions, and that's it.

Also, KDE Plasma 6 looks incredibly crisp on this screen.
```

A different template can be used by passing its path to `--template`. The template must comply with Go template syntax.

For example, a `jekyll.tmpl` template with customized front-matter :

```
---
layout: post
title: {{ substr 0 5 .Post.Id }}
published: true
---

{{ .Post.Content | toMarkdown }}
```

Passed to the command as `--template=./jekyll.tmpl` will yield a file that looks like this:

```md
---
layout: post
title: 11232
published: true
---

Back at dual-booting on the [#FrameworkLaptop](https://social.coop/tags/FrameworkLaptop). Last time it was Ubuntu, but now I have gone with [#Fedora](https://social.coop/tags/Fedora) 40 KDE.

I'm impressed with how things just work with this laptop. Major props to the [@frameworkcomputer](https://fosstodon.org/@frameworkcomputer) team for supporting these distros out of the box.

I simply decrypted my drive, shrunk it, created a partition, booted off a USB key, installed Fedora, encrypted both partitions, and that's it.

Also, KDE Plasma 6 looks incredibly crisp on this screen.
```

You might even want to use HTML as the output and thus have a `html.tmpl` file:

```html
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>{{ .Post.Id }}</title>
</head>
<body>
  {{.Post.Content}}
</body>
</html>
```

Passed to the command as `--template=./html.tmpl` will yield a file that looks like this:
```html
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>112326240503555949</title>
</head>
<body>
  <p>Back at dual-booting on the <a href="https://social.coop/tags/FrameworkLaptop" class="mention hashtag" rel="tag">#<span>FrameworkLaptop</span></a>. Last time it was Ubuntu, but now I have gone with <a href="https://social.coop/tags/Fedora" class="mention hashtag" rel="tag">#<span>Fedora</span></a> 40 KDE.</p><p>I&#39;m impressed with how things just work with this laptop. Major props to the <span class="h-card" translate="no"><a href="https://fosstodon.org/@frameworkcomputer" class="u-url mention">@<span>frameworkcomputer</span></a></span> team for supporting these distros out of the box.</p><p>I simply decrypted my drive, shrunk it, created a partition, booted off a USB key, installed Fedora, encrypted both partitions, and that&#39;s it. </p><p>Also, KDE Plasma 6 looks incredibly crisp on this screen.</p>
</body>
</html>
```

### Filename
Out of the box, this tool uses `<post id>.md` as the post filename format. For example, this [post](https://social.coop/@ggpsv/112326240503555949) is saved `112326240503555949.md`

A different format for the filename can be used by passing a template string to `--filename`. The string must comply with Go template syntax.

For example, to create post files that are prefixed with the post's creation date in `YYYY-MM-DD` format and suffixed with the post id, pass `--filename='{{.Post.CreatedAt | date "2006-01-02"}}-{{.Post.Id}}.md`.

An extension suffixed to the filename template will be used if present. Otherwise, `.md` is used as the default file extension. 

Following the HTML example in the [post template section](#post) above, you format the filename as `--filename='{{.Post.Id}}.html'` to use HTML as the extension.

### Available functions and variables

For both the post and filename templates, the following functions and variables are available:

#### Functions
* Standard Go template functions
* All [Sprig](https://masterminds.github.io/sprig/) functions
* `toMarkdown` to convert the post's HTML content to Markdown, without escaping any markdown syntax
* `toMarkdownEscaped` to convert the post's HTML content to Markdown, escaping any markdown syntax

#### Variables
* [Post](https://pkg.go.dev/git.garrido.io/gabriel/mastodon-markdown-archive/client#Post)

## Post media

By default, a post's media is not downloaded. Use the `--download-media` flag with a path to download a post's media. The post's original file is downloaded, and the image's id is used as the filename. 

For example, `--download-media=./images` saves any media to the `./images`.

Once downloaded, the media's path is available in [MediaAttachment.Path](https://pkg.go.dev/git.garrido.io/gabriel/mastodon-markdown-archive/client#MediaAttachment) as an absolute path.

Sprig's [path](https://masterminds.github.io/sprig/paths.html) functions can be used in the templates to manipulate the path as necessary. For example, the default template uses `osBase` to get the last element of the filepath.  

You can use `--download-media=bundle` to save the post media in a single directory with its original post. In this case, the post's filename will be used as the directory name and the post filename will be `index.{extension}`. This is done specifically to support Hugo [page bundles](https://gohugo.io/content-management/page-bundles).

For example, `--download-media="./bundle" --filename='{{ .Post.CreatedAt | date "2006-01-02" }}-{{.Post.Id}}.md'` will create a `YYYY-MM-DD-<post id>/` directory, with the post saved as `YYYY-MM-DD-<post id>/index.md` and media saved as `YYYY-MM-DD-<post id>/<media id>.<media ext>`.