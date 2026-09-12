-- +goose Up
-- The seed used plain quoted strings, where a backslash and an n stay two
-- characters rather than becoming a newline.
update prompts set body = replace(body, '\n', chr(10)) where body like '%\n%';

-- +goose Down
update prompts set body = replace(body, chr(10), '\n') where body like '%' || chr(10) || '%';
