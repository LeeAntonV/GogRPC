-- +goose Up
-- +goose StatementBegin
CREATE TABLE user_profile
(
    ID SERIAL PRIMARY KEY ,
    EMAIL VARCHAR(255) UNIQUE ,
    HASH VARCHAR(255) NOT NULL ,
    CODE VARCHAR(255) NOT NULL,
    VERIFIED BOOLEAN DEFAULT FALSE,
    CREATED_AT DATE DEFAULT NOW()
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS user_profile;

-- +goose StatementEnd
