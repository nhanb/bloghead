package main

import (
	"log"
	"os"

	"go.imnhan.com/bloghead/models"
)

func GenerateSite(outdir string) {
	posts := models.QueryPosts()
	for _, p := range posts {
		err := os.MkdirAll(outdir+"/"+p.Slug, 0750)
		if err != nil && !os.IsExist(err) {
			log.Fatalf("Failed to mkdir for post #%d (%s): %s", p.Id, p.Title, err)
		}

		err = os.WriteFile(outdir+"/"+p.Slug+"/index.html", []byte(p.Content), 0660)
		if err != nil {
			log.Fatal(err)
		}
	}
}
