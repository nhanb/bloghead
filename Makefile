.PHONY : build run watch init-db clean bloghead.exe

build:
	go build

run:
	go run *.go

watch:
	find . -name '*.go' -or -name '*.tmpl' -or -name Makefile | entr -rc -s\
		"go build && ./bloghead -nobrowser"

watch-build:
	find . -name '*.go' -or -name '*.tmpl' | entr -r go build

init-db:
	rm -f Site1.bloghead
	sqlite3 Site1.bloghead < models/schema.sql
	go run ./cmd/seed

clean:
	rm -rf www Site1.bloghead bloghead blogfs/djotbin\
		bloghead.exe blogfs/djotbin.exe

# Unfortunately Arch Linux doesn't provide static libs (.a),
# so for dev purposes I have to dynamically link to liblua:
# 	https://github.com/ers35/luastatic/issues/21
# The CI build is statically linked though.
blogfs/djotbin: djot/*
	cd djot; luastatic\
		bin/main.lua\
		djot.lua\
		djot/ast.lua\
		djot/attributes.lua\
		djot/block.lua\
		djot/emoji.lua\
		djot/html.lua\
		djot/inline.lua\
		djot/json.lua\
		djot/match.lua\
		-llua\
		-I/usr/include\
		-o ../blogfs/djotbin
	rm djot/main.luastatic.c

blogfs/djotbin.exe: djot/*
	cd djot; CC=x86_64-w64-mingw32-gcc luastatic\
		bin/main.lua\
		djot.lua\
		djot/ast.lua\
		djot/attributes.lua\
		djot/block.lua\
		djot/emoji.lua\
		djot/html.lua\
		djot/inline.lua\
		djot/json.lua\
		djot/match.lua\
		../vendored/lua-5.4.2_Win64_mingw6_lib/liblua54.a\
		-I ../vendored/lua-5.4.2_Win64_mingw6_lib/include\
		-o ../blogfs/djotbin
	rm djot/main.luastatic.c

bloghead.exe:
	CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc GOOS=windows go build
