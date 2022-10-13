pragma user_version = 1;
pragma foreign_keys = true;
pragma busy_timeout = 1000;

-- Site: global options & site-wide metadata
create table site (
    id integer primary key check (id = 0), -- ensures single row
    title text not null default 'My Site',
    tagline text not null default 'Let''s start this thing off right.'
);
insert into site(id) values(0);

-- Post
create table post (
    id integer primary key,
    path text unique not null,
    title text not null,
    content text not null,
    created_at text default (datetime('now')),
    updated_at text default null
);
insert into post(path, title, content) values
    ('1st', 'Hello World', 'I am your first post.'),
    ('the-second', 'Second coming?', 'I''m second.'),
    ('the/third', 'Third!!', 'Third time''s the charm.')
;
