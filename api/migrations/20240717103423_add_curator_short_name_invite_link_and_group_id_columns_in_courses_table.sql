-- +goose Up
-- +goose StatementBegin
ALTER TABLE courses ADD COLUMN curator_id BIGINT DEFAULT 0;
ALTER TABLE courses ADD COLUMN short_name TEXT DEFAULT '';
ALTER TABLE courses ADD COLUMN invite_link TEXT DEFAULT '';
ALTER TABLE courses ADD COLUMN group_id BIGINT DEFAULT 0;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE courses DROP COLUMN curator_id;
ALTER TABLE courses DROP COLUMN short_name;
ALTER TABLE courses DROP COLUMN invite_link;
ALTER TABLE courses DROP COLUMN group_id;
-- +goose StatementEnd
