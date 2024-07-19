-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS update_data (
    update_id BIGINT NOT NULL PRIMARY KEY,
    data TEXT NOT NULL DEFAULT ''
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS update_data;
-- +goose StatementEnd
