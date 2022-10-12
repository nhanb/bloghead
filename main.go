package main

import (
	"fmt"
	"log"
	"net/http"

	"go.imnhan.com/bloghead/models"
)

const dbfile = "Site1.bloghead"
const port = 8000
const outdir = "www"

func main() {
	models.InitDb(dbfile)

	GenerateSite(outdir)

	http.HandleFunc("/", handler)

	fmt.Printf("Listening on port %d\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

func handler(w http.ResponseWriter, r *http.Request) {
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
}
