# Try to keep this in sync with whatever happens at .build.yml
FROM debian:oldstable

RUN apt-get update && apt-get install -y curl gcc mingw-w64

WORKDIR /root
RUN curl -L 'https://go.dev/dl/go1.19.3.linux-amd64.tar.gz' > go.tar.gz \
    && tar -xf go.tar.gz && rm go.tar.gz

ADD . /root/bloghead
WORKDIR /root/bloghead
