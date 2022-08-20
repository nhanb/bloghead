package main

import (
	"fmt"
	"log"
	"net/http"

	"go.imnhan.com/bloghead/models"
)

const dbfile = "Site1.bloghead"
const port = 8000

func main() {
	models.InitDb(dbfile)

	http.HandleFunc("/", handler)

	fmt.Printf("Listening on port %d\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

func handler(w http.ResponseWriter, r *http.Request) {
	posts := models.QueryPosts()
	for _, p := range posts {
		fmt.Fprintf(w, "%v", p)
	}
}
