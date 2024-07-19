-- +goose Up
-- +goose StatementBegin
ALTER TABLE tasks ADD COLUMN section TEXT DEFAULT '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE tasks DROP COLUMN section;
-- +goose StatementEnd
