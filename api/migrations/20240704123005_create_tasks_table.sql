-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE TABLE IF NOT EXISTS tasks (
    task_id UUID PRIMARY KEY, -- notion id
    title TEXT,
    content TEXT,
    course_id UUID REFERENCES courses(course_id),
    next UUID,
    type TEXT DEFAULT '' 
);
CREATE INDEX IF NOT EXISTS tasks_course_idx ON tasks(course_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS tasks_course_idx;
DROP TABLE IF EXISTS tasks;
DROP EXTENSION IF EXISTS "uuid-ossp";
-- +goose StatementEnd