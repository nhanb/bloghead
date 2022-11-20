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
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"fyne.io/systray"
	"go.imnhan.com/bloghead/blogfs"
	"go.imnhan.com/bloghead/models"
	"go.imnhan.com/bloghead/tk"
)

const ErrMsgCookie = "errMsg"
const MsgCookie = "msg"

type PathDefs struct {
	Home           string
	Settings       string
	NewPost        string
	EditPost       string
	Preview        string
	Export         string
	Publish        string
	Neocities      string
	NeocitiesClear string

	InputFile string
}

func (p *PathDefs) EditPostWithId(id int64) string {
	return fmt.Sprintf("%s/%d", p.EditPost, id)
}
func (p PathDefs) InputFileName() string {
	return filepath.Base(p.InputFile)
}

var Paths = PathDefs{
	Home:           "/",
	Settings:       "/settings",
	NewPost:        "/new",
	EditPost:       "/edit/",
	Preview:        "/preview/",
	Export:         "/export",
	Publish:        "/publish",
	Neocities:      "/publish/neocities",
	NeocitiesClear: "/publish/neocities/clear",
}

//go:embed templates
var tmplsFS embed.FS

//go:embed favicon.ico
var favicon []byte

//go:embed favicon.png
var faviconpng []byte

type Templates struct {
	Home           *template.Template
	Settings       *template.Template
	NewPost        *template.Template
	EditPost       *template.Template
	Export         *template.Template
	Publish        *template.Template
	Neocities      *template.Template
	NeocitiesClear *template.Template
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
	Publish: template.Must(template.ParseFS(
		tmplsFS,
		"templates/base.tmpl",
		"templates/publish.tmpl",
	)),
	Neocities: template.Must(template.ParseFS(
		tmplsFS,
		"templates/base.tmpl",
		"templates/neocities.tmpl",
	)),
	NeocitiesClear: template.Must(template.ParseFS(
		tmplsFS,
		"templates/base.tmpl",
		"templates/neocities-clear.tmpl",
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

		errMsg = func() string {
			if exportTo == "" {
				return "Destination cannot be empty."
			}

			if !filepath.IsAbs(exportTo) {
				return fmt.Sprintf(
					`Destination must be an absolute path. "%s" isn't one.`,
					exportTo,
				)
			}

			stat, err := os.Stat(exportTo)
			if errors.Is(err, fs.ErrNotExist) {
				return fmt.Sprintf(
					`Folder "%s" does not exist. Create it first!`,
					exportTo,
				)
			}

			if !stat.IsDir() {
				return fmt.Sprintf(`"%s" is not a folder.`, exportTo)
			}

			return ""
		}()

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
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
	}

	err := tmpls.Export.Execute(w, struct {
		Paths       PathDefs
		CsrfTag     template.HTML
		ExportTo    string
		Msg         string
		ErrMsg      string
		Placeholder string
	}{
		Paths:       Paths,
		CsrfTag:     csrfTag,
		ExportTo:    exportTo,
		Msg:         msg,
		ErrMsg:      errMsg,
		Placeholder: exportPathExample,
	})
	if err != nil {
		log.Fatal(err)
	}
}

func publishHandler(w http.ResponseWriter, r *http.Request) {
	csrfTag := CsrfCheck(w, r)
	if csrfTag == "" {
		return
	}

	site := models.QuerySite()
	neocities := models.QueryNeocities()
	var msg, errMsg string

	if r.Method == "POST" {
		err := PublishNeocities(&bfs, neocities)
		if err != nil {
			errMsg = err.Error()
		} else {
			msg = "Successfully published."
		}
	}

	err := tmpls.Publish.Execute(w, struct {
		Site    models.Site
		Paths   PathDefs
		CsrfTag template.HTML
		Msg     string
		ErrMsg  string
		Nc      models.Neocities
	}{
		Site:    *site,
		Paths:   Paths,
		CsrfTag: csrfTag,
		Msg:     msg,
		ErrMsg:  errMsg,
		Nc:      *neocities,
	})
	if err != nil {
		log.Fatal(err)
	}
}

func neocitiesClearHandler(w http.ResponseWriter, r *http.Request) {
	csrfTag := CsrfCheck(w, r)
	if csrfTag == "" {
		return
	}

	neocities := models.QueryNeocities()
	if neocities.Username == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Can't clear what's not set."))
		return
	}

	site := models.QuerySite()
	var msg, errMsg string

	if r.Method == "POST" {
		err := models.ClearNeocities()
		if err == nil {
			http.Redirect(w, r, Paths.Publish, http.StatusSeeOther)
			return
		}
		errMsg = err.Error()
	}

	err := tmpls.NeocitiesClear.Execute(w, struct {
		Site    models.Site
		Paths   PathDefs
		CsrfTag template.HTML
		Msg     string
		ErrMsg  string
		Nc      models.Neocities
	}{
		Site:    *site,
		Paths:   Paths,
		CsrfTag: csrfTag,
		Msg:     msg,
		ErrMsg:  errMsg,
		Nc:      *neocities,
	})
	if err != nil {
		log.Fatal(err)
	}
}
func neocitiesHandler(w http.ResponseWriter, r *http.Request) {
	csrfTag := CsrfCheck(w, r)
	if csrfTag == "" {
		return
	}

	site := models.QuerySite()
	var msg, errMsg string
	var neocities *models.Neocities

	switch r.Method {
	case "GET":
		neocities = models.QueryNeocities()
	case "POST":
		neocities = &models.Neocities{
			Username: r.FormValue("username"),
			Password: r.FormValue("password"),
		}
		errMsg = func() string {
			err := CheckNeocitiesCreds(neocities)
			if err != nil {
				return err.Error()
			}
			err = neocities.Save()
			if err != nil {
				return err.Error()
			}
			msg = "Confirmed valid credentials."
			return ""
		}()
	}

	if errMsg != "" {
		neocities = models.QueryNeocities()
	}

	err := tmpls.Neocities.Execute(w, struct {
		Site    models.Site
		Paths   PathDefs
		CsrfTag template.HTML
		Msg     string
		ErrMsg  string
		Nc      models.Neocities
	}{
		Site:    *site,
		Paths:   Paths,
		CsrfTag: csrfTag,
		Msg:     msg,
		ErrMsg:  errMsg,
		Nc:      *neocities,
	})
	if err != nil {
		log.Fatal(err)
	}
}

// Erases dest dir then copies everything from srcFS into dest.
// It assumes dest dir already exists.
func Export(srcFs fs.FS, dest string) error {
	dir, err := os.ReadDir(dest)
	if err != nil {
		return err
	}
	for _, d := range dir {
		if err := os.RemoveAll(path.Join(dest, d.Name())); err != nil {
			log.Fatal(err)
		}
	}

	err = fs.WalkDir(srcFs, ".", func(path string, d fs.DirEntry, e error) error {
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

type Flags struct {
	NoBrowser bool
	Port      int
	Args      []string
}

func processFlags() *Flags {
	var f Flags
	flag.BoolVar(&f.NoBrowser, "nobrowser", false, "Don't automatically open browser on startup")
	flag.IntVar(&f.Port, "port", 4466, "HTTP server port")
	flag.Parse()
	f.Args = flag.Args()

	switch len(f.Args) {
	case 0: // let Paths.InputFile default to ""
	case 1:
		Paths.InputFile = f.Args[0]
	default:
		fmt.Println("Usage: bloghead [filename]")
		fmt.Println(`If filename is empty, a "Create Site" form will be shown.`)
		os.Exit(1)
	}
	return &f
}

func handleAllPaths(srv *http.Server) {
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
	http.HandleFunc(Paths.Publish, publishHandler)
	http.HandleFunc(Paths.Neocities, neocitiesHandler)
	http.HandleFunc(Paths.NeocitiesClear, neocitiesClearHandler)
}

func main() {
	flags := processFlags()
	// TODO: check if input file is a valid bloghead db

	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", flags.Port))
	if err != nil {
		// Most likely this means port is already in use because there's
		// already a running instance, so let's bail out early.
		log.Fatal(err)
	}

	// Needs to go before any sqlite is executed.
	// Putting this here before the potential "create db" flow.
	models.RegisterRegexFunc()

	// If bloghead was called without a filename argument, open a
	// tk window letting user choose between opening and creating a site.
	if Paths.InputFile == "" {
		tk.EnsureTclBin()

		action, filePath := tk.ChooseAction()
		println("Action:", action, filePath)

		switch action {
		case tk.ActionCancel:
			return
		case tk.ActionOpenFile:
			Paths.InputFile = filePath
		case tk.ActionCreateFile:
			if !strings.HasSuffix(filePath, ".bloghead") {
				filePath += ".bloghead"
			}
			_ = os.Remove(filePath)
			e := models.CreateDbFile(filePath)
			if e != nil {
				log.Fatalf("create blog file: %s", e)
			}
			Paths.InputFile = filePath
		}
	}

	// Use this to wait for both webserver and systray goroutines to finish
	var wg sync.WaitGroup

	// Start http server
	fmt.Printf("Serving %s on port %d\n", Paths.InputFile, flags.Port)
	srv := &http.Server{}
	handleAllPaths(srv)
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := srv.Serve(listener); !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	// Start systray too!
	wg.Add(1)
	go func() {
		defer wg.Done()
		onReady := func() {
			systrayOnReady(fmt.Sprintf("http://localhost:%d", flags.Port))
		}
		// Shutdown server when user clicks Exit in systray menu.
		onExit := func() {
			if err := srv.Shutdown(context.TODO()); err != nil {
				panic(err)
			}
		}
		systray.Run(onReady, onExit)
	}()

	models.SetDbFile(Paths.InputFile)
	defer models.Close()

	// This must run after the socket starts listening,
	// otherwise we risk opening an empty page.
	if !flags.NoBrowser {
		openInBrowser(fmt.Sprintf("http://localhost:%d", flags.Port))
	}

	wg.Wait()
}
