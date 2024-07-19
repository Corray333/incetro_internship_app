-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users (
    user_id BIGINT PRIMARY KEY,  -- tg chat id
    profile_id UUID,
    person_id UUID,
    verified BOOLEAN DEFAULT false,
    email TEXT DEFAULT '',
    course UUID REFERENCES courses(course_id),
    username TEXT DEFAULT '',
    phone TEXT DEFAULT '',
    fio TEXT DEFAULT '',
    avatar TEXT DEFAULT '',
    state INTEGER DEFAULT 0,
    fails INTEGER DEFAULT 0
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users;
-- +goose StatementEnd