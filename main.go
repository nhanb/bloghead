package main

import (
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
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
}

var Paths = PathDefs{
	Home:     "/",
	Preview:  "/www",
	Settings: "/settings",
	NewPost:  "/new",
}

//go:embed templates
var tmplsFS embed.FS

type Templates struct {
	Home     *template.Template
	Settings *template.Template
	NewPost  *template.Template
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
			"templates/new-post.tmpl",
		)),
	}

	GenerateSite(Outdir)

	http.HandleFunc(Paths.Home, homeHandler)
	http.HandleFunc(Paths.Settings, settingsHandler)
	http.HandleFunc(Paths.NewPost, newPostHandler)
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
		w.Write([]byte("404 Not Fun :("))
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
	err := tmpls.NewPost.Execute(w, struct{ Paths PathDefs }{Paths: Paths})
	if err != nil {
		log.Fatal(err)
	}
}

func settingsHandler(w http.ResponseWriter, r *http.Request) {
	csrfToken := CsrfCheck(w, r)
	if csrfToken == "" {
		return
	}

	var site *models.Site
	var msg string

	switch r.Method {
	case "GET":
		site = models.QuerySite()
	case "POST":
		title := r.FormValue("title")
		tagline := r.FormValue("tagline")
		models.SaveSettings(title, tagline)
		site = &models.Site{Title: title, Tagline: tagline}
		msg = fmt.Sprintf("Saved at %s", time.Now().Format("3:04:05 PM"))
	}

	err := tmpls.Settings.Execute(w, struct {
		Site      models.Site
		Paths     PathDefs
		CsrfToken string
		Msg       string
	}{
		Site:      *site,
		Paths:     Paths,
		CsrfToken: csrfToken,
		Msg:       msg,
	})
	if err != nil {
		log.Fatal(err)
	}
}
