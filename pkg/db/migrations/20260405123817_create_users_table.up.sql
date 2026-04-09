CREATE TABLE users (
    id            UUID        PRIMARY KEY DEFAULT uuidv7(),
    username      TEXT        NOT NULL,
    password_hash BYTEA       NOT NULL,
    locale        TEXT        NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX users_username_idx ON users (username);
