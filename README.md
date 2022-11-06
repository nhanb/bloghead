# Bloghead

Linux build: [![builds.sr.ht status](https://builds.sr.ht/~nhanb/bloghead/commits/master/.build.yml.svg)](https://builds.sr.ht/~nhanb/bloghead/commits/master/.build.yml?)

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
- [entr](https://eradman.com/entrproject/)
- Docker (to build a mostly-static executable for [djot](https://github.com/jgm/djot))

```sh
# run these once:
git submodule update --init --recursive --remote  # pull djot submodule
make blogfs/djotbin
make init-db

make watch
```

Things are especially messy right now. Proper desktop-friendly distribution
will be done once core features are in place.
