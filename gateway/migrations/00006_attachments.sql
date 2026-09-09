-- +goose Up
create table attachments (
    id                uuid primary key,
    user_id           uuid not null references users(id) on delete cascade,
    conversation_id   uuid references conversations(id) on delete set null,
    filename          text not null,
    mime              text not null,
    size_bytes        bigint not null,
    object_key        text not null,
    text_key          text,
    extraction_status text not null default 'pending',
    extraction_error  text,
    created_at        timestamptz not null default now()
);
create index attachments_user_recent_idx on attachments (user_id, created_at desc);

-- +goose Down
drop table attachments;
