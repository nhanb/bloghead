.PHONY : build run watch init-db clean

build:
	go build

run:
	go run *.go

watch:
	find . -name '*.go' -or -name '*.tmpl' | entr -rc go run *.go

init-db:
	rm -f Site1.bloghead
	sqlite3 Site1.bloghead < schema.sql
	go run ./cmd/seed

clean:
	rm -rf www Site1.bloghead bloghead
