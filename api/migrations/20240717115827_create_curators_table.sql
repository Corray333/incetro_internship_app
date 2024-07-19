-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS curators (
    curator_id BIGINT NOT NULL PRIMARY KEY,
    state INTEGER NOT NULL DEFAULT 0
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS curators;
-- +goose StatementEnd
