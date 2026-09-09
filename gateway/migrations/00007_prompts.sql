-- +goose Up
create table prompts (
    id         uuid primary key,
    locale     text not null,
    category   text not null,
    title      text not null,
    body       text not null,
    sort_order int  not null default 0,
    created_at timestamptz not null default now()
);
create index prompts_locale_category_idx on prompts (locale, category, sort_order);

-- +goose Down
drop table prompts;
