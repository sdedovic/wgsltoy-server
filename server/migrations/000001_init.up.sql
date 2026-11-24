CREATE TYPE email_verification_status AS ENUM ('pending', 'completed');
CREATE TYPE user_visibility           AS ENUM ('private', 'public');
CREATE TYPE shader_visibility         AS ENUM ('private', 'unlisted', 'public');

CREATE TABLE users (
    user_id             uuid                       PRIMARY KEY,
    created_at          timestamptz                NOT NULL,
    updated_at          timestamptz                NOT NULL,

    username            text                       NOT NULL,
    email               text                       NOT NULL,
    email_verification  email_verification_status  NOT NULL,
    password            text                       NOT NULL,
    visibility          user_visibility            NOT NULL  DEFAULT 'public',
    is_admin            boolean                    NOT NULL  DEFAULT false,

    CONSTRAINT unique_email    UNIQUE (email),
    CONSTRAINT unique_username UNIQUE (username)
);

CREATE TABLE shaders (
    shader_id   uuid              PRIMARY KEY,
    created_at  timestamptz       NOT NULL,
    updated_at  timestamptz       NOT NULL,

    created_by  uuid              REFERENCES users (user_id) ON DELETE SET NULL,
    visibility  shader_visibility NOT NULL,
    name        text              NOT NULL,
    description text              NOT NULL,
    content     text              NOT NULL,
    tags        text[]            NOT NULL
);

CREATE INDEX idx_shaders_created_by ON shaders (created_by);
CREATE INDEX idx_shaders_visibility ON shaders (visibility);
