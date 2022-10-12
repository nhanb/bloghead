package models

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func InitDb(dbfile string) {
	var err error
	db, err = sql.Open("sqlite3", dbfile)
	if err != nil {
		log.Fatal(err)
	}
}

type Site struct {
	Name        string
	Description string
}

type Post struct {
	Id    int64
	Path  string
	Title string
	Body  string
}

func QueryPosts() (posts []Post) {
	rows, err := db.Query("select id, path, title, body from post order by id desc;")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var p Post
		err = rows.Scan(&p.Id, &p.Path, &p.Title, &p.Body)
		if err != nil {
			log.Fatal(err)
		}
		posts = append(posts, p)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	return posts
}

func QuerySite() *Site {
	var s Site
	row := db.QueryRow("select name, description from site;")
	err := row.Scan(&s.Name, &s.Description)
	if err != nil {
		panic(err)
	}
	return &s
}
