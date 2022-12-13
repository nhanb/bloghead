package blogfs

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log"
	"strings"
	"time"

	"go.imnhan.com/bloghead/models"
)

var _ fs.FS = (*BlogFS)(nil)
var _ fs.ReadDirFile = (*blogFile)(nil)
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
		return &blogFile{name: ".", isDir: true}, nil
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

	parts := strings.Split(name, "/")

	switch len(parts) {
	case 1:
		_, err := models.GetPostBySlug(name)
		if err != nil {
			return nil, fs.ErrNotExist
		}
		return &blogFile{name: name, isDir: true}, nil

	case 2:
		postSlug, fileName := parts[0], parts[1]
		attachment, err := models.QueryAttachment(postSlug, fileName)
		if err != nil {
			fmt.Printf("query attachment: %s", err)
			return nil, fs.ErrNotExist
		}
		return &blogFile{
			name:    attachment.Name,
			content: attachment.Data,
		}, nil
	}

	return nil, fs.ErrNotExist
}

func homeFile(posts []models.Post, site *models.Site) fs.File {
	buf := bytes.NewBufferString("")
	err := outTmpls.Home.Execute(buf,
		struct {
			HomePath string
			Site     *models.Site
			Posts    []models.Post
			Title    string
		}{
			HomePath: ".",
			Site:     site,
			Posts:    posts,
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
			HomePath    string
			Title       string
			Site        *models.Site
			Post        *models.Post
			HtmlContent template.HTML
		}{
			HomePath:    "../",
			Title:       p.Title,
			Site:        site,
			Post:        p,
			HtmlContent: djotToHtml(p.Content),
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
	name           string
	content        []byte
	offset         int64
	isDir          bool
	childrenOffset int64
}

func (f *blogFile) Stat() (fs.FileInfo, error) {
	return f, nil
}

// TODO: optimize pagination. Currently if caller makes use of the "n" param
// then we'll be querying the same thing on each iteration.
func (f *blogFile) ReadDir(n int) ([]fs.DirEntry, error) {
	fmt.Println("ReadDir:", f.name)
	if !f.isDir {
		return nil, errors.New(fmt.Sprintf("%s is a file, not a dir!", f.name))
	}

	var children []fs.DirEntry
	if f.name == "." {
		for _, slug := range models.QueryPostSlugs() {
			children = append(children, &blogFile{
				name:  slug,
				isDir: true,
			})
		}

	} else {
		for _, attachment := range models.QueryAttachmentsBySlug(f.name) {
			children = append(children, &blogFile{
				name:  attachment.Name,
				isDir: false,
			})
		}
	}

	children = append(children, &blogFile{name: "index.html", isDir: false})

	if n == -1 {
		return children, nil
	}

	returnVal := children[f.childrenOffset:n]
	f.childrenOffset += int64(n)
	if f.childrenOffset >= int64(len(children)) {
		f.childrenOffset = 0
	}
	return returnVal, nil
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
