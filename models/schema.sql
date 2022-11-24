pragma user_version = 1;
pragma foreign_keys = on;
pragma busy_timeout = 4000;

-- All time values must be in UTC, in this format: 'YYYY-MM-DD HH:MM:SSZ'.
-- I like it because it's human-friendly and leaves no room for interpretation.
--
-- Sqlite's date/time functions default to UTC, but datetime() doesn't include
-- the Z, so to get the current time in our desired format, use this instead:
-- strftime('%Y-%m-%d %H:%M:%SZ')

-- global options & site-wide metadata
create table site (
    id integer primary key check (id = 0), -- ensures single row
    title text not null default 'My Site',
    tagline text not null default 'Let''s start this thing off right.',
    export_to text not null default '',

    neocities_user text not null default '',
    neocities_password text not null default ''
) strict;
insert into site(id) values(0);

create table post (
    id integer primary key,
    slug text unique check (slug regexp '^[\w\-\.\~]+$') not null,
    title text unique not null,
    content text not null default '',
    created_at datetime not null default (strftime('%Y-%m-%d %H:%M:%SZ')),
    updated_at datetime default null,
    is_draft boolean not null default true
);

create table attachment (
    id integer primary key,
    name text not null check (name regexp '^[\w\-\.\~]+$'),
    data blob not null,

    post_id integer not null,
    foreign key (post_id) references post(id) on delete cascade,

    unique(post_id, name)
) strict;
