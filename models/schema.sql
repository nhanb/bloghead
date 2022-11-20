pragma user_version = 1;
pragma foreign_keys = on;
pragma busy_timeout = 4000;

-- global options & site-wide metadata
create table site (
    id integer primary key check (id = 0), -- ensures single row
    title text not null default 'My Site',
    tagline text not null default 'Let''s start this thing off right.',
    export_to text not null default '',

    neocities_user text not null default '',
    neocities_password text not null default ''
);
insert into site(id) values(0);

create table post (
    id integer primary key,
    slug text unique check (slug regexp '^[\w\-\.\~]+$') not null,
    title text unique not null,
    content text not null,
    created_at text default (datetime('now', 'localtime')),
    updated_at text default null
);

create table file (
    id integer primary key,
    name text not null check (name regexp '^[\w\-\.\~]+$'),
    data blob not null,

    post_id integer not null,
    foreign key (post_id) references post(id) on delete cascade,

    unique(post_id, name)
);
