-- +goose Up
create table subscriptions (
    id         uuid primary key,
    user_id    uuid not null references users(id) on delete cascade,
    plan_code  text not null,
    source     text not null default 'manual',
    expires_at timestamptz,
    created_at timestamptz not null default now()
);
create index subscriptions_user_idx on subscriptions (user_id, expires_at desc);

create table usage_events (
    id              uuid primary key,
    user_id         uuid not null references users(id) on delete cascade,
    message_id      uuid not null,
    model           text not null,
    tokens_in       int not null,
    tokens_out      int not null,
    cost_usd_micros bigint not null,
    created_at      timestamptz not null default now()
);
create index usage_events_user_recent_idx on usage_events (user_id, created_at desc);

-- +goose Down
drop table usage_events;
drop table subscriptions;
