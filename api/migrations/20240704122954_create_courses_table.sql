-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS courses (
    name TEXT UNIQUE,
    course_id UUID PRIMARY KEY -- notion id
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS courses;
-- +goose StatementEnd