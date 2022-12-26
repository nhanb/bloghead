.PHONY : build linux windows run watch watch-build init-db clean watch-tk blogfs/djotbin blogfs/djotbin.exe

build:
	go build -o dist/

linux:
	CGO_ENABLED=1 GOOS=linux go build -o dist/bloghead

windows:
	CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc GOOS=windows \
		go build -o dist/bloghead.exe -ldflags -H=windowsgui

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

blogfs/djot.lua: djot/* cmd/vendordjot/*
	go run ./cmd/vendordjot

bloghead.syso: favicon.ico
	# needs `go install github.com/akavel/rsrc@latest`
	~/go/bin/rsrc -ico favicon.ico -o bloghead.syso

# Unfortunately Arch Linux doesn't provide static libs (.a),
# so for dev purposes I have to dynamically link to liblua:
# 	https://github.com/ers35/luastatic/issues/21
# The CI build is statically linked though.
blogfs/djotbin: blogfs/djot.lua
	cd blogfs; luastatic\
		djot.lua\
		-llua\
		-I/usr/include\
		-o djotbin
	rm blogfs/djot.luastatic.c

blogfs/djotbin.exe: blogfs/djot.lua
	cd blogfs; CC=x86_64-w64-mingw32-gcc luastatic\
		djot.lua\
		../vendored/lua-5.4.2_Win64_mingw6_lib/liblua54.a\
		-I ../vendored/lua-5.4.2_Win64_mingw6_lib/include\
		-o djotbin.exe -static -Wl,-subsystem,windows
	rm blogfs/djot.luastatic.c
