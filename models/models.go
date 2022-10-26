package models

import (
	"database/sql"
	"fmt"
	"log"
	"regexp"

	"github.com/mattn/go-sqlite3"
	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func Init(dbfile string) {
	// Register REGEXP function using Go
	// https://pkg.go.dev/github.com/mattn/go-sqlite3#hdr-Go_SQlite3_Extensions
	regex := func(re, s string) (bool, error) {
		return regexp.MatchString(re, s)
	}
	sql.Register("sqlite3_extended",
		&sqlite3.SQLiteDriver{
			ConnectHook: func(conn *sqlite3.SQLiteConn) error {
				return conn.RegisterFunc("regexp", regex, true)
			},
		},
	)

	var err error
	db, err = sql.Open("sqlite3_extended", dbfile)
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
	Slug    string
	Title   string
	Content string
}

func QueryPosts() (posts []Post) {
	rows, err := db.Query("select id, slug, title, content from post order by id desc;")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var p Post
		err = rows.Scan(&p.Id, &p.Slug, &p.Title, &p.Content)
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

func SaveSettings(title string, tagline string) {
	db.Exec("update site set title=?, tagline=?;", title, tagline)
}

func CreateNewPost(title string, slug string, content string) {
	result, err := db.Exec(
		"insert into post (title, slug, content) values (?, ?, ?);",
		title, slug, content,
	)
	fmt.Printf("%v\n%v", result, err)
}
