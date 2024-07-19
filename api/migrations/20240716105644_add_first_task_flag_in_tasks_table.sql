-- +goose Up
-- +goose StatementBegin
ALTER TABLE tasks ADD is_first boolean NOT NULL DEFAULT false;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- +goose StatementEnd
