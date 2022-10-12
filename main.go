package main

import (
	"fmt"
	"log"
	"net/http"

	"go.imnhan.com/bloghead/models"
)

const Dbfile = "Site1.bloghead"
const Port = 8000
const Outdir = "www"
const PreviewPath = "/www/"

func main() {
	models.InitDb(Dbfile)

	GenerateSite(Outdir)

	http.Handle(
		PreviewPath,
		http.StripPrefix(PreviewPath, http.FileServer(http.Dir(Outdir))),
	)
	http.HandleFunc("/", indexHandler)

	fmt.Printf("Listening on port %d\n", Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", Port), nil))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	site := models.QuerySite()
	fmt.Fprintf(w, "<h1>%s</h1>\n<p>%s</p>", site.Name, site.Description)

	posts := models.QueryPosts()
	fmt.Fprintln(w, "<ul>")
	for _, p := range posts {
		fmt.Fprintf(
			w,
			"<li><code>%s/</code> - <b>%s</b> - %s</li>",
			p.Path, p.Title, p.Body,
		)
	}
	fmt.Fprintln(w, "</ul>")

	fmt.Fprintln(w, "<a href='www'>Preview output</a>")
}
