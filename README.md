# Bloghead

Linux & Windows builds: [![builds.sr.ht status](https://builds.sr.ht/~nhanb/bloghead/commits/master/.build.yml.svg)](https://builds.sr.ht/~nhanb/bloghead/commits/master/.build.yml?)

Very early WIP.
The goal is to eventually become a user-friendly static site generator that:

- Uses an SQLite database per site instead of collections of flat files
- Has a traditional web-based CMS interface
- Simplifies deployment to popular targets (GH/GL/SRHT Pages, Tilde/SDF-likes via rsync, etc.)
- Acts reasonably like a desktop program, with proper .bloghead filetype association

The average computer-literate person deserves to completely own their blog
publishing software & data in $current_year!

This also doubles as my daily therapy session to recover from workday-induced
architectureastronomyphobia.

# Dev

Current dev dependencies:

- [go](https://go.dev/)
- [luastatic](https://github.com/ers35/luastatic): to compile djot - I installed it using luarocks.
- (optional) [entr](https://eradman.com/entrproject/): for `make watch`
- (optional) [mingw-w64](https://archlinux.org/groups/x86_64/mingw-w64/): to
  cross-compile from Linux to Windows.

I've been developing mainly on Arch Linux, but the CI builds work on Debian 9
("stretch") to ensure maximum glibc compatibiliy. Also cross-compiles to
Windows just fine. MacOS is TODO.

```sh
make init-db
make watch
```

Things are especially messy right now. Proper desktop-friendly distribution
will be done once core features are in place.

## Update vendored djot lua script

```sh
cd djot && git pull && cd ..
make blogfs/djot.lua
```

## Local build container

Although the Makefile works with plain Arch linux, there's a Dockerfile that's
supposed to match the CI's build environment. Use it to quickly debug and
iterate on CI tasks:

```sh
docker build -t bloghead .
docker run --rm -it bloghead bash
```

## Linux desktop integration

They're in the `./freedesktop/` dir. Try `stow freedesktop`.

# Prior art

[Lektor](https://www.getlektor.com/) is a heavy inspiration, but doesn't go far
enough to accommodate non-developers IMHO.
