// Remember to call models.RegisterRegexFunc() before doing anything.
package models

import (
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"time"

	_ "embed"

	"github.com/mattn/go-sqlite3"
	_ "github.com/mattn/go-sqlite3"
)

//go:embed schema.sql
var initSQL string

var db *sql.DB

// https://pkg.go.dev/github.com/mattn/go-sqlite3#hdr-Go_SQlite3_Extensions
func RegisterRegexFunc() {
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
}

func SetDbFile(path string) {
	if db != nil {
		err := db.Close()
		if err != nil {
			log.Fatal(err)
		}
	}

	var err error
	db, err = sql.Open("sqlite3_extended", path)
	if err != nil {
		log.Fatal(err)
	}
}

type Site struct {
	Title   string
	Tagline string
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

func GetExportTo() (s string) {
	row := db.QueryRow("select export_to from site;")
	if err := row.Scan(&s); err != nil {
		panic(err)
	}
	return s
}

func UpdateExportTo(s string) {
	db.Exec("update site set export_to=?;", s)
}

type Post struct {
	Id        int64
	Slug      string
	Title     string
	Content   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func QueryPosts() (posts []Post) {
	rows, err := db.Query("select id, slug, title, content, created_at from post order by id desc;")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var p Post
		var createdAt string
		err = rows.Scan(&p.Id, &p.Slug, &p.Title, &p.Content, &createdAt)
		if err != nil {
			log.Fatal(err)
		}
		p.CreatedAt, err = time.Parse("2006-01-02 15:04:05", createdAt)
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

func QueryPostSlugs() (names []string) {
	rows, err := db.Query("select slug from post order by slug;")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var slugs []string
	for rows.Next() {
		var slug string
		err = rows.Scan(&slug)
		slugs = append(slugs, slug)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	return slugs
}

func QueryPost(id int64) (*Post, error) {
	p := Post{Id: id}
	rows, err := db.Query("select slug, title, content, created_at, updated_at from post where id=?;", id)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	found := rows.Next()
	if !found {
		return nil, errors.New(fmt.Sprintf("Post id=%d not found.", id))
	}
	var createdAt, updatedAt string

	rows.Scan(&p.Slug, &p.Title, &p.Content, &createdAt, &updatedAt)

	p.CreatedAt, err = time.Parse("2006-01-02 15:04:05", createdAt)
	if err != nil {
		log.Fatal(err)
	}

	if updatedAt != "" {
		p.UpdatedAt, err = time.Parse("2006-01-02 15:04:05", updatedAt)
		if err != nil {
			log.Fatal(err)
		}
	}
	return &p, nil
}

func GetPostBySlug(slug string) (*Post, error) {
	p := Post{Slug: slug}
	rows, err := db.Query(
		"select id, title, content, created_at from post where slug=?;",
		slug,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	found := rows.Next()
	if !found {
		return nil, errors.New(fmt.Sprintf(`Post slug="%s" not found.`, slug))
	}
	var createdAt string
	rows.Scan(&p.Id, &p.Title, &p.Content, &createdAt)
	p.CreatedAt, err = time.Parse("2006-01-02 15:04:05", createdAt)
	if err != nil {
		log.Fatal(err)
	}
	return &p, nil
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
	now := time.Now()
	_, err := db.Exec(
		"update post set title=?, slug=?, content=?, updated_at=? where id=?;",
		p.Title, p.Slug, p.Content, now.Format("2006-01-02 15:04:05"), p.Id,
	)
	if err != nil {
		return processPostError(err)
	}
	p.UpdatedAt = now
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

func CreateDbFile(fullPath string) error {
	dir := filepath.Dir(fullPath)
	stat, err := os.Stat(dir)
	if errors.Is(err, os.ErrNotExist) {
		err := os.MkdirAll(dir, fs.FileMode(0755))
		if err != nil {
			return err
		}
	}

	if !stat.IsDir() {
		return errors.New(fmt.Sprintf(`"%s" is not a folder!`, dir))
	}

	db, err = sql.Open("sqlite3_extended", fullPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec(initSQL)
	if err != nil {
		return fmt.Errorf("create initial db: %s", err)
	}

	return nil
}

func Close() error {
	return db.Close()
}

type Neocities struct {
	Username string
	Password string
	// TODO: offer symmetry encryption for password.
	// Ideally we could integrate with the OS's native keychain.
}

func QueryNeocities() *Neocities {
	nc := &Neocities{}
	row := db.QueryRow("select neocities_user, neocities_password from site;")
	err := row.Scan(&nc.Username, &nc.Password)
	if err != nil {
		panic(err)
	}
	return nc
}
func (nc *Neocities) Save() error {
	_, err := db.Exec(
		"update site set neocities_user=?, neocities_password=?;",
		nc.Username, nc.Password,
	)
	if err != nil {
		return err
	}
	return nil
}
