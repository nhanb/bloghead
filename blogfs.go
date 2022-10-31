package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log"
	"strings"
	"time"

	"go.imnhan.com/bloghead/djot"
	"go.imnhan.com/bloghead/models"
)

var _ fs.FS = (*BlogFS)(nil)
var _ fs.File = (*BlogFile)(nil)
var _ fs.FileInfo = (*BlogFile)(nil)
var _ fs.DirEntry = (*BlogFile)(nil)

type BlogFS struct{}

func (b *BlogFS) Open(name string) (fs.File, error) {
	site := models.QuerySite()
	println("Open:", name)

	if name == "." {
		return &BlogFile{isDir: true}, nil
	}

	if name == "index.html" {
		posts := models.QueryPosts()
		return homeFile(posts, site), nil
	}

	if strings.HasSuffix(name, "/index.html") {
		post, err := models.GetPostBySlug(name[:len(name)-len("/index.html")])
		if err == nil {
			return postFile(post, site), nil
		}
	}

	_, err := models.GetPostBySlug(name)
	if err == nil {
		return &BlogFile{isDir: true}, nil
	}

	return nil, fs.ErrNotExist
}

func homeFile(posts []models.Post, site *models.Site) fs.File {
	buf := bytes.NewBufferString("")
	err := outTmpls.Home.Execute(buf,
		struct {
			Site  *models.Site
			Posts []models.Post
			Title string
		}{
			Site:  site,
			Posts: posts,
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	return &BlogFile{
		name:    "index.html",
		content: buf.Bytes(),
	}
}

func postFile(p *models.Post, site *models.Site) fs.File {
	buf := bytes.NewBufferString("")

	err := outTmpls.Post.Execute(buf,
		struct {
			Site        *models.Site
			Title       string
			HtmlContent template.HTML
		}{
			Site:        site,
			Title:       p.Title,
			HtmlContent: djot.ToHtml(p.Content),
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	return &BlogFile{
		name:    "index.html",
		content: buf.Bytes(),
	}
}

type BlogFile struct {
	name    string
	content []byte
	offset  int64
	isDir   bool
}

func (f *BlogFile) Stat() (fs.FileInfo, error) {
	return f, nil
}

func (bf *BlogFile) Read(buf []byte) (int, error) {
	if bf.offset >= int64(len(bf.content)) {
		return 0, io.EOF
	}
	n := copy(buf, bf.content[bf.offset:])
	bf.offset += int64(n)
	return n, nil
}

func (bf *BlogFile) Seek(offset int64, whence int) (int64, error) {
	fmt.Printf("Seek: %v, %v\n", offset, whence)

	switch whence {
	case io.SeekStart:
		bf.offset = offset
	case io.SeekCurrent:
		bf.offset += offset
	case io.SeekEnd:
		bf.offset = int64(len(bf.content)) + offset
	default:
		log.Fatalf("Bogus whence: %d", whence)
	}
	return 0, nil
}

func (f *BlogFile) Close() error {
	return nil
}
func (f *BlogFile) Info() (fs.FileInfo, error) {
	return f, nil
}
func (f *BlogFile) Type() fs.FileMode {
	return f.Mode()
}
func (f *BlogFile) Name() string {
	return f.name
}
func (f *BlogFile) Size() int64 {
	return int64(len(f.content))
}
func (f *BlogFile) Mode() fs.FileMode {
	if f.isDir {
		return fs.FileMode(0555)
	}
	return fs.FileMode(0444)
}
func (f *BlogFile) ModTime() time.Time {
	return time.Now() // TODO
}
func (f *BlogFile) IsDir() bool {
	return f.isDir
}
func (f *BlogFile) Sys() any {
	return nil
}
