package models

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func Init(dbfile string) {
	var err error
	db, err = sql.Open("sqlite3", dbfile)
	if err != nil {
		log.Fatal(err)
	}
}

type Site struct {
	Title   string
	Tagline string
}

type Post struct {
	Id      int64
	Path    string
	Title   string
	Content string
}

func QueryPosts() (posts []Post) {
	rows, err := db.Query("select id, path, title, content from post order by id desc;")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var p Post
		err = rows.Scan(&p.Id, &p.Path, &p.Title, &p.Content)
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
	row := db.QueryRow("select title, tagline from site;")
	err := row.Scan(&s.Title, &s.Tagline)
	if err != nil {
		panic(err)
	}
	return &s
}
