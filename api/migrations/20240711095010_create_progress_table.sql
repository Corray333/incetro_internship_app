-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS progress (
    user_id BIGINT NOT NULL REFERENCES users(user_id),
    task_id UUID NOT NULL REFERENCES tasks(task_id),
    status INTEGER NOT NULL DEFAULT 1,
    completed_at BIGINT DEFAULT 0,
    homework TEXT DEFAULT '',
    PRIMARY KEY (user_id, task_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS progress;
-- +goose StatementEnd
