package main

import (
	"fmt"
	"log"
	"net/http"

	"go.imnhan.com/bloghead/models"
)

const dbfile = "Site1.bloghead"

func main() {
	models.InitDb(dbfile)

	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":8000", nil))
}

func handler(w http.ResponseWriter, r *http.Request) {
	posts := models.QueryPosts()
	for _, p := range posts {
		fmt.Fprintf(w, "%v", p)
	}
}
