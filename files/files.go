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
	dir string
}

type TemplateContext struct {
	Post *client.Post
}

type PostFile struct {
	Dir  string
	Name string
	File *os.File
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
	}, nil
}

func (f FileWriter) Write(post *client.Post, templateFile string) error {
	hasMedia := len(post.AllMedia()) > 0
	postFile, err := f.createFile(post, hasMedia)

	if err != nil {
		return err
	}
	defer postFile.File.Close()

	if len(post.MediaAttachments) > 0 {
		err = downloadAttachments(post.MediaAttachments, postFile.Dir)
		if err != nil {
			return err
		}
	}

	for _, descendant := range post.Descendants() {
		if len(descendant.MediaAttachments) > 0 {
			err = downloadAttachments(descendant.MediaAttachments, postFile.Dir)
			if err != nil {
				return err
			}
		}
	}

	tmpl, err := resolveTemplate(templateFile)
	context := TemplateContext{
		Post: post,
	}

	err = tmpl.Execute(postFile.File, context)

	if err != nil {
		return err
	}

	return nil
}

func (f FileWriter) createFile(post *client.Post, shouldBundle bool) (PostFile, error) {
	var postFile PostFile

	if shouldBundle {
		dir := filepath.Join(f.dir, post.Id)

		_, err := os.Stat(dir)
		if os.IsNotExist(err) {
			os.Mkdir(dir, os.ModePerm)
		}

		name := filepath.Join(dir, "index.md")
		file, err := os.Create(name)

		if err != nil {
			return postFile, err
		}

		postFile = PostFile{
			Name: name,
			Dir:  dir,
			File: file,
		}

		return postFile, nil
	}

	name := filepath.Join(f.dir, fmt.Sprintf("%s.md", post.Id))
	file, err := os.Create(name)

	if err != nil {
		return postFile, err
	}

	postFile = PostFile{
		Name: name,
		Dir:  f.dir,
		File: file,
	}

	return postFile, nil
}

func downloadAttachments(attachments []client.MediaAttachment, dir string) error {
	for i := 0; i < len(attachments); i++ {
		media := &attachments[i]
		if media.Type != "image" {
			continue
		}

		imageFilename, err := downloadAttachment(dir, media.Id, media.URL)

		if err != nil {
			return err
		}

		media.Path = imageFilename
	}

	return nil
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
