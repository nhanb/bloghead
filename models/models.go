package models

import (
	"database/sql"
	"errors"
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

func QueryPost(id int64) (*Post, error) {
	p := Post{Id: id}
	rows, err := db.Query("select slug, title, content from post where id=?;", id)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	found := rows.Next()
	if !found {
		return nil, errors.New(fmt.Sprintf("Post id=%d not found.", id))
	}
	rows.Scan(&p.Slug, &p.Title, &p.Content)
	return &p, nil
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

func (s *Site) Update() {
	db.Exec("update site set title=?, tagline=?;", s.Title, s.Tagline)
}

func (p *Post) Create() error {
	if p.Id != 0 {
		log.Fatalf("Calling Create() on existing Post: id=%d\n", p.Id)
	}
	result, err := db.Exec(
		"insert into post (title, slug, content) values (?,?,?);",
		p.Title, p.Slug, p.Content,
	)
	if err != nil {
		return processPostError(err)
	}

	p.Id, _ = result.LastInsertId()
	// mattn/go-sqlite3's LastInsertId() always returns a nil err:
	// https://github.com/mattn/go-sqlite3/blob/4ef63c9c0db77925ab91b95237f9e3802c4710a4/sqlite3.go#L2013-L2016
	return nil
}

func (p *Post) Update() error {
	if p.Id == 0 {
		log.Fatalln("Calling Update() on new Post (id=0).")
	}
	_, err := db.Exec(
		"update post set title=?, slug=?, content=? where id=?;",
		p.Title, p.Slug, p.Content, p.Id,
	)
	if err != nil {
		return processPostError(err)
	}
	return nil
}

var uniquenessErrMsg = regexp.MustCompile(
	`^UNIQUE constraint failed: [a-z]+\.([a-z]+)$`,
)
var regexpErrMsg = regexp.MustCompile(
	`^CHECK constraint failed: ([a-z]+) regexp .+$`,
)

// Turns sqlite3 errors into user-friendly error messages.
// Returns the error as-is if not recognized.
func processPostError(err error) error {
	errno := err.(sqlite3.Error).ExtendedCode
	errmsg := err.Error()

	switch errno {
	case sqlite3.ErrConstraintUnique:
		match := uniquenessErrMsg.FindStringSubmatch(errmsg)
		if len(match) > 0 {
			column := match[1]
			return errors.New(fmt.Sprintf(`%s already exists.`, column))
		}
	case sqlite3.ErrConstraintCheck:
		match := regexpErrMsg.FindStringSubmatch(errmsg)
		if len(match) > 0 {
			column := match[1]
			return errors.New(fmt.Sprintf(`%s has invalid format.`, column))
		}
	}

	return err
}
