package main

import (
	"context"
	"embed"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"fyne.io/systray"
	"go.imnhan.com/bloghead/blogfs"
	"go.imnhan.com/bloghead/models"
)

type PathDefs struct {
	Home       string
	Settings   string
	NewPost    string
	EditPost   string
	Preview    string
	Export     string
	ChangeSite string

	InputFile string
}

func (p *PathDefs) EditPostWithId(id int64) string {
	return fmt.Sprintf("%s/%d", p.EditPost, id)
}
func (p PathDefs) InputFileName() string {
	return filepath.Base(p.InputFile)
}

var Paths = PathDefs{
	Home:       "/",
	Settings:   "/settings",
	NewPost:    "/new",
	EditPost:   "/edit/",
	Preview:    "/preview/",
	Export:     "/export",
	ChangeSite: "/change",
}

//go:embed templates
var tmplsFS embed.FS

//go:embed favicon.ico
var favicon []byte

//go:embed favicon.png
var faviconpng []byte

type Templates struct {
	Home       *template.Template
	Settings   *template.Template
	NewPost    *template.Template
	EditPost   *template.Template
	Export     *template.Template
	ChangeSite *template.Template
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
	ChangeSite: template.Must(template.ParseFS(
		tmplsFS,
		"templates/base.tmpl",
		"templates/change-site.tmpl",
	)),
}

var bfs blogfs.BlogFS = blogfs.BlogFS{}

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

func changeSiteHandler(w http.ResponseWriter, r *http.Request) {
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

	err := tmpls.ChangeSite.Execute(w, struct {
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

type Flags struct {
	NoBrowser bool
	Port      int
	Args      []string
}

func processFlags() *Flags {
	var f Flags
	flag.BoolVar(&f.NoBrowser, "nobrowser", false, "Don't automatically open browser on startup")
	flag.IntVar(&f.Port, "port", 0, "Editor server port")
	flag.Parse()
	f.Args = flag.Args()

	switch len(f.Args) {
	case 0:
		Paths.InputFile = "Site1.bloghead"
	case 1:
		Paths.InputFile = f.Args[0]
	default:
		fmt.Println("Usage: bloghead [filename]")
		fmt.Println("  filename defaults to Site1.bloghead")
		os.Exit(1)
	}
	return &f
}

func startHttpServer(flags *Flags, portChan chan int) *http.Server {
	srv := &http.Server{}
	http.HandleFunc("/favicon.ico", faviconHandler)
	http.HandleFunc(Paths.Home, homeHandler)
	http.HandleFunc(Paths.Settings, settingsHandler)
	http.HandleFunc(Paths.NewPost, newPostHandler)
	http.HandleFunc(Paths.EditPost, editPostHandler)
	http.Handle(
		Paths.Preview,
		http.StripPrefix(
			Paths.Preview,
			http.FileServer(http.FS(&bfs)),
		),
	)
	http.HandleFunc(Paths.Export, exportHandler)
	http.HandleFunc(Paths.ChangeSite, changeSiteHandler)

	go func() {
		listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", flags.Port))
		if err != nil {
			log.Fatal(err)
		}

		port := listener.Addr().(*net.TCPAddr).Port
		portChan <- port
		close(portChan)

		// This must run after the socket starts listening, but before the
		// blocking HTTP server actually starts.
		if !flags.NoBrowser {
			openInBrowser(fmt.Sprintf("http://localhost:%d", port))
		}

		fmt.Printf("Serving %s on port %d\n", Paths.InputFile, port)
		if err := srv.Serve(listener); !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	return srv
}

func main() {
	flags := processFlags()
	// TODO: check if input file is a valid bloghead db

	models.Init()
	models.SetDbFile(Paths.InputFile)

	cleanUpDjotbin := blogfs.CreateDjotbin()
	defer cleanUpDjotbin()

	// We don't know what port we get until we actually start the server, but
	// the server starts asynchronously, so we need a channel here to wait
	// until we get back the actual port.
	portChan := make(chan int)
	srv := startHttpServer(flags, portChan)
	port := <-portChan

	// Let systray take over the main thread.
	// We shutdown the server when user clicks Exit in the systray menu.
	onReady := func() {
		systrayOnReady(fmt.Sprintf("http://localhost:%d", port))
	}
	onExit := func() {
		if err := srv.Shutdown(context.TODO()); err != nil {
			panic(err)
		}
	}
	systray.Run(onReady, onExit)
}
