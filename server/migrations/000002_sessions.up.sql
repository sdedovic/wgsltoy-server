CREATE TABLE sessions (
    session_id  text        PRIMARY KEY,
    user_id     uuid        NOT NULL REFERENCES users (user_id) ON DELETE CASCADE,
    created_at  timestamptz NOT NULL,
    expires_at  timestamptz NOT NULL
);

CREATE INDEX idx_sessions_user_id ON sessions (user_id);
