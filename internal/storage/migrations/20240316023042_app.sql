-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS apps(
    ID SERIAL PRIMARY KEY,
    NAME TEXT NOT NULL UNIQUE,
    SECRET TEXT NOT NULL UNIQUE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS apps;
-- +goose StatementEnd
