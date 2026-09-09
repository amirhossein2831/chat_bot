-- +goose Up
create table auth_identities (
    id          uuid primary key,
    user_id     uuid not null references users(id) on delete cascade,
    kind        text not null,
    identifier  text not null,
    secret      text,
    verified_at timestamptz,
    created_at  timestamptz not null default now(),
    unique (kind, identifier)
);

-- +goose Down
drop table auth_identities;
