package files

import (
	"bytes"
	"embed"
	"fmt"
	"github.com/Masterminds/sprig/v3"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"git.garrido.io/gabriel/mastodon-markdown-archive/client"
	md "github.com/JohannesKaufmann/html-to-markdown"
)

//go:embed templates/post.tmpl
var templates embed.FS

type FileWriter struct {
	dir              string
	templateFile     string
	filenameTemplate string
	downloadMedia    string
}

type TemplateContext struct {
	Post *client.Post
}

type FilenameDate struct {
	Year  int
	Month string
	Day   string
}

type FilenameTemplateContext struct {
	Post *client.Post
	Date FilenameDate
}

type PostFile struct {
	Dir  string
	Name string
	File *os.File
}

func New(dir, templateFile, filenameTemplate, downloadMedia string) (FileWriter, error) {
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
		dir:              absDir,
		templateFile:     templateFile,
		filenameTemplate: filenameTemplate,
		downloadMedia:    downloadMedia,
	}, nil
}

func (f *FileWriter) Write(post *client.Post) error {
	postFile, err := f.createFile(post)

	if err != nil {
		return err
	}
	defer postFile.File.Close()

	if f.downloadMedia != "" && len(post.AllMedia()) > 0 {
		var mediaDir string

		if f.downloadMedia == "bundle" {
			mediaDir = postFile.Dir
		} else {
			_, err := os.Stat(f.downloadMedia)
			if os.IsNotExist(err) {
				os.Mkdir(f.downloadMedia, os.ModePerm)
			}
			mediaDir = f.downloadMedia
		}

		if len(post.MediaAttachments) > 0 {
			err = downloadAttachments(post.MediaAttachments, mediaDir)
			if err != nil {
				return err
			}
		}

		for _, descendant := range post.Descendants() {
			if len(descendant.MediaAttachments) > 0 {
				err = downloadAttachments(descendant.MediaAttachments, mediaDir)
				if err != nil {
					return err
				}
			}
		}
	}

	tmpl, err := resolveTemplate(f.templateFile)
	context := TemplateContext{
		Post: post,
	}

	err = tmpl.Execute(postFile.File, context)

	if err != nil {
		return err
	}

	return nil
}

func (f *FileWriter) formatFilename(post *client.Post) (string, error) {
	tmplString := "{{.Post.Id}}"

	if f.filenameTemplate != "" {
		tmplString = f.filenameTemplate
	}

	tmpl := template.Must(template.New("filename").Funcs(sprig.FuncMap()).Parse(tmplString))

	filenameData := FilenameTemplateContext{
		Post: post,
	}

	var nameBuffer bytes.Buffer

	if err := tmpl.Execute(&nameBuffer, filenameData); err != nil {
		return "", err
	}

	return nameBuffer.String(), nil
}

func (f FileWriter) createFile(post *client.Post) (PostFile, error) {
	var postFile PostFile

	outputFilename, err := f.formatFilename(post)
	extension := filepath.Ext(outputFilename)
	shouldBundle := f.downloadMedia == "bundle" && len(post.AllMedia()) > 0

	if extension == "" {
		extension = ".md"
	} else {
		outputFilename = strings.TrimSuffix(outputFilename, extension)
	}

	if err != nil {
		return postFile, err
	}

	if shouldBundle {
		dir := filepath.Join(f.dir, outputFilename)

		_, err := os.Stat(dir)
		if os.IsNotExist(err) {
			os.Mkdir(dir, os.ModePerm)
		}

		name := filepath.Join(dir, fmt.Sprintf("index%s", extension))
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

	name := filepath.Join(f.dir, fmt.Sprintf("%s%s", outputFilename, extension))
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

		imageFile, err := downloadAttachment(dir, media.Id, media.URL)

		if err != nil {
			return err
		}

		absImageFile, err := filepath.Abs(imageFile.Name())

		if err != nil {
			return err
		}

		media.Path = absImageFile
	}

	return nil
}

func downloadAttachment(dir string, id string, url string) (*os.File, error) {
	var file *os.File

	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Accept", "image/*")
	res, err := client.Do(req)

	if err != nil {
		return file, err
	}

	defer res.Body.Close()

	contentType := res.Header.Get("Content-Type")
	extensions, err := mime.ExtensionsByType(contentType)

	if err != nil {
		return file, err
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
		return file, fmt.Errorf("could not match extension for media")
	}

	filename := fmt.Sprintf("%s%s", id, extension)
	file, err = os.Create(filepath.Join(dir, filename))

	if err != nil {
		return file, err
	}

	defer file.Close()
	_, err = io.Copy(file, res.Body)

	if err != nil {
		return file, err
	}

	return file, nil
}

func resolveTemplate(templateFile string) (*template.Template, error) {
	converter := md.NewConverter("", true, &md.Options{
		EscapeMode: "disabled",
	})
	converterEscaped := md.NewConverter("", true, &md.Options{
		EscapeMode: "basic",
	})

	funcs := sprig.FuncMap()
	funcs["toMarkdown"] = converter.ConvertString
	funcs["toMarkdownEscaped"] = converterEscaped.ConvertString

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
