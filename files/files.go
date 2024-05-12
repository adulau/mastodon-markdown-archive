package files

import (
	"embed"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"text/template"

	"git.garrido.io/gabriel/mastodon-markdown-archive/client"
	md "github.com/JohannesKaufmann/html-to-markdown"
)

//go:embed templates/post.tmpl
var templates embed.FS

type FileWriter struct {
	dir    string
	repies map[string]client.Post
}

type TemplateContext struct {
	Post        client.Post
	Descendants []client.Post
	Tags        []client.Tag
}

func New(dir string) (FileWriter, error) {
	var fileWriter FileWriter
	_, err := os.Stat(dir)

	if os.IsNotExist(err) {
		os.Mkdir(dir, os.ModePerm)
	}

	absDir, err := filepath.Abs(dir)

	if err != nil {
		return fileWriter, err
	}

	return FileWriter{
		dir:    absDir,
		repies: make(map[string]client.Post),
	}, nil
}

func (f FileWriter) Write(post client.Post, threaded bool, templateFile string) error {
	if threaded && post.InReplyToId != "" {
		f.repies[post.InReplyToId] = post
		return nil
	}

	var descendants []client.Post
	f.getReplies(post.Id, &descendants)

	var file *os.File
	var err error

	if len(post.MediaAttachments) == 0 {
		name := fmt.Sprintf("%s.md", post.Id)
		filename := filepath.Join(f.dir, name)
		file, err = os.Create(filename)
	} else {
		dir := filepath.Join(f.dir, post.Id)
		os.Mkdir(dir, os.ModePerm)

		for i := 0; i < len(post.MediaAttachments); i++ {
			media := &post.MediaAttachments[i]
			if media.Type != "image" {
				continue
			}

			imageFilename, err := downloadAttachment(dir, media.Id, media.URL)

			if err != nil {
				return err
			}

			media.Path = imageFilename
		}

		filename := filepath.Join(dir, "index.md")
		file, err = os.Create(filename)
	}

	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}

	defer file.Close()

	tmpl, err := resolveTemplate(templateFile)
	context := TemplateContext{
		Post:        post,
		Descendants: descendants,
		Tags:        client.TagsForPost(post, descendants),
	}

	err = tmpl.Execute(file, context)

	if err != nil {
		return fmt.Errorf("error executing template: %w", err)
	}

	return nil
}

func (f FileWriter) getReplies(postId string, replies *[]client.Post) {
	if reply, ok := f.repies[postId]; ok {
		*replies = append(*replies, reply)
		f.getReplies(reply.Id, replies)
	}
}

func downloadAttachment(dir string, id string, url string) (string, error) {
	var filename string

	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Accept", "image/*")
	res, err := client.Do(req)

	if err != nil {
		return filename, err
	}

	defer res.Body.Close()

	contentType := res.Header.Get("Content-Type")
	extensions, err := mime.ExtensionsByType(contentType)

	if err != nil {
		return filename, err
	}

	var extension string
	urlExtension := filepath.Ext(url)

	for _, i := range extensions {
		if i == urlExtension {
			extension = i
			break
		}
	}

	if extension == "" {
		return filename, fmt.Errorf("could not match extension for media")
	}

	filename = fmt.Sprintf("%s%s", id, extension)
	file, err := os.Create(filepath.Join(dir, filename))

	if err != nil {
		return filename, err
	}

	defer file.Close()
	_, err = io.Copy(file, res.Body)

	if err != nil {
		return filename, err
	}

	return filename, nil
}

func resolveTemplate(templateFile string) (*template.Template, error) {
	converter := md.NewConverter("", true, nil)

	funcs := template.FuncMap{
		"tomd": converter.ConvertString,
	}

	if templateFile == "" {
		tmpl, err := template.New("post.tmpl").Funcs(funcs).ParseFS(templates, "templates/*.tmpl")

		if err != nil {
			return tmpl, err
		}

		return tmpl, nil
	}

	tmpl, err := template.New(filepath.Base(templateFile)).Funcs(funcs).ParseGlob(templateFile)

	if err != nil {
		return tmpl, err
	}

	return tmpl, nil
}
