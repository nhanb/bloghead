.PHONY : run watch init-db clean

run:
	go run *.go

watch:
	find . -name '*.go' | entr -rc go run *.go

init-db:
	rm -f Site1.bloghead
	sqlite3 Site1.bloghead < initdb.sql

clean:
	rm -rf www Site1.bloghead bloghead
