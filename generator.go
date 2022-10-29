package main

import (
	"embed"
	"fmt"
	"html/template"
	"log"
	"os"

	"go.imnhan.com/bloghead/djot"
	"go.imnhan.com/bloghead/models"
)

//go:embed output-templates
var outTmplsFS embed.FS

type OutputTemplates struct {
	Home *template.Template
	Post *template.Template
}

var outTmpls = OutputTemplates{
	Home: template.Must(template.ParseFS(
		outTmplsFS,
		"output-templates/base.tmpl",
		"output-templates/home.tmpl",
	)),
	Post: template.Must(template.ParseFS(
		outTmplsFS,
		"output-templates/base.tmpl",
		"output-templates/post.tmpl",
	)),
}

func GenerateSite(outdir string) {
	site := models.QuerySite()
	posts := models.QueryPosts()

	err := os.MkdirAll(outdir, 0750)
	if err != nil {
		log.Fatal(err)
	}

	generateHome(outdir, site, posts)

	for _, p := range posts {
		generatePost(outdir, site, &p)
	}
}

func generateHome(outdir string, site *models.Site, posts []models.Post) {
	w, err := os.Create(outdir + "/index.html")
	if err != nil {
		log.Fatal(err)
	}

	err = outTmpls.Home.Execute(w,
		struct {
			Site  *models.Site
			Posts []models.Post
			Title string
		}{
			Site:  site,
			Posts: posts,
		},
	)
	if err != nil {
		log.Fatal(err)
	}
}

func generatePost(outdir string, site *models.Site, p *models.Post) {
	err := os.MkdirAll(outdir+"/"+p.Slug, 0750)
	if err != nil && !os.IsExist(err) {
		log.Fatalf("Failed to mkdir for post #%d (%s): %s", p.Id, p.Title, err)
	}

	w, err := os.Create(fmt.Sprintf("%s/%s/index.html", outdir, p.Slug))
	if err != nil {
		log.Fatal(err)
	}

	err = outTmpls.Post.Execute(w,
		struct {
			Site        *models.Site
			Title       string
			HtmlContent template.HTML
		}{
			Site:        site,
			Title:       p.Title,
			HtmlContent: djot.ToHtml(p.Content),
		},
	)
	if err != nil {
		log.Fatal(err)
	}
}
