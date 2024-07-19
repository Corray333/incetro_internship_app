-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS system (
    id INTEGER NOT NULL GENERATED ALWAYS AS IDENTITY ( INCREMENT 1 START 1 CACHE 1 ),
    last_synced INTEGER 
);
INSERT INTO system (last_synced) VALUES (0);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS system;
-- +goose StatementEnd