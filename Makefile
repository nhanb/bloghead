.PHONY : build run watch init-db clean

build:
	go build

run:
	go run *.go

watch:
	find . -name '*.go' -or -name '*.tmpl' | entr -rc go run *.go -nobrowser

watch-build:
	find . -name '*.go' -or -name '*.tmpl' | entr -r go build

init-db:
	rm -f Site1.bloghead
	sqlite3 Site1.bloghead < schema.sql
	go run ./cmd/seed

clean:
	rm -rf www Site1.bloghead bloghead blogfs/djotbin

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
