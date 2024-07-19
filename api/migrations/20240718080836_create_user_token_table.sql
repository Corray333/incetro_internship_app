-- +goose Up
-- +goose StatementBegin
CREATE TABLE user_token (
    user_id BIGINT NOT NULL,
    token TEXT NOT NULL,
    CONSTRAINT user_token_pk PRIMARY KEY (user_id, token),
    CONSTRAINT user_token_user_id_fk FOREIGN KEY (user_id) REFERENCES users (user_id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE user_token;
-- +goose StatementEnd
