pragma user_version = 1;
pragma foreign_keys = true;
pragma busy_timeout = 1000;

-- global options & site-wide metadata
create table site (
    id integer primary key check (id = 0), -- ensures single row
    title text not null default 'My Site',
    tagline text not null default 'Let''s start this thing off right.',
    export_to text not null default '',

    github_user text
        check (
            github_user = ''
            or (
                github_user regexp '^[A-Za-z0-9\-]{1,39}$'
                and not github_user like '%--%'
                and not github_user like '-%'
                and not github_user like '%-'
            )
        )
        not null default '',
    github_repo text check (github_repo = '' or github_repo regexp '^[A-Za-z0-9_\.\-]+$') not null default '',
    github_pub_key text not null default '',
    github_priv_key text not null default ''
);
insert into site(id) values(0);

create table post (
    id integer primary key,
    slug text unique check (slug regexp '^[a-zA-Z0-9-._]+$') not null,
    title text unique not null,
    content text not null,
    created_at text default (datetime('now', 'localtime')),
    updated_at text default null
);
