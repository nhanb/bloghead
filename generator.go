package main

import (
	"log"
	"os"

	"go.imnhan.com/bloghead/models"
)

func GenerateSite(outdir string) {
	posts := models.QueryPosts()
	for _, p := range posts {
		err := os.MkdirAll(outdir+"/"+p.Path, 0750)
		if err != nil && !os.IsExist(err) {
			log.Fatalf("Failed to mkdir for post #%d (%s): %s", p.Id, p.Title, err)
		}

		err = os.WriteFile(outdir+"/"+p.Path+"/index.html", []byte(p.Body), 0660)
		if err != nil {
			log.Fatal(err)
		}
	}
}
