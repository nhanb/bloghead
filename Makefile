.PHONY : build dist-linux dist-windows run watch watch-build init-db clean watch-tk

build:
	go build -o dist/

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
		"tclsh tk/choose-action.tcl"

init-db:
	rm -f Site1.bloghead
	sqlite3 Site1.bloghead < models/schema.sql
	go run ./cmd/seed

dist-linux:
	CGO_ENABLED=1 GOOS=linux go build -o dist/linux/bloghead
	cp vendored/djot.lua dist/linux/djot.lua

dist-windows:
	CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc GOOS=windows \
		go build -o dist/windows/bloghead.exe -ldflags -H=windowsgui
	cp vendored/lua-5.4.2_Win64_bin/lua54.dll dist/windows/
	cp vendored/lua-5.4.2_Win64_bin/wlua54.exe dist/windows/
	cp vendored/djot.lua dist/windows/
	cp vendored/tclkit/tclkit.exe dist/windows/

clean:
	rm -rf dist/* www *.bloghead bloghead bloghead.exe vendordjot seed

bloghead.syso: favicon.ico
	# needs `go install github.com/akavel/rsrc@latest`
	~/go/bin/rsrc -ico favicon.ico -o bloghead.syso
