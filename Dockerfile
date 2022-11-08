FROM debian:oldstable

RUN apt-get update && apt-get install -y curl gcc luarocks mingw-w64

ADD . /bloghead
WORKDIR /bloghead
