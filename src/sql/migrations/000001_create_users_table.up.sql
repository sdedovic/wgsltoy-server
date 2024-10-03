CREATE TYPE email_verification_type as ENUM ('pending', 'completed');

CREATE TABLE IF NOT EXISTS users (
  user_id             SERIAL                    PRIMARY KEY  ,
  created_at          timestamp with time zone  NOT NULL     ,
  updated_at          timestamp with time zone  NOT NULL     ,

  email               text                      NOT NULL     ,
  email_verification  email_verification_type   NOT NULL     ,
  username            text                      NOT NULL     ,
  password            text                      NOT NULL
);

ALTER TABLE users
    ADD CONSTRAINT unique_email UNIQUE (email),
    ADD CONSTRAINT unique_username UNIQUE (username);
