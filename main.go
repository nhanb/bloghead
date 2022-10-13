package main

import (
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"go.imnhan.com/bloghead/models"
)

const Dbfile = "Site1.bloghead"
const Port = 8000
const Outdir = "www"
const PreviewPath = "/www/"

//go:embed templates
var tmplsFS embed.FS

type Templates struct {
	NewPost *template.Template
	Home    *template.Template
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
		NewPost: template.Must(template.ParseFS(
			tmplsFS,
			"templates/base.tmpl",
			"templates/new-post.tmpl",
		)),
	}

	GenerateSite(Outdir)

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/new", newPostHandler)
	http.Handle(
		PreviewPath,
		http.StripPrefix(PreviewPath, http.FileServer(http.Dir(Outdir))),
	)

	fmt.Printf("Listening on port %d\n", Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", Port), nil))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
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
		}{
			Site:  site,
			Posts: posts,
		},
	)
	if err != nil {
		log.Fatal(err)
	}
}

func newPostHandler(w http.ResponseWriter, r *http.Request) {
	err := tmpls.NewPost.Execute(w, nil)
	if err != nil {
		log.Fatal(err)
	}
}
