package models

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"

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

func CreateNewPost(p *Post) error {
	if p.Id != 0 {
		log.Fatalf("Calling CreateNewPost on existing Post: id=%d\n", p.Id)
	}

	result, err := db.Exec(
		"insert into post (title, slug, content) values (?,?,?);",
		p.Title, p.Slug, p.Content,
	)
	if err != nil {
		errno := err.(sqlite3.Error).ExtendedCode
		errmsg := err.Error()

		if errno == sqlite3.ErrConstraintUnique && strings.Contains(errmsg, "post.slug") {
			return errors.New(fmt.Sprintf(`Slug "%s" already exists.`, p.Slug))
		}
		if errno == sqlite3.ErrConstraintCheck && strings.Contains(errmsg, "slug") {
			return errors.New(fmt.Sprintf(`Slug "%s" has invalid format.`, p.Slug))
		}
		return err
	}

	p.Id, _ = result.LastInsertId()
	// mattn/go-sqlite3's LastInsertId() always returns a nil err:
	// https://github.com/mattn/go-sqlite3/blob/4ef63c9c0db77925ab91b95237f9e3802c4710a4/sqlite3.go#L2013-L2016
	return nil
}
