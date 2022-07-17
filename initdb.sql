pragma user_version = 1;
pragma foreign_keys = true;
pragma busy_timeout = 1000;

-- Site: global options & site-wide metadata
drop table if exists site;
create table site (
    id integer primary key check (id = 0), -- ensures single row
    name text not null default 'My Site',
    description text not null default 'Let''s start this thing off right.'
);
insert into site(id) values(0);

-- Post
drop table if exists post;
create table post (
    id integer primary key,
    title text not null,
    body text not null
);
insert into post(title, body) values('Hello World', 'I am your first post.')
