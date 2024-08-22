-- +goose Up
-- +goose StatementBegin
ALTER TABLE tasks ADD COLUMN description TEXT NOT NULL DEFAULT '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE tasks DROP COLUMN description; 
-- +goose StatementEnd
