package blogfs

import (
	"bytes"
	"embed"
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
var _ fs.File = (*blogFile)(nil)
var _ fs.FileInfo = (*blogFile)(nil)
var _ fs.DirEntry = (*blogFile)(nil)

//go:embed blog-templates
var outTmplsFS embed.FS

type OutputTemplates struct {
	Home *template.Template
	Post *template.Template
}

var outTmpls = OutputTemplates{
	Home: template.Must(template.ParseFS(
		outTmplsFS,
		"blog-templates/base.tmpl",
		"blog-templates/home.tmpl",
	)),
	Post: template.Must(template.ParseFS(
		outTmplsFS,
		"blog-templates/base.tmpl",
		"blog-templates/post.tmpl",
	)),
}

type BlogFS struct{}

func (b *BlogFS) Open(name string) (fs.File, error) {
	site := models.QuerySite()
	println("Open:", name)

	if name == "." {
		return &blogFile{isDir: true}, nil
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
		return &blogFile{isDir: true}, nil
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
	return &blogFile{
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
	return &blogFile{
		name:    "index.html",
		content: buf.Bytes(),
	}
}

type blogFile struct {
	name    string
	content []byte
	offset  int64
	isDir   bool
}

func (f *blogFile) Stat() (fs.FileInfo, error) {
	return f, nil
}

func (bf *blogFile) Read(buf []byte) (int, error) {
	if bf.offset >= int64(len(bf.content)) {
		return 0, io.EOF
	}
	n := copy(buf, bf.content[bf.offset:])
	bf.offset += int64(n)
	return n, nil
}

func (bf *blogFile) Seek(offset int64, whence int) (int64, error) {
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

func (f *blogFile) Close() error {
	return nil
}
func (f *blogFile) Info() (fs.FileInfo, error) {
	return f, nil
}
func (f *blogFile) Type() fs.FileMode {
	return f.Mode()
}
func (f *blogFile) Name() string {
	return f.name
}
func (f *blogFile) Size() int64 {
	return int64(len(f.content))
}
func (f *blogFile) Mode() fs.FileMode {
	if f.isDir {
		return fs.FileMode(0555)
	}
	return fs.FileMode(0444)
}
func (f *blogFile) ModTime() time.Time {
	return time.Now() // TODO
}
func (f *blogFile) IsDir() bool {
	return f.isDir
}
func (f *blogFile) Sys() any {
	return nil
}
