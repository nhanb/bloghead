package main

import (
	"database/sql"
)

type Site struct {
	name        string
	description string
}

type Post struct {
	id    int64
	title string
	body  string
}

func QueryPosts(db *sql.DB) (posts []Post) {
	rows, err := db.Query("select id, title, body from post order by id desc;")
	check(err)
	defer rows.Close()
	for rows.Next() {
		var p Post
		err = rows.Scan(&p.id, &p.title, &p.body)
		check(err)
		posts = append(posts, p)
	}
	err = rows.Err()
	check(err)

	return posts
}

func QuerySite(db *sql.DB) (s *Site) {
	row := db.QueryRow("select name, description from site;")
	err := row.Scan(&s.name, &s.description)
	check(err)
	return s
}
