.PHONY : build linux windows run watch watch-build init-db clean watch-tk

build:
	go build -o dist/

linux:
	CGO_ENABLED=1 GOOS=linux go build -o dist/linux/bloghead
	cp vendored/djot.lua dist/linux/djot.lua

windows:
	CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc GOOS=windows \
		go build -o dist/windows/bloghead.exe -ldflags -H=windowsgui
	cp vendored/lua-5.4.2_Win64_bin/lua54.dll dist/windows/lua54.dll
	cp vendored/lua-5.4.2_Win64_bin/wlua54.exe dist/windows/wlua54.exe
	cp vendored/djot.lua dist/windows/djot.lua

run:
	go build -o dist/ && ./dist/bloghead

watch:
	find . -name '*.go' -or -name '*.tmpl' -or -name '*.tcl' \
		| entr -rc -s \
		"go build -o dist/ && ./dist/bloghead -nobrowser Site1.bloghead"

watch-build:
	find . -name '*.go' -or -name '*.tmpl' -or -name '*.tcl' \
		| entr -r go build -o dist/

watch-tk:
	find . -name '*.tcl' | entr -rc -s \
		"tclsh tk/scripts/choose-action.tcl"

init-db:
	rm -f Site1.bloghead
	sqlite3 Site1.bloghead < models/schema.sql
	go run ./cmd/seed

clean:
	rm -rf dist/* www *.bloghead bloghead bloghead.exe vendordjot seed

bloghead.syso: favicon.ico
	# needs `go install github.com/akavel/rsrc@latest`
	~/go/bin/rsrc -ico favicon.ico -o bloghead.syso
