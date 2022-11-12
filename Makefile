.PHONY : build run watch init-db clean bloghead.exe

build:
	go build

run:
	go run *.go

watch:
	find . -name '*.go' -or -name '*.tmpl' -or -name Makefile \
		| entr -rc -s "go build && ./bloghead -nobrowser Site1.bloghead"

watch-build:
	find . -name '*.go' -or -name '*.tmpl' -or -name Makefile \
		| entr -r go build

init-db:
	rm -f Site1.bloghead
	sqlite3 Site1.bloghead < models/schema.sql
	go run ./cmd/seed

clean:
	rm -rf www Site1.bloghead bloghead bloghead.exe

bloghead.exe:
	CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc GOOS=windows go build

blogfs/djot.lua: djot/* cmd/vendordjot/*
	go run ./cmd/vendordjot
