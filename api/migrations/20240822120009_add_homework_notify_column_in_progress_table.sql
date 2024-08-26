-- +goose Up
-- +goose StatementBegin
ALTER TABLE progress ADD COLUMN homework_notify BIGINT NOT NULL DEFAULT 0;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE progress DROP COLUMN homework_notify;
-- +goose StatementEnd
