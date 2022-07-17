package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

const dbfile = "Site1.bloghead"
const initsql = "initdb.sql"

func main() {
	db, err := sql.Open("sqlite3", dbfile)
	check(err)
	defer db.Close()

	sqlStmt, err := os.ReadFile(initsql)
	check(err)
	_, err = db.Exec(string(sqlStmt))
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return
	}

	rows, err := db.Query("select id, title, body from post order by id desc;")
	check(err)
	defer rows.Close()
	for rows.Next() {
		var id int
		var title string
		var body string
		err = rows.Scan(&id, &title, &body)
		check(err)
		fmt.Printf("(%d) %s: %s\n", id, title, body)
	}
	err = rows.Err()
	check(err)
}
