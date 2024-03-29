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

	"fyne.io/systray"
	"go.imnhan.com/bloghead/blogfs"
	"go.imnhan.com/bloghead/models"
	"go.imnhan.com/bloghead/tk"
)

const ErrMsgCookie = "errMsg"
const MsgCookie = "msg"

const DraftHint = "A draft post won't be listed on your home page, but still accessible via direct link."

type PathDefs struct {
	Home              string
	Settings          string
	NewPost           string
	EditPost          string
	Attachments       string
	AttachmentsDelete string
	Preview           string
	Export            string
	Publish           string
	Neocities         string
	NeocitiesClear    string
	DjotToHtml        string

	InputFile string
}

var Paths = &PathDefs{
	Home:              "/",
	Settings:          "/settings",
	NewPost:           "/new",
	EditPost:          "/edit/",
	Attachments:       "/attachments/",
	AttachmentsDelete: "/attachments/delete",
	Preview:           "/preview/",
	Export:            "/export",
	Publish:           "/publish",
	Neocities:         "/publish/neocities",
	NeocitiesClear:    "/publish/neocities/clear",
	DjotToHtml:        "/djot-to-html",
}

func (p *PathDefs) EditPostWithId(id int64) string {
	return fmt.Sprintf("%s%d/", p.EditPost, id)
}
func (p *PathDefs) AttachmentsOfPost(id int64) string {
	return fmt.Sprintf("%s%d/", p.Attachments, id)
}
func (p *PathDefs) AttachmentPreview(postSlug string, filename string) string {
	return fmt.Sprintf("%s%s/%s", p.Preview, postSlug, filename)
}
func (p *PathDefs) GetPostIdFromAttachmentsPath(path string) (int64, error) {
	path = path[len(p.Attachments):]
	path = strings.TrimSuffix(path, "/")
	id, err := strconv.ParseInt(path, 10, 64)
	if id <= 0 {
		err = errors.New("Invalid (non-positive) post id")
	}
	return id, err
}
func (p *PathDefs) InputFileName() string {
	return filepath.Base(p.InputFile)
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
	Attachments    *template.Template
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
		"templates/new-post.tmpl",
	)),
	EditPost: template.Must(template.ParseFS(
		tmplsFS,
		"templates/base.tmpl",
		"templates/edit-post.tmpl",
	)),
	Attachments: template.Must(template.ParseFS(
		tmplsFS,
		"templates/base.tmpl",
		"templates/attachments.tmpl",
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
			Site      *models.Site
			Posts     []models.Post
			Paths     *PathDefs
			DraftHint string
		}{
			Site:      site,
			Posts:     posts,
			Paths:     Paths,
			DraftHint: DraftHint,
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
		post.Slug = r.FormValue("slug")
		post.IsDraft = true
		err := post.Create()
		if err == nil {
			http.Redirect(
				w, r,
				Paths.EditPostWithId(post.Id)+"?msg="+url.QueryEscape(
					fmt.Sprintf("Successfully created post #%d.", post.Id),
				),
				http.StatusSeeOther,
			)
			return
		}
		errMsg = err.Error()
	}

	err := tmpls.NewPost.Execute(w, struct {
		Paths   *PathDefs
		CsrfTag template.HTML
		ErrMsg  string
		Post    models.Post
		Title   string
	}{
		Paths:   Paths,
		CsrfTag: csrfTag,
		ErrMsg:  errMsg,
		Post:    post,
		Title:   "New post",
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

	validPath := regexp.MustCompile("^" + Paths.EditPost + `([0-9]+)/(.*)$`)
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

	if r.Method == "DELETE" {
		post.Delete()
		w.WriteHeader(200)
		return
	}

	// small hack so that embedded attachments work in live prevew
	if match[2] != "" {
		previewPath := Paths.Preview + post.Slug + "/" + match[2]
		http.Redirect(w, r, previewPath, http.StatusSeeOther)
		return
	}

	var msg, errMsg string

	if r.Method == "POST" {
		post.Title = r.FormValue("title")
		post.Content = r.FormValue("content")
		post.Slug = r.FormValue("slug")
		post.IsDraft = r.FormValue("is-draft") != ""
		err := post.Update()
		if err == nil {
			msg = "Successfully saved."
		} else {
			errMsg = err.Error()
		}

	} else if r.Method == "GET" {
		msg = r.URL.Query().Get("msg")
	}

	err = tmpls.EditPost.Execute(w, struct {
		Paths           *PathDefs
		CsrfTag         template.HTML
		Msg             string
		ErrMsg          string
		Post            models.Post
		Title           string
		ActionPath      string
		DraftHint       string
		Attachments     []models.Attachment
		PostContentHtml template.HTML
	}{
		Paths:           Paths,
		CsrfTag:         csrfTag,
		Msg:             msg,
		ErrMsg:          errMsg,
		Post:            *post,
		Title:           fmt.Sprintf("Editing post: %s", post.Title),
		ActionPath:      r.URL.Path,
		DraftHint:       DraftHint,
		Attachments:     models.QueryAttachments(postId),
		PostContentHtml: blogfs.DjotToHtml(post.Content),
	})
	if err != nil {
		log.Fatal(err)
	}
}

func write404(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("404 Not Fun :(\n"))
	w.Write([]byte(r.URL.Path))
}

func attachmentsHandler(w http.ResponseWriter, r *http.Request) {
	postId, err := Paths.GetPostIdFromAttachmentsPath(r.URL.Path)
	if err != nil {
		write404(w, r)
		return
	}

	post, err := models.QueryPost(postId)
	if err != nil {
		write404(w, r)
		return
	}

	csrfTag := CsrfCheck(w, r)
	if csrfTag == "" {
		return
	}

	var msg, errMsg string

	if r.Method == "POST" {
		var numFiles int
		errMsg = (func() string {
			err := r.ParseMultipartForm(50 * 1024 * 1024)
			if err != nil {
				return fmt.Sprintf("Multipart form parse error: %v", err)
			}

			files, ok := r.MultipartForm.File["attachments"]
			if !ok {
				return "No files selected."
			}

			for _, fileHeader := range files {
				file, err := fileHeader.Open()
				defer file.Close()
				if err != nil {
					log.Fatalf("open multipart file %s: %s", fileHeader.Filename, err)
				}
				fileData, err := io.ReadAll(file)
				if err != nil {
					log.Fatalf("copy multipart file %s: %s", fileHeader.Filename, err)
				}
				attm := &models.Attachment{
					Name:   fileHeader.Filename,
					Data:   fileData,
					PostId: postId,
				}
				err = attm.Create()
				if err != nil {
					return fmt.Sprintf("Error uploading %s: %s", attm.Name, err)
				}
				numFiles += 1
			}
			return ""
		})()

		if errMsg == "" {
			msg = fmt.Sprintf("Successfully uploaded %d files.", numFiles)
		}
	}

	attachments := models.QueryAttachments(postId)
	err = tmpls.Attachments.Execute(w, struct {
		Paths       *PathDefs
		CsrfTag     template.HTML
		Site        *models.Site
		Msg         string
		ErrMsg      string
		Post        *models.Post
		Attachments []models.Attachment
	}{
		Paths:       Paths,
		CsrfTag:     csrfTag,
		Site:        models.QuerySite(),
		Post:        post,
		Attachments: attachments,
		Msg:         msg,
		ErrMsg:      errMsg,
	})
	if err != nil {
		log.Fatal(err)
	}
}

func attachmentsDeleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("405 method not allowed."))
		return
	}

	csrfTag := CsrfCheck(w, r)
	if csrfTag == "" {
		return
	}

	postSlug := r.FormValue("post-slug")
	fileName := r.FormValue("file-name")

	if postSlug == "" || fileName == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("empty post-slug/file-name."))
		return
	}

	attm, err := models.QueryAttachment(postSlug, fileName)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("attachment not found."))
		return
	}

	attm.Delete()
	http.Redirect(w, r, Paths.AttachmentsOfPost(attm.PostId), http.StatusSeeOther)
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
		msg = "Successfully saved."
	}

	err := tmpls.Settings.Execute(w, struct {
		Site    models.Site
		Paths   *PathDefs
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
				msg = "Successfully exported."
			}
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
	}

	err := tmpls.Export.Execute(w, struct {
		Paths       *PathDefs
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
		Paths   *PathDefs
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
		Paths   *PathDefs
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
	var errMsg string
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
			return ""
		}()
		if errMsg == "" {
			http.Redirect(w, r, Paths.Publish, http.StatusSeeOther)
			return
		}
	}

	if errMsg != "" {
		neocities = models.QueryNeocities()
	}

	err := tmpls.Neocities.Execute(w, struct {
		Site    models.Site
		Paths   *PathDefs
		CsrfTag template.HTML
		ErrMsg  string
		Nc      models.Neocities
	}{
		Site:    *site,
		Paths:   Paths,
		CsrfTag: csrfTag,
		ErrMsg:  errMsg,
		Nc:      *neocities,
	})
	if err != nil {
		log.Fatal(err)
	}
}

func djotToHtmlHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("405 method not allowed."))
		return
	}

	input, err := io.ReadAll(r.Body)
	if err != nil {
		log.Fatalf("read djot input: %v", err)
	}

	output := blogfs.DjotToHtml(string(input))
	w.Write([]byte(output))
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
	http.HandleFunc(Paths.Attachments, attachmentsHandler)
	http.HandleFunc(Paths.AttachmentsDelete, attachmentsDeleteHandler)
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
	http.HandleFunc(Paths.DjotToHtml, djotToHtmlHandler)
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
