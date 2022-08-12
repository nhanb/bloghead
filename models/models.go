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
	name        string
	description string
}

type Post struct {
	id    int64
	title string
	body  string
}

func QueryPosts() (posts []Post) {
	rows, err := db.Query("select id, title, body from post order by id desc;")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var p Post
		err = rows.Scan(&p.id, &p.title, &p.body)
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

func QuerySite() (s *Site) {
	row := db.QueryRow("select name, description from site;")
	err := row.Scan(&s.name, &s.description)
	if err != nil {
		panic(err)
	}
	return s
}
