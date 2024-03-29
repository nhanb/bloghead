# Bloghead

Linux & Windows builds: [![builds.sr.ht status](https://builds.sr.ht/~nhanb/bloghead/commits/master/.build.yml.svg)](https://builds.sr.ht/~nhanb/bloghead/commits/master/.build.yml?)

Quick demo:

https://user-images.githubusercontent.com/1446315/209965663-ace0b296-e0f6-4e5f-9529-7de0ebe59679.mp4

Very early WIP.
The goal is to eventually become a user-friendly static site generator that:

- Uses an SQLite database per site instead of collections of flat files
- Has a traditional web-based CMS interface
- Simplifies deployment to popular targets (GH/GL/SRHT Pages, Tilde/SDF-likes via rsync, etc.)
- Acts reasonably like a desktop program, with proper .bloghead filetype association

The average computer-literate person deserves to completely own their blog
publishing pipeline in $current_year!

This also doubles as my daily therapy session to recover from workday-induced
architectureastronomyphobia. That means rejecting abstraction until it hurts,
_then_ we can consider a minimum viable refactor. This is half useful software,
half contrarian art piece.

# Dev

Current dev dependencies:

- [go](https://go.dev/)
- [lua](https://www.lua.org/): to run the djot-to-html script
- [tcl/tk](https://archlinux.org/packages/extra/x86_64/tk/): for startup dialog/filepicker
- (optional) [entr](https://eradman.com/entrproject/): for `make watch`
- (optional) [mingw-w64](https://archlinux.org/groups/x86_64/mingw-w64/): to
  cross-compile from Linux to Windows.

I've been developing mainly on Arch Linux, but the CI builds work on Debian 9
("stretch") to ensure maximum glibc compatibiliy. Also cross-compiles to
Windows just fine. MacOS is TODO.

```sh
make init-db
ln -s "$PWD/vendored/djot.lua" /usr/bin/djot.lua  # or anywhere in your $PATH
make watch
```

Runtime dependencies:

- For Windows: everything is included in the zip. Just extract and run `bloghead.exe`.
- For Linux: `bloghead` assumes these executables are available from your $PATH:
  + `tclsh` - must include `tk` too.
  + `lua`
  + `djot.lua` - this is included in the zip, just put it somewhere in your $PATH.

Things are especially messy right now. Proper desktop-friendly distribution
will be done once core features are in place.

## Update vendored djot lua script

```sh
cd djot && git pull && cd ..
go run ./cmd/vendordjot
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

They're in the `./freedesktop/` dir.
Try editing the hardcoded paths there then `stow freedesktop`.

# Prior art

[Lektor](https://www.getlektor.com/) is a heavy inspiration, but doesn't go far
enough to accommodate non-developers IMHO.
