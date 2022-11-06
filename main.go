package main

import (
	"embed"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"go.imnhan.com/bloghead/blogfs"
	"go.imnhan.com/bloghead/models"
)

const Dbfile = "Site1.bloghead"
const EditorPort = 8000

type PathDefs struct {
	Home     string
	Settings string
	NewPost  string
	EditPost string
	Preview  string
	Export   string
}

func (p *PathDefs) EditPostWithId(id int64) string {
	return fmt.Sprintf("%s/%d", p.EditPost, id)
}

var Paths = PathDefs{
	Home:     "/",
	Settings: "/settings",
	NewPost:  "/new",
	EditPost: "/edit/",
	Preview:  "/preview/",
	Export:   "/export",
}

//go:embed templates
var tmplsFS embed.FS

//go:embed favicon.ico
var favicon []byte

type Templates struct {
	Home     *template.Template
	Settings *template.Template
	NewPost  *template.Template
	EditPost *template.Template
	Export   *template.Template
}

var tmpls = Templates{
	Home: template.Must(template.ParseFS(
		tmplsFS,
		"templates/base.tmpl",
		"templates/home.tmpl",
	)),
	Settings: template.Must(template.ParseFS(
		tmplsFS,
		"templates/base.tmpl",
		"templates/settings.tmpl",
	)),
	NewPost: template.Must(template.ParseFS(
		tmplsFS,
		"templates/base.tmpl",
		"templates/edit-post.tmpl",
	)),
	EditPost: template.Must(template.ParseFS(
		tmplsFS,
		"templates/base.tmpl",
		"templates/edit-post.tmpl",
	)),
	Export: template.Must(template.ParseFS(
		tmplsFS,
		"templates/base.tmpl",
		"templates/export.tmpl",
	)),
}

var bfs blogfs.BlogFS = blogfs.BlogFS{}

func main() {
	models.Init(Dbfile)
	blogfs.CreateDjotbin()

	mux := http.NewServeMux()
	mux.HandleFunc("/favicon.ico", faviconHandler)
	mux.HandleFunc(Paths.Home, homeHandler)
	mux.HandleFunc(Paths.Settings, settingsHandler)
	mux.HandleFunc(Paths.NewPost, newPostHandler)
	mux.HandleFunc(Paths.EditPost, editPostHandler)
	mux.Handle(
		Paths.Preview,
		http.StripPrefix(
			Paths.Preview,
			http.FileServer(http.FS(&bfs)),
		),
	)
	mux.HandleFunc(Paths.Export, exportHandler)

	fmt.Printf("Editor server listening on port %d\n", EditorPort)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", EditorPort), mux))
}

func faviconHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write(favicon)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != Paths.Home {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 Not Fun :(\n"))
		w.Write([]byte(r.URL.Path))
		return
	}
	site := models.QuerySite()
	posts := models.QueryPosts()

	err := tmpls.Home.Execute(w,
		struct {
			Site  *models.Site
			Posts []models.Post
			Paths PathDefs
		}{
			Site:  site,
			Posts: posts,
			Paths: Paths,
		},
	)
	if err != nil {
		log.Fatal(err)
	}
}

func newPostHandler(w http.ResponseWriter, r *http.Request) {
	csrfTag := CsrfCheck(w, r)
	if csrfTag == "" {
		return
	}

	var errMsg string
	var post models.Post

	if r.Method == "POST" {
		post.Title = r.FormValue("title")
		post.Content = r.FormValue("content")
		post.Slug = r.FormValue("slug")
		err := post.Create()
		if err == nil {
			http.Redirect(
				w, r,
				Paths.EditPostWithId(post.Id)+"?msg="+url.QueryEscape(
					fmt.Sprintf("Successfully created post #%d", post.Id),
				),
				http.StatusSeeOther,
			)
			return
		}
		errMsg = err.Error()
	}

	err := tmpls.NewPost.Execute(w, struct {
		Paths      PathDefs
		CsrfTag    template.HTML
		Msg        string
		ErrMsg     string
		Post       models.Post
		Title      string
		SubmitText string
		ActionPath string
	}{
		Paths:      Paths,
		CsrfTag:    csrfTag,
		Msg:        "",
		ErrMsg:     errMsg,
		Post:       post,
		Title:      "New post",
		SubmitText: "Create",
		ActionPath: Paths.NewPost,
	})
	if err != nil {
		log.Fatal(err)
	}
}
func editPostHandler(w http.ResponseWriter, r *http.Request) {
	csrfTag := CsrfCheck(w, r)
	if csrfTag == "" {
		return
	}

	validPath := regexp.MustCompile("^" + Paths.EditPost + `([0-9]+)$`)
	match := validPath.FindStringSubmatch(r.URL.Path)
	if len(match) == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	postId, _ := strconv.ParseInt(match[1], 10, 64)

	post, err := models.QueryPost(postId)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(err.Error()))
		return
	}

	var msg, errMsg string

	if r.Method == "POST" {
		post.Title = r.FormValue("title")
		post.Content = r.FormValue("content")
		post.Slug = r.FormValue("slug")
		err := post.Update()
		if err == nil {
			msg = fmt.Sprintf("Updated at %s", post.UpdatedAt.Format("3:04:05 PM"))
		} else {
			errMsg = err.Error()
		}

	} else if r.Method == "GET" {
		msg = r.URL.Query().Get("msg")
	}

	err = tmpls.NewPost.Execute(w, struct {
		Paths      PathDefs
		CsrfTag    template.HTML
		Msg        string
		ErrMsg     string
		Post       models.Post
		Title      string
		SubmitText string
		ActionPath string
	}{
		Paths:      Paths,
		CsrfTag:    csrfTag,
		Msg:        msg,
		ErrMsg:     errMsg,
		Post:       *post,
		Title:      fmt.Sprintf("Editing post: %s", post.Title),
		SubmitText: "Update",
		ActionPath: r.URL.Path,
	})
	if err != nil {
		log.Fatal(err)
	}
}

func settingsHandler(w http.ResponseWriter, r *http.Request) {
	csrfTag := CsrfCheck(w, r)
	if csrfTag == "" {
		return
	}

	var site *models.Site
	var msg string

	switch r.Method {
	case "GET":
		site = models.QuerySite()
	case "POST":
		site = &models.Site{
			Title:   r.FormValue("title"),
			Tagline: r.FormValue("tagline"),
		}
		site.Update()
		msg = fmt.Sprintf("Saved at %s", time.Now().Format("3:04:05 PM"))
	}

	err := tmpls.Settings.Execute(w, struct {
		Site    models.Site
		Paths   PathDefs
		CsrfTag template.HTML
		Msg     string
	}{
		Site:    *site,
		Paths:   Paths,
		CsrfTag: csrfTag,
		Msg:     msg,
	})
	if err != nil {
		log.Fatal(err)
	}
}

func exportHandler(w http.ResponseWriter, r *http.Request) {
	csrfTag := CsrfCheck(w, r)
	if csrfTag == "" {
		return
	}

	var exportTo, msg, errMsg string

	switch r.Method {
	case "GET":
		exportTo = models.GetExportTo()
	case "POST":
		exportTo = r.FormValue("export-to")

		if exportTo == "" {
			errMsg = "Destination cannot be empty."
		} else {
			if !filepath.IsAbs(exportTo) {
				errMsg = fmt.Sprintf(
					`Destination must be an absolute path. "%s" isn't one.`,
					exportTo,
				)
			} else if _, err := os.Stat(exportTo); errors.Is(err, fs.ErrNotExist) {
				w.WriteHeader(http.StatusBadRequest)
				errMsg = fmt.Sprintf(
					`Folder "%s" does not exist. Create it first!`,
					exportTo,
				)
			}
		}

		if errMsg == "" {
			err := Export(&bfs, exportTo)
			if err != nil {
				errMsg = err.Error()
			} else {
				models.UpdateExportTo(exportTo)
				msg = fmt.Sprintf(
					"Exported successfully at %s",
					time.Now().Format("3:04:05 PM"),
				)
			}
		}
	}

	err := tmpls.Export.Execute(w, struct {
		Paths    PathDefs
		CsrfTag  template.HTML
		ExportTo string
		Msg      string
		ErrMsg   string
	}{
		Paths:    Paths,
		CsrfTag:  csrfTag,
		ExportTo: exportTo,
		Msg:      msg,
		ErrMsg:   errMsg,
	})
	if err != nil {
		log.Fatal(err)
	}
}

// Erases dest dir then copies everything from srcFS into dest.
// It assumes dest dir already exists.
func Export(srcFs fs.FS, dest string) error {
	dir, err := ioutil.ReadDir(dest)
	if err != nil {
		log.Fatal(err)
	}
	for _, d := range dir {
		if err := os.RemoveAll(path.Join(dest, d.Name())); err != nil {
			log.Fatal(err)
		}
	}

	err = fs.WalkDir(srcFs, ".", func(path string, d fs.DirEntry, err error) error {
		if path == "." {
			return nil
		}

		targetPath := filepath.Join(dest, path)

		if d.IsDir() {
			err := os.Mkdir(targetPath, os.FileMode(0755))
			if err != nil {
				return fmt.Errorf("create dest dir: %w", err)
			}
			return nil
		}

		destFile, err := os.Create(targetPath)
		if err != nil {
			return fmt.Errorf("create dest file: %w", err)
		}
		defer destFile.Close()

		srcFile, err := srcFs.Open(path)
		if err != nil {
			return fmt.Errorf("open src file: %w", err)
		}
		defer srcFile.Close()

		_, err = io.Copy(destFile, srcFile)
		if err != nil {
			return fmt.Errorf("cp src dest: %w", err)
		}
		return nil
	})

	return err
}
