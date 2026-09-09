-- +goose Up
create extension if not exists citext;
create extension if not exists pg_trgm;

create table users (
    id            uuid primary key,
    email         citext not null unique,
    password_hash text   not null,
    locale        text   not null default 'fa',
    dir_pref      text   not null default 'auto',
    status        text   not null default 'active',
    created_at    timestamptz not null default now()
);

create table invites (
    code       text primary key,
    created_by uuid references users(id),
    used_by    uuid references users(id),
    used_at    timestamptz,
    created_at timestamptz not null default now()
);

-- +goose Down
drop table invites;
drop table users;
