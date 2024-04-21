package files

import (
	"fmt"
	"git.hq.ggpsv.com/gabriel/mastodon-pesos/client"
	"os"
	"path/filepath"
	"text/template"
)

type FileWriter struct {
	dir    string
	repies map[string]client.Post
}

type TemplateContext struct {
	Post    client.Post
	Content string
	Replies []client.Post
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

func (f FileWriter) Write(post client.Post) {
	tpmlFilename := "templates/post.tmpl"
	tmplFile, err := filepath.Abs(tpmlFilename)

	if err != nil {
		panic(err)
	}

	if post.InReplyToId != "" {
		f.repies[post.InReplyToId] = post
		return
	}

	var descendants []client.Post
	f.getReplies(post.Id, &descendants)

	tmpl, err := template.New(filepath.Base(tpmlFilename)).ParseFiles(tmplFile)

	if err != nil {
		panic(err)
	}

	name := fmt.Sprintf("%s.md", post.Id)
	filename := filepath.Join(f.dir, name)
	file, err := os.Create(filename)

	if err != nil {
		panic(err)
	}

	defer file.Close()

	context := TemplateContext{
		Post:    post,
		Content: post.Content,
		Replies: descendants,
	}
	err = tmpl.Execute(file, context)

	if err != nil {
		panic(err)
	}
}

func (f FileWriter) getReplies(postId string, replies *[]client.Post) {
	if reply, ok := f.repies[postId]; ok {
		*replies = append(*replies, reply)
		f.getReplies(reply.Id, replies)
	}
}
