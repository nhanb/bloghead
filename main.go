package main

import (
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"time"

	"go.imnhan.com/bloghead/models"
)

const Dbfile = "Site1.bloghead"
const Port = 8000
const Outdir = "www"

type PathDefs struct {
	Home     string
	Preview  string
	Settings string
	NewPost  string
	EditPost string
}

func (p *PathDefs) EditPostWithId(id int64) string {
	return fmt.Sprintf("%s/%d", p.EditPost, id)
}

var Paths = PathDefs{
	Home:     "/",
	Preview:  "/www/",
	Settings: "/settings",
	NewPost:  "/new",
	EditPost: "/edit/",
}

//go:embed templates
var tmplsFS embed.FS

type Templates struct {
	Home     *template.Template
	Settings *template.Template
	NewPost  *template.Template
	EditPost *template.Template
}

var tmpls Templates

func main() {
	models.Init(Dbfile)

	tmpls = Templates{
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
	}

	GenerateSite(Outdir)

	http.HandleFunc(Paths.Home, homeHandler)
	http.HandleFunc(Paths.Settings, settingsHandler)
	http.HandleFunc(Paths.NewPost, newPostHandler)
	http.HandleFunc(Paths.EditPost, editPostHandler)
	http.Handle(
		Paths.Preview,
		http.StripPrefix(Paths.Preview, http.FileServer(http.Dir(Outdir))),
	)

	fmt.Printf("Listening on port %d\n", Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", Port), nil))
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
			msg = fmt.Sprintf("Updated at %s", time.Now().Format("3:04:05 PM"))
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
		Title:      fmt.Sprintf("Editing post #%d", post.Id),
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
