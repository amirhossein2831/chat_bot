-- +goose Up
create table user_settings (
    user_id             uuid primary key references users(id) on delete cascade,
    custom_instructions text not null default '',
    theme               text not null default 'system',
    updated_at          timestamptz not null default now()
);

create table memory_facts (
    id                uuid primary key,
    user_id           uuid not null references users(id) on delete cascade,
    text              text not null,
    text_norm         text not null,
    source_message_id uuid,
    pinned            boolean not null default false,
    created_at        timestamptz not null default now(),
    unique (user_id, text_norm)
);
create index memory_facts_user_idx on memory_facts (user_id, pinned desc, created_at desc);

-- +goose Down
drop table memory_facts;
drop table user_settings;
