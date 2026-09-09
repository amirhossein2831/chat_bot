-- +goose Up
create table conversations (
    id              uuid primary key,
    user_id         uuid not null references users(id) on delete cascade,
    title           text not null default '',
    title_norm      text not null default '',
    default_model   text not null,
    head_message_id uuid,
    archived_at     timestamptz,
    created_at      timestamptz not null default now()
);
create index conversations_user_recent_idx
    on conversations (user_id, created_at desc)
    where archived_at is null;
create index conversations_title_norm_trgm_idx
    on conversations using gin (title_norm gin_trgm_ops);

create table messages (
    id              uuid primary key,
    conversation_id uuid not null references conversations(id) on delete cascade,
    parent_id       uuid references messages(id) on delete cascade,
    role            text not null,
    content         text not null default '',
    content_norm    text not null default '',
    status          text not null,
    model           text not null default '',
    tokens_in       int not null default 0,
    tokens_out      int not null default 0,
    cost_usd_micros bigint not null default 0,
    error           text,
    created_at      timestamptz not null default now()
);
create index messages_branch_idx on messages (conversation_id, parent_id);
create index messages_content_norm_trgm_idx
    on messages using gin (content_norm gin_trgm_ops);

alter table conversations
    add constraint conversations_head_fkey
    foreign key (head_message_id) references messages(id) on delete set null;

-- +goose Down
alter table conversations drop constraint conversations_head_fkey;
drop table messages;
drop table conversations;
