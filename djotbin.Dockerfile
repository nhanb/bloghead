# Using the oldest supported distro so that our compiled executable
# works on any Linux distro with a reasonably non-ancient glibc.
FROM debian:oldstable
RUN apt-get update && apt-get install -y curl gcc luarocks
RUN luarocks install luastatic
ADD ./djot /djot

# Create a djotbin executable that only depends on glibc:
WORKDIR /djot
RUN luastatic\
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
  /usr/lib/x86_64-linux-gnu/liblua5.1.a\
  -I /usr/include/lua5.1\
  -o djotbin
