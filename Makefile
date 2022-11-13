.PHONY : build linux windows run watch watch-build init-db clean

build:
	go build -o dist/

linux:
	CGO_ENABLED=1 GOOS=linux go build -o dist/bloghead

windows:
	CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc GOOS=windows \
		go build -o dist/bloghead.exe -ldflags -H=windowsgui

run:
	go run *.go

watch:
	find . -name '*.go' -or -name '*.tmpl' -or -name Makefile \
		| entr -rc -s \
		"go build -o dist/ && ./dist/bloghead -nobrowser"

watch-build:
	find . -name '*.go' -or -name '*.tmpl' -or -name Makefile \
		| entr -r go build -o dist/

init-db:
	rm -f Site1.bloghead
	sqlite3 Site1.bloghead < models/schema.sql
	go run ./cmd/seed

clean:
	rm -rf dist/* www *.bloghead bloghead bloghead.exe vendordjot seed

blogfs/djot.lua: djot/* cmd/vendordjot/*
	go run ./cmd/vendordjot

bloghead.syso: favicon.ico
	# needs `go install github.com/akavel/rsrc@latest`
	~/go/bin/rsrc -ico favicon.ico -o bloghead.syso
