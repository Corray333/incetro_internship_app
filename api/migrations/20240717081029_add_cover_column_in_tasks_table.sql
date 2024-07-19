-- +goose Up
-- +goose StatementBegin
ALTER TABLE tasks ADD COLUMN cover TEXT DEFAULT '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE tasks DROP COLUMN cover;
-- +goose StatementEnd
