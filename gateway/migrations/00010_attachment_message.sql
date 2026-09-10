-- +goose Up
-- A document belongs to the turn it was sent with, so regenerating or editing
-- that turn can put the same documents back in front of the model.
alter table attachments add column message_id uuid references messages(id) on delete set null;
create index attachments_message_idx on attachments (message_id) where message_id is not null;

-- +goose Down
drop index if exists attachments_message_idx;
alter table attachments drop column message_id;
