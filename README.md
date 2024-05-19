# Mastodon markdown archive

Fetch a Mastodon account's posts and save them as text files using Mastodon's [statuses API](https://docs.joinmastodon.org/methods/accounts/#statuses).

This program essentially wraps the Mastodon API with a command line interface with some additional features.

**Features**
- Supports all parameters in Mastodon's statuses API
- Convert post to markdown
- Customize output file location, name, and extension
- Customize output format and front matter
- Optionally download of post media
- Optionally threading of posts
- Optionally filter based on post visibility
- Optional affordances for scripting
- Optionally persist fetched post id cursors
- Optionally set authorization token to fetch private posts


I use this tool to create an archive of my Mastodon posts and [syndicate them to my own site](https://garrido.io/microblog/), per IndieWeb's [PESOS philosophy](https://indieweb.org/PESOS).

## Table of contents
* [Installation](Installation)
* [Usage](#usage)
  * [Environment variables](#environment-variables)
* [Examples](#examples)
  * [Generating an entire archive](#generating-an-entire-archive)
  * [Getting the latest posts](#getting-the-latest-posts)
* [Threading](#threading)
  * [Orphaned posts](#orphaned-posts)
* [Templating](#templating)
  * [Post](#post)
  * [Filename](#filename)
  * [Available functions and variables](#available-functions-and-variables)
    * [Functions](#functions)
    * [Variables](#variables)
  * [Template examples](#template-examples)
    * [Jekyll](#jekyll)
    * [Hugo and 11ty](#hugo-and-11ty)
    * [HTML](#html)
    * [Text only](#text-only)
* [Post media](#post-media)
  * [Bundling](#bundling)

## Installation

[Go](https://go.dev/doc/install) is required for installation.

You can clone this repo and run `go build main.go` in the repository's directory, or you can run `go install git.garrido.io/gabriel/mastodon-markdown-archive@latest` to install a binary of the latest version.

## Usage
```
Usage of mastodon-markdown-archive:
  -dist string
        Path to directory where files will be written (default "./posts")
  -download-media string
        Path where post attachments will be downloaded. Omit to skip downloading attachments.
  -exclude-reblogs
        Mastodon API parameter: Filter out boosts from the response
  -exclude-replies
        Mastodon API parameter: Filter out statuses in reply to a different account
  -filename string
        Template for post filename
  -limit int
        Mastodon API parameter: Maximum number of results to return. Defaults to 20 statuses. Max 40 statuses (default 40)
  -max-id string
        Mastodon API parameter: All results returned will be lesser than this ID. In effect, sets an upper bound on results.
  -min-id string
        Mastodon API parameter: Returns results immediately newer than this ID. In effect, sets a cursor at this ID and paginates forward.
  -only-media
        Mastodon API parameter: Filter out status without attachments
  -persist-first string
        Location to persist the post id of the first post returned
  -persist-last string
        Location to persist the post id of the last post returned
  -pinned
        Mastodon API parameter: Filter for pinned statuses only
  -porcelain
        Prints the amount of fetched posts to stdout in a parsable manner
  -since-id string
        Mastodon API parameter: All results returned will be greater than this ID. In effect, sets a lower bound on results.
  -tagged string
        Mastodon API parameter: Filter for statuses using a specific hashtag
  -template string
        Template to use for post rendering, if passed
  -threaded
        Thread replies for a post in a single file
  -user string
        URL of Mastodon account whose toots will be fetched
  -visibility string
        Filter out posts whose visibility does not match the passed visibility value
```

The only required flags for this program to work is `dist` and `user`. All other flags are there for Mastodon's API parameters, or to support more complex use cases. See the [examples](#examples) section. 

### Environment variables

If the `MASTODON_AUTH_TOKEN` environment variable is set then this program will set the `Authorization` header for the statuses and statuses context API requests. This token only needs the `read:statuses` permission.

In the context of the statuses request, this allows you to fetch private statuses that only you can normally see. For example, if a post's visibility is set to "Followers only".

In the context of the status context request for [orphaned posts](#orphaned-posts), this allows you to fetch private statuses and surpass the limited amount of ancestors and descendants.

## Examples

I use this tool programatically, and I do not want to recreate the archive from scratch each time. I thread posts, exclude replies to others, exclude reblogs, and filter out any post that is not public.

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
--visibility=public \
--download-media=bundle \
--threaded=true \
--max-id=$(test -f ./last && cat ./last || echo "")
```

Calling this for the first time will fetch the most recent 40 posts. With `--persist-last./last`, the oldest fetched post id will be saved at `./last`. Caling this command again will set the `last` cursor to the oldest post of the next 40 posts, and so on.

You can use a simple bash script to automate this process. Adding the `--porcelain` flag prints the amount of fetched posts to stdout, which can then be used to continue or stop fetching posts:

```bash
#!/bin/bash

while true; do
  command="mastodon-markdown-archive --dist=./example \
    --exclude-replies=true \
    --exclude-reblogs=true \
    --user=https://social.coop/@ggpsv \
    --porcelain=true \
    --visibility=public \
    --download-media=bundle \
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

Having created the entire archive, I now want to run this on a schedule to retrieve only the latest posts.

With `--persist-first=./first`, the most recent post id will be saved at `./first`.

Calling this command iteratively will only fetch posts that have been made since then.

```sh
mastodon-markdown-archive \
--user=https://social.coop/@ggpsv \
--dist=./posts \
--exclude-replies=true \
--exclude-reblogs=true \
--visibility=public \
--download-media=bundle \
--threaded=true \
--persist-first=./first \
--since-id=$(test -f ./first && cat ./first || echo "")
```

## Threading 

By default, posts by the author in reply to another post by the author will be written out as separate files.

Alternatively, posts can be threaded together using the `--threaded=true` flag. With threading, the descendants of a post will not be written out as a separate files. Instead, only the top post will be written out. 

The program will aggregate the post's descendants in reverse chronological order and make them available in the template via the [Descendants](https://pkg.go.dev/git.garrido.io/gabriel/mastodon-markdown-archive/client#Post.Descendants) method. This can be used in [templates](#templating) to render threaded posts as a single post, which the [default template does](./files/templates/post.tmpl#L33).

When threading, the `AllMedia` and `AllTags` methods will yield the aggregated [MediaAttachment](https://pkg.go.dev/git.garrido.io/gabriel/mastodon-markdown-archive/client#MediaAttachment) and [Tag](https://pkg.go.dev/git.garrido.io/gabriel/mastodon-markdown-archive/client#Tag), respectively.

When the `--visibility` flag is used, only the top post's visibility is evaluated. This is done explicitly to support the common practice in Mastodon of setting threaded replies as `unlisted`.

### Orphaned posts

Mastodon limits their statuses API to a maximum 40 posts at a time, and the `--limit` flag can be used to limit this further.

Because of this limit, it is possible that posts in a thread end up split across different responses. Or, a user may maintain a long-lived thread of posts that gets updated sporadically and thus rarely will a single batch of posts have all the descendants of the post. 

An orphaned post is a post whose parent is not within a batch of posts returned by a single API call.

In either case, the program will fallback to using the [status context](https://docs.joinmastodon.org/methods/statuses/#context) endpoint to rebuild the corresponding thread from the top.

## Templating

The contents of the file and the filename for each post can be customized using templates. This provides enough flexibility to use this tool for various purposes. The templates are evaluated as Go [text templates](https://pkg.go.dev/text/template), so it should be possible to do anything that's normally supported in a Go template.

For example, if you're using this to syndicate posts to a site built using a static site generator, you can customize the output so that it adheres to specific requirements around front matter structure or filename formats.

### Post
Out of the box, this tool uses the [post.tmpl](./files/templates/post.tmpl) template to create the post file. It converts the post content to markdown,  threads replies, and defines some attributes in the front matter using YAML. 

For example, this [post](https://social.coop/@ggpsv/112326240503555949) is converted to this markdown file:

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

For example, a `jekyll.tmpl` template with customized front matter :

```
---
layout: post
title: {{ substr 0 5 .Post.Id }}
published: true
---

{{ .Post.Content | toMarkdown }}
```

Passed to the command as `--template=./jekyll.tmpl` will instead yield a file that looks like this:

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

You might even want to use HTML as the output and thus pass a `--template=./html.tmpl` flag for a `html.tmpl` template that looks like this:

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

### Filename
Out of the box, this tool uses the post's id and the `.md` extension for the filename. For example, this [post](https://social.coop/@ggpsv/112326240503555949) is saved `112326240503555949.md`

A different format for the filename can be used by passing a template string to `--filename`. The string must comply with Go template syntax.

For example, to create post files that are prefixed with the post's creation date in `YYYY-MM-DD` format and suffixed with the post id, pass `--filename='{{.Post.CreatedAt | date "2006-01-02"}}-{{.Post.Id}}.md`.

An extension in the filename template will be used if present. Otherwise, `.md` is used as the default file extension. 

Following the HTML example in the [post template section](#post) above, you may customize the filename as `--filename='{{.Post.Id}}.html'` to use HTML as the output file extension.

### Available functions and variables

For both the post and filename templates, the following functions and variables are available:

#### Functions
* Standard Go template functions
* All [Sprig](https://masterminds.github.io/sprig/) functions
* `toMarkdown` to convert the post's HTML content to Markdown, without escaping any markdown syntax
* `toMarkdownEscaped` to convert the post's HTML content to Markdown, escaping any markdown syntax

#### Variables
* [Post](https://pkg.go.dev/git.garrido.io/gabriel/mastodon-markdown-archive/client#Post)

### Template examples

Here are some examples for basic templates that can be used. For an example on threading replies, see the [default template](files/templates/post.tmpl).

For any of these, save the template file somewhere and pass its path to the command as a value of the `--template` flag.

For the filename, pass it as a string to the command as a value of the `--filename` flag.

#### Jekyll

Template: 
```md
---
layout: post
title: {{ .Post.Id }}
---

{{ .Post.Content | toMarkdown }}
```

Filename:
`{{.Post.CreatedAt | date "2006-01-02"}}-{{.Post.Id}}.md`

#### Hugo and 11ty
The default template and filename is built for Hugo as that's the static site generator that I use, but a minimum viable template that works for either can look like this:

```md
---
title: {{ .Post.Id }}
date: {{ .Post.CreatedAt | date "2006-01-02" }}
---

{{ .Post.Content | toMarkdown }}
```

Filename:
`{{.Post.CreatedAt | date "2006-01-02"}}-{{.Post.Id}}.md`

#### HTML

Template:

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
Filename: `{{ .Post.Id }}.html`

#### Text only

Template:

```txt
{{ .Post.Content }}
```

Filename: `{{ .Post.Id }}.txt`

## Post media

By default, a post's media is not downloaded. Use the `--download-media` flag with a path to download a post's media. The post's original file is downloaded, and the image's id is used as the filename. 

For example, `--download-media=./images` saves any media to the `./images`.

Once downloaded, the media's path is available in [MediaAttachment.Path](https://pkg.go.dev/git.garrido.io/gabriel/mastodon-markdown-archive/client#MediaAttachment) as an absolute path.

Sprig's [path](https://masterminds.github.io/sprig/paths.html) functions can be used in the templates to manipulate the path as necessary. For example, the [default template](files/templates/post.tmpl#L25-L27) uses `osBase` to get the last element of the filepath.  

### Bundling

You can use `--download-media=bundle` to save the post media in a single directory with its original post. In this case, the post's filename will be used as the directory name and the post filename will be `index.{extension}`.

For example, `--download-media="./bundle" --filename='{{ .Post.CreatedAt | date "2006-01-02" }}-{{.Post.Id}}.md'` will create a `YYYY-MM-DD-<post id>/` directory, with the post saved as `YYYY-MM-DD-<post id>/index.md` and media saved as `YYYY-MM-DD-<post id>/<media id>.<media ext>`.

This is done specifically to support Hugo [page bundles](https://gohugo.io/content-management/page-bundles).
