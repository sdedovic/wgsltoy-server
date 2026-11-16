CREATE TYPE email_verification_type AS ENUM ('pending', 'completed');

CREATE TABLE users (
    user_id             uuid                      PRIMARY KEY,
    created_at          timestamp with time zone  NOT NULL,
    updated_at          timestamp with time zone  NOT NULL,

    email               text                      NOT NULL,
    email_verification  email_verification_type   NOT NULL,
    username            text                      NOT NULL,
    password            text                      NOT NULL,

    CONSTRAINT unique_email UNIQUE (email),
    CONSTRAINT unique_username UNIQUE (username)
);

CREATE TYPE shader_visibility_type AS ENUM ('private', 'unlisted', 'public');

CREATE TABLE shaders (
    shader_id           uuid                      PRIMARY KEY,
    created_at          timestamp with time zone  NOT NULL,
    updated_at          timestamp with time zone  NOT NULL,

    created_by          uuid                      NOT NULL REFERENCES users (user_id),
    visibility          shader_visibility_type    NOT NULL,
    name                text                      NOT NULL,
    description         text                      NOT NULL,
    content             text                      NOT NULL,
    tags                text[]                    NOT NULL
);
