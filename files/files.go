package files

import (
	"fmt"
	"git.hq.ggpsv.com/gabriel/mastodon-pesos/client"
	md "github.com/JohannesKaufmann/html-to-markdown"
	"os"
	"path/filepath"
	"text/template"
)

type FileWriter struct {
	dir    string
	repies map[string]client.Post
}

type TemplateContext struct {
	Post        client.Post
	Descendants []client.Post
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

func (f FileWriter) Write(post client.Post) error {
	tpmlFilename := "templates/post.tmpl"
	tmplFile, err := filepath.Abs(tpmlFilename)

	if err != nil {
		return fmt.Errorf("error resolving template absolute path: %w", err)
	}

	if post.InReplyToId != "" {
		f.repies[post.InReplyToId] = post
		return nil
	}

	var descendants []client.Post
	f.getReplies(post.Id, &descendants)

	name := fmt.Sprintf("%s.md", post.Id)
	filename := filepath.Join(f.dir, name)
	file, err := os.Create(filename)

	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}

	defer file.Close()

	converter := md.NewConverter("", true, nil)

	funcs := template.FuncMap{
		"tomd": converter.ConvertString,
	}

	tmpl, err := template.New(filepath.Base(tpmlFilename)).Funcs(funcs).ParseFiles(tmplFile)

	context := TemplateContext{
		Post:        post,
		Descendants: descendants,
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
